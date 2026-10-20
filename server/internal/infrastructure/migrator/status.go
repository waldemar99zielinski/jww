package migrator

import (
	"context"
	"fmt"
	"io/fs"
	"sort"

	"github.com/jackc/pgx/v5/pgxpool"
)

// StatusEntry describes one migration's state relative to the database.
type StatusEntry struct {
	Migration Migration
	Applied   bool
	// Drifted is true when the migration was applied but its on-disk
	// checksum no longer matches the one recorded at apply time.
	Drifted bool
	// Missing is true when the migration was applied but fsys no longer has
	// its files — Migration.Up/Down are empty in this case; only Version,
	// Name and Checksum (as recorded at apply time) are populated.
	Missing bool
}

// Status reports, for every migration in fsys plus every migration recorded
// as applied in the database, whether it has been applied, whether its files
// have drifted since, or whether its files are gone entirely. Unlike
// Up/DownN it does not error on any of this — it reports it, for `migrate
// status` to display.
func Status(ctx context.Context, pool *pgxpool.Pool, fsys fs.FS) ([]StatusEntry, error) {
	migrations, err := Load(fsys)
	if err != nil {
		return nil, err
	}

	conn, err := pool.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("migrator: acquiring connection: %w", err)
	}
	defer conn.Release()

	if err := ensureSchema(ctx, conn); err != nil {
		return nil, err
	}

	applied, err := readRecords(ctx, conn)
	if err != nil {
		return nil, err
	}
	byVersion := make(map[string]Record, len(applied))
	for _, r := range applied {
		byVersion[r.Version] = r
	}

	seen := make(map[string]bool, len(migrations))
	entries := make([]StatusEntry, 0, len(migrations)+len(applied))
	for _, m := range migrations {
		record, isApplied := byVersion[m.Version]
		seen[m.Version] = true
		entries = append(entries, StatusEntry{
			Migration: m,
			Applied:   isApplied,
			Drifted:   isApplied && record.Checksum != m.Checksum,
		})
	}

	for _, r := range applied {
		if seen[r.Version] {
			continue
		}
		entries = append(entries, StatusEntry{
			Migration: Migration{Version: r.Version, Name: r.Name, Checksum: r.Checksum},
			Applied:   true,
			Missing:   true,
		})
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Migration.Version < entries[j].Migration.Version
	})
	return entries, nil
}
