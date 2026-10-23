// Package syncrates backfills currency exchange rates for a date range by
// running the same CurrencyExchangeRateWorkflow.Sync the daily scheduler runs,
// once per day, through the server's own gateway repositories, currency
// service, and Frankfurter rate source.
//
// It holds the reusable half of the sync-rates script: the command that calls
// it lives in cmd/. Nothing here reaches into Postgres or the provider
// directly, so the rows it writes are indistinguishable from scheduled ones.
package syncrates

import (
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"jww/internal/adapter/gateway"
	"jww/internal/domain"
	"jww/internal/infrastructure/database"
	"jww/internal/service/currency"
	currency_exchange_rate "jww/internal/service/currency_exchange_rate"
	"jww/internal/service/workflow"
)

// dateLayout is the only accepted date format: a bare calendar day, so there
// is no timezone to misread.
const dateLayout = "2006-01-02"

// syncer is the one workflow operation this tool needs. Keeping the dependency
// this narrow lets the range logic be tested with a fake, while NewClient wires
// the real CurrencyExchangeRateWorkflow for the command.
type syncer interface {
	Sync(ctx domain.AppContext, asOf time.Time) (workflow.SyncResult, error)
}

// Client runs the exchange-rate sync over an explicit date range.
type Client struct {
	syncer syncer
}

// NewClient wires the exchange-rate workflow over pool using the same adapters
// and services as cmd/server. The rate source is injected so the command can
// pass the configured Frankfurter client (and a test could pass a fake).
func NewClient(pool *pgxpool.Pool, rateSource currency_exchange_rate.RateSource) *Client {
	uowManager := database.NewPostgresUnitOfWorkManager(pool)
	currencyService := currency.NewService(gateway.NewPostgresCurrencyRepository(pool), uowManager)
	exchangeRateService := currency_exchange_rate.NewService(gateway.NewPostgresCurrencyExchangeRateRepository(pool))
	return newClient(workflow.NewCurrencyExchangeRateWorkflow(currencyService, exchangeRateService, rateSource, uowManager))
}

// newClient is the seam the unit tests use to supply a fake syncer.
func newClient(syncer syncer) *Client {
	return &Client{syncer: syncer}
}

// DayResult is one calendar day's outcome. Weekend marks a day that was
// skipped without calling the provider; Err is set when that day's sync
// failed (the range continues past it).
type DayResult struct {
	Date    time.Time
	Weekend bool
	Result  workflow.SyncResult
	Err     error
}

// RangeResult aggregates every day in the requested range. Inserted and
// Updated are summed across successful days; SkippedPairs counts provider
// omissions (not skipped days).
type RangeResult struct {
	Days         []DayResult
	Synced       int
	Failed       int
	WeekendDays  int
	Inserted     int
	Updated      int
	SkippedPairs int
}

// SyncRange records rates for every business day from from through to,
// inclusive. Both bounds are compared as UTC calendar days, so any time of day
// in either value is ignored. Weekends are skipped unless includeWeekends is
// set: the provider publishes business days only, and a weekend request just
// echoes the previous business day. A day that fails does not abort the range
// — the error is recorded on that DayResult and the loop continues — but the
// returned error still reports how many days failed, so the command exits
// non-zero.
func (c *Client) SyncRange(ctx domain.AppContext, from, to time.Time, includeWeekends bool) (RangeResult, error) {
	from, to = dateOnly(from), dateOnly(to)
	if to.Before(from) {
		return RangeResult{}, fmt.Errorf("end date %s is before start date %s", to.Format(dateLayout), from.Format(dateLayout))
	}

	today := dateOnly(time.Now().UTC())
	if to.After(today) {
		return RangeResult{}, fmt.Errorf("end date %s is in the future (today is %s)", to.Format(dateLayout), today.Format(dateLayout))
	}

	var result RangeResult
	for day := from; !day.After(to); day = day.AddDate(0, 0, 1) {
		// A long backfill must stop promptly on Ctrl-C/SIGTERM instead of
		// walking every remaining day with an already-cancelled context.
		if err := ctx.Context().Err(); err != nil {
			return result, fmt.Errorf("range sync interrupted after %d day(s): %w", result.Synced+result.Failed, err)
		}

		if !includeWeekends && isWeekend(day) {
			result.WeekendDays++
			result.Days = append(result.Days, DayResult{Date: day, Weekend: true})
			continue
		}

		syncResult, err := c.syncer.Sync(ctx, day)
		result.Days = append(result.Days, DayResult{Date: day, Result: syncResult, Err: err})
		if err != nil {
			result.Failed++
			continue
		}

		result.Synced++
		result.Inserted += syncResult.Inserted
		result.Updated += syncResult.Updated
		result.SkippedPairs += syncResult.Skipped
	}

	if result.Failed > 0 {
		return result, fmt.Errorf("%d of %d day(s) failed", result.Failed, result.Synced+result.Failed)
	}
	return result, nil
}

// isWeekend reports whether day falls on Saturday or Sunday.
func isWeekend(day time.Time) bool {
	weekday := day.Weekday()
	return weekday == time.Saturday || weekday == time.Sunday
}

// dateOnly strips the time of day, keeping the value in UTC.
func dateOnly(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}
