// Command syncrates backfills currency exchange rates for a date range,
// bypassing the HTTP API and the scheduler. It is a maintenance script rather
// than part of cmd/server: it is built and run on demand, e.g.
//
//	go run ./scripts/syncrates/cmd -from 2026-01-01 -to 2026-03-31 -database-url postgres://...
//
// Each day goes through the same CurrencyExchangeRateWorkflow.Sync the nightly
// scheduler uses, so every directed pair among the currencies already in the
// database is fetched from Frankfurter and upserted. Re-running a range is
// safe: rows are updated in place, not duplicated.
//
// Each value may instead come from an environment variable (see parseParams);
// the flag wins when both are set, and a value that neither supplies is an
// error. RATES_INCLUDE_WEEKENDS is optional and defaults to false.
package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/google/uuid"

	"jww/internal/adapter/frankfurter"
	"jww/internal/domain"
	"jww/internal/infrastructure/database"
	"jww/scripts/syncrates"
	"jww/utils/logger"
)

var log = logger.New().WithComponent("scripts.syncrates")

func main() {
	params, err := parseParams(os.Args[1:], os.Getenv)
	if err != nil {
		log.Slog().Error("sync rates failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	if err := run(params); err != nil {
		log.Slog().Error("sync rates failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

// params is everything run needs: the database to talk to and which days to
// fill in.
type params struct {
	databaseURL     string
	from            time.Time
	to              time.Time
	includeWeekends bool
}

// parseParams resolves every value from a command-line flag first, then the
// matching environment variable, and fails when neither supplies a required
// one. Dates are bare calendar days (YYYY-MM-DD) interpreted as UTC.
func parseParams(args []string, getenv func(string) string) (params, error) {
	fs := flag.NewFlagSet("syncrates", flag.ContinueOnError)
	var p params

	defaultIncludeWeekends, _ := strconv.ParseBool(getenv("RATES_INCLUDE_WEEKENDS"))

	fs.StringVar(&p.databaseURL, "database-url", getenv("DATABASE_URL"), "Postgres DSN (or DATABASE_URL)")
	from := fs.String("from", getenv("RATES_FROM"), "first day, YYYY-MM-DD (or RATES_FROM)")
	to := fs.String("to", getenv("RATES_TO"), "last day, YYYY-MM-DD (or RATES_TO)")
	fs.BoolVar(&p.includeWeekends, "include-weekends", defaultIncludeWeekends, "also fetch Saturdays and Sundays (or RATES_INCLUDE_WEEKENDS)")
	if err := fs.Parse(args); err != nil {
		return p, err
	}

	var missing []string
	if p.databaseURL == "" {
		missing = append(missing, "-database-url or DATABASE_URL")
	}
	if *from == "" {
		missing = append(missing, "-from or RATES_FROM")
	}
	if *to == "" {
		missing = append(missing, "-to or RATES_TO")
	}
	if len(missing) > 0 {
		return p, fmt.Errorf("missing required value(s): %s", strings.Join(missing, ", "))
	}

	var err error
	if p.from, err = parseDate("-from", *from); err != nil {
		return p, err
	}
	if p.to, err = parseDate("-to", *to); err != nil {
		return p, err
	}
	return p, nil
}

// parseDate accepts only YYYY-MM-DD so the day never shifts with the machine's
// timezone.
func parseDate(flagName, raw string) (time.Time, error) {
	date, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(raw), time.UTC)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid %s date %q: want YYYY-MM-DD", flagName, raw)
	}
	return date, nil
}

func run(p params) error {
	loggerConfig, err := logger.LoadConfig()
	if err != nil {
		return err
	}
	if err := logger.Configure(loggerConfig); err != nil {
		return err
	}

	rateSourceConfig, err := frankfurter.LoadConfig()
	if err != nil {
		return err
	}

	// A range can run for a long time, so there is no fixed deadline: the
	// context is cancellation-only, and each provider request keeps the
	// adapter's own timeout.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := database.NewPool(ctx, database.Config{URL: p.databaseURL})
	if err != nil {
		return err
	}
	defer pool.Close()

	var appCtx domain.AppContext
	appCtx.SetContext(ctx)
	appCtx.SetFlowID(uuid.NewString())

	client := syncrates.NewClient(pool, frankfurter.NewClient(rateSourceConfig))
	result, err := client.SyncRange(appCtx, p.from, p.to, p.includeWeekends)

	// Log every day the workflow actually ran, then the aggregate — matching
	// the scheduler's WARN-per-warning, INFO-per-summary contract.
	for _, day := range result.Days {
		if day.Weekend {
			continue
		}
		date := day.Date.Format("2006-01-02")
		if day.Err != nil {
			log.Slog().Error("exchange rate sync failed",
				slog.String("date", date),
				slog.String("error", day.Err.Error()))
			continue
		}
		for _, warning := range day.Result.Warnings {
			log.Slog().Warn("exchange rate sync warning",
				slog.String("date", date),
				slog.String("detail", warning))
		}
		log.Slog().Info("exchange rates synced",
			slog.String("date", date),
			slog.Int("currencies", day.Result.Currencies),
			slog.Int("fetched", day.Result.Fetched),
			slog.Int("inserted", day.Result.Inserted),
			slog.Int("updated", day.Result.Updated),
			slog.Int("skipped", day.Result.Skipped))
	}

	if result.Synced == 0 && result.WeekendDays > 0 {
		log.Slog().Warn("range contains no business days",
			slog.String("from", p.from.Format("2006-01-02")),
			slog.String("to", p.to.Format("2006-01-02")))
	}

	log.Slog().Info("range sync complete",
		slog.String("from", p.from.Format("2006-01-02")),
		slog.String("to", p.to.Format("2006-01-02")),
		slog.Int("synced_days", result.Synced),
		slog.Int("weekend_days", result.WeekendDays),
		slog.Int("failed_days", result.Failed),
		slog.Int("inserted", result.Inserted),
		slog.Int("updated", result.Updated),
		slog.Int("skipped_pairs", result.SkippedPairs))

	return err
}
