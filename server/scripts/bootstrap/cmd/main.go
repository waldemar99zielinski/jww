// Command bootstrap seeds the local development test accounts and
// currencies.
//
// It runs the bootstrap package, which reuses scripts/signup's Client to insert
// the fixed set of throwaway logins (julcia, wozi, baby — each password equal to
// its username) and scripts/addcurrency's Client to insert the fixed set of
// currencies (PLN, USD, EUR — PLN is the base currency) straight into the
// database. Re-running is safe: existing rows are left untouched.
//
//	go run ./scripts/bootstrap/cmd -database-url postgres://...
//
// The DSN may instead come from DATABASE_URL; the flag wins when both are set,
// and supplying neither is an error.
package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	"jww/internal/infrastructure/database"
	"jww/scripts/bootstrap"
	"jww/utils/logger"
)

var log = logger.New().WithComponent("scripts.bootstrap")

func main() {
	databaseURL, err := parseDatabaseURL(os.Args[1:], os.Getenv)
	if err != nil {
		log.Slog().Error("bootstrap failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	if err := run(databaseURL); err != nil {
		log.Slog().Error("bootstrap failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

// parseDatabaseURL resolves the DSN from the -database-url flag first, then the
// DATABASE_URL environment variable, failing when neither is set.
func parseDatabaseURL(args []string, getenv func(string) string) (string, error) {
	fs := flag.NewFlagSet("bootstrap", flag.ContinueOnError)
	databaseURL := fs.String("database-url", getenv("DATABASE_URL"), "Postgres DSN (or DATABASE_URL)")
	if err := fs.Parse(args); err != nil {
		return "", err
	}
	if *databaseURL == "" {
		return "", fmt.Errorf("missing required value: -database-url or DATABASE_URL")
	}
	return *databaseURL, nil
}

func run(databaseURL string) error {
	loggerConfig, err := logger.LoadConfig()
	if err != nil {
		return err
	}
	if err := logger.Configure(loggerConfig); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	pool, err := database.NewPool(ctx, database.Config{URL: databaseURL})
	if err != nil {
		return err
	}
	defer pool.Close()

	return bootstrap.Seed(ctx, pool)
}
