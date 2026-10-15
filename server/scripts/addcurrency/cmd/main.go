// Command addcurrency inserts a currency straight into the database,
// bypassing the HTTP API. It is a maintenance script rather than part of
// cmd/server: it is built and run on demand, e.g.
//
//	go run ./scripts/addcurrency/cmd -code USD -name "US Dollar" -database-url postgres://...
//
// Each value may instead come from an environment variable (see parseParams);
// the flag wins when both are set, and a value that neither supplies is an
// error. The insert goes through the addcurrency package, which builds the
// same gateway the server uses.
package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"jww/internal/domain"
	"jww/internal/infrastructure/database"
	"jww/scripts/addcurrency"
	"jww/utils/logger"
)

var log = logger.New().WithComponent("scripts.addcurrency")

func main() {
	params, err := parseParams(os.Args[1:], os.Getenv)
	if err != nil {
		log.Slog().Error("add currency failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	if err := run(params); err != nil {
		log.Slog().Error("add currency failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

// params is everything run needs: the database to talk to and the currency
// to create.
type params struct {
	databaseURL string
	code        string
	name        string
	isBase      bool
}

// parseParams resolves every value from a command-line flag first, then the
// matching environment variable, and fails when neither supplies a required
// one.
func parseParams(args []string, getenv func(string) string) (params, error) {
	fs := flag.NewFlagSet("addcurrency", flag.ContinueOnError)
	var p params

	defaultIsBase, _ := strconv.ParseBool(getenv("CURRENCY_IS_BASE"))

	fs.StringVar(&p.databaseURL, "database-url", getenv("DATABASE_URL"), "Postgres DSN (or DATABASE_URL)")
	fs.StringVar(&p.code, "code", getenv("CURRENCY_CODE"), "3-letter ISO 4217 currency code (or CURRENCY_CODE)")
	fs.StringVar(&p.name, "name", getenv("CURRENCY_NAME"), "currency name (or CURRENCY_NAME)")
	fs.BoolVar(&p.isBase, "base", defaultIsBase, "mark this currency as the base currency (or CURRENCY_IS_BASE)")
	if err := fs.Parse(args); err != nil {
		return p, err
	}

	var missing []string
	if p.databaseURL == "" {
		missing = append(missing, "-database-url or DATABASE_URL")
	}
	if p.code == "" {
		missing = append(missing, "-code or CURRENCY_CODE")
	}
	if p.name == "" {
		missing = append(missing, "-name or CURRENCY_NAME")
	}
	if len(missing) > 0 {
		return p, fmt.Errorf("missing required value(s): %s", strings.Join(missing, ", "))
	}

	p.code = strings.ToUpper(strings.TrimSpace(p.code))
	return p, nil
}

func run(p params) error {
	loggerConfig, err := logger.LoadConfig()
	if err != nil {
		return err
	}
	if err := logger.Configure(loggerConfig); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := database.NewPool(ctx, database.Config{URL: p.databaseURL})
	if err != nil {
		return err
	}
	defer pool.Close()

	var appCtx domain.AppContext
	appCtx.SetContext(ctx)
	appCtx.SetFlowID(uuid.NewString())

	created, err := addcurrency.NewClient(pool).Create(appCtx, p.code, p.name, p.isBase)
	if err != nil {
		return err
	}

	log.Slog().Info("currency created",
		slog.String("code", created.Code),
		slog.String("name", created.Name),
		slog.Bool("is_base", created.IsBase),
	)
	return nil
}
