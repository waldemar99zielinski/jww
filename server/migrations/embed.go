// Package migrations embeds the SQL migration files into the server binary.
//
// Embedding lets cmd/server apply pending migrations on startup (via
// internal/infrastructure/database.Migrate, which calls
// internal/infrastructure/migrator) without shipping the .sql files
// separately. The same files also work with the CLI: `go run ./cmd/migrate
// up` (see cmd/migrate/README.md).
package migrations

import "embed"

// FS holds every migration file (*.sql) at the root of the embedded tree:
// paired "<version>_<name>.up.sql" / "<version>_<name>.down.sql" files.
//
//go:embed *.sql
var FS embed.FS
