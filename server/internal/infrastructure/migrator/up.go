package migrator

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Up applies every pending migration in fsys, in version order, after
// verifying that no already-applied migration's files have changed. It
// returns the migrations it applied, in the order they ran.
func Up(ctx context.Context, pool *pgxpool.Pool, fsys fs.FS, log *slog.Logger) ([]Migration, error) {
	log = orDefault(log)

	migrations, err := Load(fsys)
	if err != nil {
		return nil, err
	}

	conn, err := pool.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("migrator: acquiring connection: %w", err)
	}
	defer conn.Release()

	if err := lock(ctx, conn); err != nil {
		return nil, err
	}
	defer unlock(ctx, conn, log)

	if err := ensureSchema(ctx, conn); err != nil {
		return nil, err
	}

	applied, err := readRecords(ctx, conn)
	if err != nil {
		return nil, err
	}

	if err := verify(migrations, applied); err != nil {
		return nil, err
	}

	appliedVersions := make(map[string]bool, len(applied))
	for _, r := range applied {
		appliedVersions[r.Version] = true
	}

	var ran []Migration
	for _, m := range migrations {
		if appliedVersions[m.Version] {
			continue
		}
		if err := applyUp(ctx, conn, m); err != nil {
			return ran, fmt.Errorf("migrator: applying %s_%s: %w", m.Version, m.Name, err)
		}
		log.Info("migration applied", "version", m.Version, "name", m.Name)
		ran = append(ran, m)
	}
	return ran, nil
}
