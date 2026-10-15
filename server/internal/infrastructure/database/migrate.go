package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"jww/migrations"
	"jww/utils/logger"
)

var log = logger.New().WithComponent("database")

// Migrate applies every embedded goose migration that has not run yet, tracking
// versions in the same `goose_db_version` table the goose CLI uses — so inline
// startup migrations and manual `goose -dir migrations up` stay in sync.
//
// It is called once at process startup, right after the pool is created, so the
// schema is current before the first request is served. Applying an up-to-date
// database is a no-op.
func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	// goose speaks database/sql; borrow the existing pgx pool instead of opening
	// a second connection. Closing the *sql.DB does not close the pool.
	db := stdlib.OpenDBFromPool(pool)
	defer db.Close()

	provider, err := goose.NewProvider(
		goose.DialectPostgres,
		db,
		migrations.FS,
		goose.WithSlog(log.Slog()),
	)
	if err != nil {
		return fmt.Errorf("creating migration provider: %w", err)
	}

	results, err := provider.Up(ctx)
	if err != nil {
		return fmt.Errorf("running migrations: %w", err)
	}
	if len(results) == 0 {
		log.Slog().Info("database schema up to date")
		return nil
	}
	for _, result := range results {
		log.Slog().Info("migration applied",
			"version", result.Source.Version,
			"file", result.Source.Path,
			"duration", result.Duration.String(),
		)
	}
	return nil
}
