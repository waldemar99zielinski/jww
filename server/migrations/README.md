# migrations

**Role.** Source of truth for the Postgres schema. Goose-format SQL embedded into the server binary
(`embed.go`) and applied on startup by `internal/infrastructure/database.Migrate`; the same files run
with the `goose` CLI for manual up/down/status.

## Files

Numbered, applied in numeric order: `00001_create_transaction_categories_table.sql` … `00008_restrict_transaction_attachment_deletion.sql`. Later migrations reuse earlier enum types (e.g. `transactions.type` reuses `transaction_category_type`), so ordering matters. `00006` is a data seed (fixed UUIDv5 ids, `ON CONFLICT ... DO NOTHING`) rather than DDL — see its own header comment for the derivation. `00007` adds `transaction_attachments`: a nullable `transaction_id` (the whole "unattached draft" state), no ordinal/count column, and a partial index for the attached-list query. `00008` makes transaction deletion restrictive so attachment object and metadata cleanup cannot be bypassed. `embed.go` holds the `//go:embed *.sql` directive and the exported `FS` consumed by `database.Migrate`.

## Rules

- **File name:** zero-padded 5-digit sequence + snake_case verb: `NNNNN_create_<table>_table.sql` (or `_add_`/`_alter_`).
- Every file has `-- +goose Up` and `-- +goose Down`.
- **Up** starts with a comment explaining purpose + which invariants are DB-enforced vs app-enforced, then DDL.
- **Down** reverses Up exactly, in reverse dependency order, using `IF EXISTS`.
- **Never edit an applied migration** — add a new numbered file.
- **PKs:** `uuid PRIMARY KEY DEFAULT gen_random_uuid()`; natural keys where meaningful (e.g. `currencies.code char(3)`); `bigint GENERATED ALWAYS AS IDENTITY` for append-only logs.
- **Named constraints** everywhere: `<table>_<column>_<rule>`; unique `<table>_<col>_uq`; FK `<table>_..._fk`; checks `btrim(col) <> ''`.
- **Money/rates:** `numeric(12,2)` / `numeric(18,4)`, never `float`/`double precision`. Timestamps: `timestamptz NOT NULL DEFAULT now()`.
- **Indexes:** `<table>_<cols>_idx`, each with a comment naming the query it serves.
- Ensure `CREATE EXTENSION IF NOT EXISTS pgcrypto;` exists for `gen_random_uuid()`.
- Keep the schema consistent with `internal/domain/*.Validate()` and `internal/adapter/schema`.

## How to add a migration

1. Pick the next number (list the directory).
2. Write Up (comment + DDL) and the matching `IF EXISTS` Down.
3. Restart the server — the new file is embedded and applied automatically. To apply/verify without
   starting the server, use the CLI:
   ```bash
   export GOOSE_DRIVER=postgres GOOSE_DBSTRING="$DATABASE_URL"
   goose -dir migrations up && goose -dir migrations status
   ```

## Avoid

- Editing/renumbering existing migrations.
- `serial`/`bigserial`; floats for money; unnamed constraints; missing `Down`; `timestamp` without time zone; `created_at` without `DEFAULT now()`.

## Testing requirement for this directory

No dedicated test file — every migration is applied for real, in order,
against a real Postgres by [`internal/tests/pgtest`](../internal/tests/pgtest/README.md)
(used by `internal/adapter/gateway`'s and `e2e`'s suites) and by
`pgtest`'s own test. A broken migration fails `go test -tags=integration
./internal/adapter/gateway/...` or `go test -tags=e2e ./tests/e2e/...` outright
(both need Docker) — that failure *is* this directory's test. If you add a
constraint, make sure at least one gateway test exercises violating it (see
`internal/adapter/gateway/README.md`'s testing requirement).
