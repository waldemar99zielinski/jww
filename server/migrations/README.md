# migrations

**Role.** Source of truth for the Postgres schema. Plain SQL files embedded into the server binary
(`embed.go`) and applied on startup by `internal/infrastructure/database.Migrate` (which calls
`internal/infrastructure/migrator`); the same files also run with the `cmd/migrate` CLI for manual
up/down/status/baseline.

## Files

Numbered, applied in numeric order: `00001_create_transaction_categories_table` …
`00009_create_webauthn_credentials_table`. Each version is a pair of files, `<version>_<name>.up.sql`
and `<version>_<name>.down.sql` — not goose-style `+goose Up`/`+goose Down` sections in one file.
Later migrations reuse earlier enum types (e.g. `transactions.type` reuses
`transaction_category_type`), so ordering matters. `00006` is a data seed (fixed UUIDv5 ids,
`ON CONFLICT ... DO NOTHING`) rather than DDL — see its own header comment for the derivation. `00007`
adds `transaction_attachments`: a nullable `transaction_id` (the whole "unattached draft" state), no
ordinal/count column, and a partial index for the attached-list query. `00008` creates
`transaction_sub_transactions` with a non-zero (not merely positive) amount CHECK, so discount/coupon
line items can be negative. `00009` creates `webauthn_credentials` including `backup_eligible` from the
start. `embed.go` holds the `//go:embed *.sql` directive and the exported `FS` consumed by
`database.Migrate`.

Previously this directory had 11 goose-format migrations; `00008`/`00009` here are each a merge of two
of those (an original `CREATE TABLE` plus a later `ALTER TABLE` against the same table) — see "Squashing"
below.

## Rules

- **File name:** zero-padded 5-digit sequence + snake_case verb, two files per version:
  `NNNNN_create_<table>_table.up.sql` / `NNNNN_create_<table>_table.down.sql` (or `_add_`/`_alter_`).
- **Up** starts with a comment explaining purpose + which invariants are DB-enforced vs app-enforced,
  then DDL.
- **Down** reverses Up exactly, in reverse dependency order, using `IF EXISTS`.
- **Never edit an applied migration** — its checksum (sha256 of Up+Down) is recorded in
  `schema_migrations` when it runs; editing it later makes `migrate up`/the server's startup `Migrate`
  fail loudly instead of silently drifting. Add a new numbered pair instead.
- **PKs:** `uuid PRIMARY KEY DEFAULT gen_random_uuid()`; natural keys where meaningful (e.g.
  `currencies.code char(3)`); `bigint GENERATED ALWAYS AS IDENTITY` for append-only logs.
- **Named constraints** everywhere: `<table>_<column>_<rule>`; unique `<table>_<col>_uq`; FK
  `<table>_..._fk`; checks `btrim(col) <> ''`.
- **Money/rates:** `numeric(12,2)` / `numeric(18,4)`, never `float`/`double precision`. Timestamps:
  `timestamptz NOT NULL DEFAULT now()`.
- **Indexes:** `<table>_<cols>_idx`, each with a comment naming the query it serves.
- Ensure `CREATE EXTENSION IF NOT EXISTS pgcrypto;` exists for `gen_random_uuid()`.
- Keep the schema consistent with `internal/domain/*.Validate()` and `internal/adapter/schema`.

## How to add a migration

1. Pick the next number (list the directory).
2. Write `NNNNN_<name>.up.sql` and the matching `IF EXISTS` `NNNNN_<name>.down.sql`.
3. Restart the server — the new pair is embedded and applied automatically. To apply/verify without
   starting the server:
   ```bash
   export DATABASE_URL=postgres://...
   go run ./cmd/migrate up && go run ./cmd/migrate status
   ```

## Squashing

A migration that only adds to or alters the table created by the immediately preceding migration (not
one touched by data already living in other tables) may be folded into that earlier migration instead of
staying a separate numbered step — do this only while cutting over the whole directory's numbering (see
below), never by silently editing an already-applied migration in place. `00008` and `00009` were squashed
this way once, each absorbing one later `ALTER TABLE` against the table its `CREATE TABLE` had just
defined. Don't squash a seed/data migration into a schema migration, or two migrations that touch
different tables.

## Cutting over an existing database

`schema_migrations` (created automatically by `migrator`) starts empty. A database already brought up to
date by a prior tool (or a prior version of this directory) needs to be told "these migrations already
ran" without re-running their DDL:

```bash
go run ./cmd/migrate baseline
```

This stamps every migration currently in this directory as applied (recording its checksum) without
executing any SQL, and refuses to run if `schema_migrations` already has rows. Only run it once, against
a database whose schema you've confirmed already matches what these files would produce. Afterwards you
may drop any old tracking table left behind by a previous tool (e.g. `DROP TABLE goose_db_version;`).

## Avoid

- Editing/renumbering existing migrations.
- `serial`/`bigserial`; floats for money; unnamed constraints; missing `.down.sql`; `timestamp` without
  time zone; `created_at` without `DEFAULT now()`.

## Testing requirement for this directory

No dedicated test file — every migration is applied for real, in order,
against a real Postgres by [`internal/tests/pgtest`](../internal/tests/pgtest/README.md)
(used by `internal/adapter/gateway`'s and `e2e`'s suites) and by
`pgtest`'s own test. A broken migration fails `go test -tags=integration
./internal/adapter/gateway/...` or `go test -tags=e2e ./tests/e2e/...` outright
(both need Docker) — that failure *is* this directory's test. If you add a
constraint, make sure at least one gateway test exercises violating it (see
`internal/adapter/gateway/README.md`'s testing requirement).
