// Package migrations embeds the goose-format SQL files into the server binary.
//
// Embedding lets cmd/server apply pending migrations on startup (via
// internal/infrastructure/database.Migrate) without shipping the .sql files or
// requiring the goose CLI. The same files still work with the CLI:
// `goose -dir migrations up`.
package migrations

import "embed"

// FS holds every goose migration (*.sql) at the root of the embedded tree.
//
//go:embed *.sql
var FS embed.FS
