package migrator

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DownN reverses the last n applied migrations, most-recently-applied first,
// after verifying checksums exactly like Up does.
func DownN(ctx context.Context, pool *pgxpool.Pool, fsys fs.FS, n int, log *slog.Logger) ([]Migration, error) {
	log = orDefault(log)

	if n <= 0 {
		return nil, nil
	}

	migrations, err := Load(fsys)
	if err != nil {
		return nil, err
	}
	byVersion := make(map[string]Migration, len(migrations))
	for _, m := range migrations {
		byVersion[m.Version] = m
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

	if n > len(applied) {
		n = len(applied)
	}

	var reversed []Migration
	for i := len(applied) - 1; i >= len(applied)-n; i-- {
		m, ok := byVersion[applied[i].Version]
		if !ok {
			return reversed, fmt.Errorf("%w: version %s", ErrMissingFile, applied[i].Version)
		}
		if err := applyDown(ctx, conn, m); err != nil {
			return reversed, fmt.Errorf("migrator: reversing %s_%s: %w", m.Version, m.Name, err)
		}
		log.Info("migration reversed", "version", m.Version, "name", m.Name)
		reversed = append(reversed, m)
	}
	return reversed, nil
}
