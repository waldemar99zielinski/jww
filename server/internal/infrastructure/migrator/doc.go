// Package migrator is a small, dependency-free (stdlib + pgx only) SQL
// migration runner for Postgres.
//
// Each migration is a pair of files, "<version>_<name>.up.sql" and
// "<version>_<name>.down.sql". The concatenation of both halves is
// checksummed (sha256) and the checksum is recorded in the schema_migrations
// table when a migration is applied. Every later run reloads the files and
// compares their checksum against what was recorded, so a migration edited
// after being applied is detected and blocks Up instead of silently drifting
// from the database.
//
// It speaks pgx directly (no database/sql, no driver abstraction) since this
// project only ever targets Postgres. A Postgres advisory lock serializes
// concurrent callers (e.g. two server instances booting at once) so they
// don't race applying the same pending migration twice.
package migrator
