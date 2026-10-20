package migrator

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Baseline marks every migration in fsys as already applied, without running
// any SQL. It exists for a one-time cutover of a database whose schema was
// brought up to date by some other means (e.g. a prior migration tool) onto
// this package's tracking table. It errors if schema_migrations already has
// any rows, to avoid silently overwriting real history.
func Baseline(ctx context.Context, pool *pgxpool.Pool, fsys fs.FS, log *slog.Logger) ([]Migration, error) {
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
	if len(applied) > 0 {
		return nil, fmt.Errorf("migrator: refusing to baseline, schema_migrations already has %d row(s)", len(applied))
	}

	for _, m := range migrations {
		if _, err := conn.Exec(ctx, `INSERT INTO schema_migrations (version, name, checksum) VALUES ($1, $2, $3)`, m.Version, m.Name, m.Checksum); err != nil {
			return nil, fmt.Errorf("migrator: baselining %s_%s: %w", m.Version, m.Name, err)
		}
		log.Info("migration baselined", "version", m.Version, "name", m.Name)
	}
	return migrations, nil
}
