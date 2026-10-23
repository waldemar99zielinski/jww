package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"jww/internal/infrastructure/migrator"
	"jww/migrations"
	"jww/utils/logger"
)

var log = logger.New().WithComponent("database")

// Migrate applies every embedded migration that has not run yet, tracking
// versions (and a checksum per version) in schema_migrations — so inline
// startup migrations and manual `migrate up` runs stay in sync, and a
// migration file edited after being applied is detected instead of silently
// drifting from the database.
//
// It is called once at process startup, right after the pool is created, so
// the schema is current before the first request is served. Applying an
// up-to-date database is a no-op.
func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	applied, err := migrator.Up(ctx, pool, migrations.FS, log.Slog())
	if err != nil {
		return fmt.Errorf("running migrations: %w", err)
	}
	if len(applied) == 0 {
		log.Slog().Info("database schema up to date")
	}
	return nil
}
