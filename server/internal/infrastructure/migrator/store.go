package migrator

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)

// advisoryLockKey scopes the Postgres advisory lock to this package; the
// exact value only needs to be fixed and unlikely to collide with locks
// taken elsewhere in the same database.
const advisoryLockKey int64 = 0x6a77775f6d6967 // "jww_mig" read as bytes

const createTableSQL = `
CREATE TABLE IF NOT EXISTS schema_migrations (
    version     text PRIMARY KEY,
    name        text NOT NULL,
    checksum    text NOT NULL,
    applied_at  timestamptz NOT NULL DEFAULT now()
)`

func ensureSchema(ctx context.Context, conn *pgxpool.Conn) error {
	if _, err := conn.Exec(ctx, createTableSQL); err != nil {
		return fmt.Errorf("migrator: creating schema_migrations table: %w", err)
	}
	return nil
}

func readRecords(ctx context.Context, conn *pgxpool.Conn) ([]Record, error) {
	rows, err := conn.Query(ctx, `SELECT version, name, checksum FROM schema_migrations ORDER BY version`)
	if err != nil {
		return nil, fmt.Errorf("migrator: reading applied migrations: %w", err)
	}
	defer rows.Close()

	var records []Record
	for rows.Next() {
		var r Record
		if err := rows.Scan(&r.Version, &r.Name, &r.Checksum); err != nil {
			return nil, fmt.Errorf("migrator: scanning applied migration: %w", err)
		}
		records = append(records, r)
	}
	return records, rows.Err()
}

func applyUp(ctx context.Context, conn *pgxpool.Conn, m Migration) error {
	tx, err := conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, m.Up); err != nil {
		return fmt.Errorf("executing up script: %w", err)
	}
	if _, err := tx.Exec(ctx, `INSERT INTO schema_migrations (version, name, checksum) VALUES ($1, $2, $3)`, m.Version, m.Name, m.Checksum); err != nil {
		return fmt.Errorf("recording migration: %w", err)
	}
	return tx.Commit(ctx)
}

func applyDown(ctx context.Context, conn *pgxpool.Conn, m Migration) error {
	tx, err := conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, m.Down); err != nil {
		return fmt.Errorf("executing down script: %w", err)
	}
	if _, err := tx.Exec(ctx, `DELETE FROM schema_migrations WHERE version = $1`, m.Version); err != nil {
		return fmt.Errorf("deleting migration record: %w", err)
	}
	return tx.Commit(ctx)
}

func lock(ctx context.Context, conn *pgxpool.Conn) error {
	if _, err := conn.Exec(ctx, `SELECT pg_advisory_lock($1)`, advisoryLockKey); err != nil {
		return fmt.Errorf("migrator: acquiring advisory lock: %w", err)
	}
	return nil
}

func unlock(ctx context.Context, conn *pgxpool.Conn, log *slog.Logger) {
	if _, err := conn.Exec(ctx, `SELECT pg_advisory_unlock($1)`, advisoryLockKey); err != nil {
		log.Error("releasing advisory lock", "error", err.Error())
	}
}

func orDefault(log *slog.Logger) *slog.Logger {
	if log == nil {
		return slog.Default()
	}
	return log
}
