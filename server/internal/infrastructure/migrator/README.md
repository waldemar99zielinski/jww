# internal/infrastructure/migrator

**Clean Architecture layer: Frameworks & Drivers.** A generic SQL migration runner; knows nothing about
this project's domain.

**Role.** Applies, reverses, and reports on versioned SQL migrations stored as
`<version>_<name>.{up,down}.sql` file pairs in any `fs.FS`, tracking applied versions and a checksum
per version in a `schema_migrations` table. Pure stdlib + `pgx/v5` — no third-party migration library.

## Files

One file per entry point, plus shared pieces:

- `doc.go` — package doc only.
- `migration.go` — `Migration{Version, Name, Up, Down, Checksum}`, `Record{Version, Name, Checksum}`
  (the `schema_migrations` row shape), and `checksum()` (`sha256(Up+"\x00"+Down)`, hex-encoded).
- `filename.go` — `parseFilename(name) (version, name, direction string, ok bool)`, the
  `^(\d{5})_([a-z0-9_]+)\.(up|down)\.sql$` pattern.
- `load.go` — `Load(fsys) ([]Migration, error)`: parses `<version>_<name>.{up,down}.sql` pairs, errors
  if a version is missing either half.
- `errors.go` — `ErrChecksumMismatch`, `ErrMissingFile`, and `verify(migrations, applied)`: every
  already-applied version must still exist on disk with a matching checksum, or it fails before
  anything else runs. Called by `Up`/`DownN`/`Baseline` (not `Status`, which reports drift instead of
  failing on it).
- `store.go` — the shared low-level pieces every entry point uses: `createTableSQL`/`ensureSchema`,
  `readRecords`, `applyUp`/`applyDown` (each wraps one migration's DDL + its `schema_migrations`
  row in a single transaction), `lock`/`unlock` (the Postgres advisory lock), `orDefault` (nil-logger
  fallback).
- `up.go` / `down.go` / `status.go` / `baseline.go` — one file per exported entry point:
  - `Up` applies every pending migration in order.
  - `DownN` reverses the last `n` applied migrations, most-recent first.
  - `Status` reports applied/pending/drifted per migration without erroring on drift, and also reports
    applied migrations whose files no longer exist in `fsys` (`StatusEntry.Missing`) — these carry only
    `Version`/`Name`/`Checksum` from `schema_migrations`, since there's no file left to read `Up`/`Down`
    from.
  - `Baseline` stamps every migration as applied (recording its checksum) without running any SQL;
    refuses to run if `schema_migrations` already has rows. For cutting an existing database over onto
    this tracking table — see `migrations/README.md`.

  All but `Status` take the advisory lock, via the same shape: acquire a connection, lock, ensure the
  schema, read + verify existing records, do the work, unlock (deferred, so it always runs).

`schema_migrations` itself has `version` (text, the zero-padded filename prefix, e.g. `"00001"`), `name`
(text, e.g. `"create_users_table"` — so `psql` alone tells you what ran without cross-referencing code),
`checksum`, and `applied_at`.

## Rules

- Postgres-only; this project has exactly one target database, so there's no `Driver` abstraction to
  support other engines.
- A migration's checksum covers both halves (`Up` and `Down`) — editing either after the migration has
  applied is drift, not just editing the half that already ran.
- `Up`/`DownN`/`Baseline` all take the advisory lock before touching `schema_migrations`; never add a
  code path that writes to that table without it.
- Each migration applies in its own transaction — never batch multiple migrations into one transaction
  (a later migration's failure must not be able to roll back an earlier, already-logically-committed one
  once `Up` has moved on).

## How to extend

- New reporting field on `Status`: add it to `StatusEntry`, keep `Up`/`DownN` unaffected.
- Non-Postgres target: out of scope for this project; would need a real `Driver` abstraction first.

## Avoid

- Importing this package's SQL constants from outside — `migrations/` is the only caller that should
  know the file-pair format exists.
- Executing `Migration.Up`/`.Down` with query arguments — pgx only uses the simple protocol (which
  supports multi-statement scripts) when a call has zero arguments; see `applyUp`/`applyDown`.

## Testing requirement for this directory

No tests of its own. Exercised for real every time `internal/infrastructure/database.Migrate` runs —
see that package's and `migrations/`'s testing requirements (`pgtest`, gateway integration tests, e2e
boot smoke test). If you change this package's behavior, verify with `go test -tags=integration
./internal/adapter/gateway/...` and `go test -tags=e2e ./tests/e2e/...` (both need Docker).
