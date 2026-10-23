# internal/adapter/gateway

**Clean Architecture layer: Interface Adapter (outbound/driven).** Implements the `Repository` ports declared by the use-case layer.

**Role.** The **only** place that translates between domain entities and database rows. Implements each service's `Repository` port against Postgres using `pgx` + Squirrel (no ORM).

## Shared plumbing

- `db.go` — `DBTX` interface (the subset shared by `*pgxpool.Pool` and `pgx.Tx`); package-wide `var psql = newStatementBuilder()`; `dbFromUnitOfWork(defaultDB, uow...)` selects the active tx or the default handle (errors if >1 uow).
- `squirrel.go` — `newStatementBuilder()` → `PlaceholderFormat(squirrel.Dollar)` (write `?`, it emits `$N`).
- `numeric.go` — carries the package doc; `moneyToNumeric`/`numericToMoney`, `exchangeRateToNumeric`/`numericToExchangeRate` always round-trip through the domain type's **text** form, never `float64`.
- `health.go` — `PostgresHealthChecker` + `NewPostgresHealthChecker(pool)` implements `health.Checker` with `pgxpool.Pool.Ping`; readiness wiring uses this adapter instead of reaching into the pool directly.

## Rules

- **Files:** `postgres_<entity>_repository.go`.- **Types:** `Postgres<Entity>Repository{ db DBTX }`; `NewPostgres<Entity>Repository(db DBTX) *Postgres<Entity>Repository`.
- **Method signatures mirror the service port exactly**, `ctx domain.AppContext` first, trailing variadic `uow ...unit_of_work.UnitOfWork`.
- **Every method starts** with:
  ```go
  db, err := dbFromUnitOfWork(r.db, uow...)
  if err != nil { return ..., err }
  ```
  and passes `ctx.Context()` to pgx (never `ctx`).
- **SQL is always Squirrel** via `psql`, ending in `.ToSql()`; always check the returned `err`. Writes use `Exec`; single reads `QueryRow(...).Scan(...)`; lists `Query` + `defer rows.Close()` + `for rows.Next()` + final `rows.Err()`.
- **Not found:** `errors.Is(err, pgx.ErrNoRows)` → `return nil, nil`. Gateways never invent domain errors.
- **Money/rates:** scan into `pgtype.Numeric` and convert via `numeric.go`. **Enums:** scan into a `string`, then cast to the domain type.
- Factored SELECT/JOIN shapes live in package consts (`transactionSelectColumns`, `transactionFrom`, …); they begin with a newline and are used with `[1:]`.
- List ordering is explicit; conditional filters are appended only when non-nil.
- Reuse a `scan<Entity>Row(row pgx.Row)` helper when >1 method scans the same shape. A gateway may implement **several** service ports.
- Nullable scalars (e.g. `transaction_attachments.transaction_id`) are handled with `pgtype.Int8` on read and a `nullableInt64` helper on write, so a nil pointer becomes SQL `NULL` rather than a typed nil arg.
- `LockByIDs` uses `.OrderBy("a.id").Suffix("FOR UPDATE")` so concurrent claims lock rows in one order; `LinkToTransaction` is conditional (`WHERE transaction_id IS NULL`) and errors if the affected-row count does not match the request.

## How to add a repository

1. Confirm the target service package already defines the `Repository` interface.
2. Add `postgres_<entity>_repository.go`, struct + constructor taking `DBTX`.
3. Add `scan<Entity>Row` if needed (handle `pgx.ErrNoRows` → `nil, nil`; cast enums; convert numerics).
4. Implement each port method using the idioms above.
5. Wire the constructor in `cmd/server/deps.go`.

## Avoid

- `float64` money / string interpolation.
- Returning `pgx.ErrNoRows` to callers.
- SQL or column names outside this package; business logic / `Validate()` here.
- Literal `$1` placeholders; skipping the `ToSql()` error check; ignoring write errors.
- Imports: pgx, squirrel, `domain`, `unit_of_work`, `uuid`, and the specific service whose port (or port-adjacent type like `transaction.Filter`) you implement — nothing else internal.

## Testing requirement for this directory

**Location:** `tests/` subdirectory (black-box, imports this package's
exported API — see `server/TESTING.md#conventions`).

Black-box (`package gateway_test`), **integration** (`//go:build integration`),
a real Postgres via [`internal/tests/pgtest`](../../tests/pgtest/README.md)
— never a mocked `DBTX`/`pgx.Tx`. One container per test binary (shared
`TestMain`); per-test isolation via `pgtest.BeginRollback`, not `TRUNCATE`.
Run with `go test -tags=integration ./internal/adapter/gateway/...` (needs
Docker). Every repository method needs: a round-trip (Create+Get) test,
`(nil, nil)` on a genuine miss, and a constraint-violation test for any
unique/FK/CHECK constraint the migration declares on its table.
`UpsertMany` additionally needs: a fresh batch reporting `Inserted = N`; a
re-run of the same pairs+date reporting `Updated = N` while still leaving
exactly one row per pair (the bulk `ON CONFLICT … DO UPDATE` replaces, never
duplicates); and a mixed batch splitting the counts. Two gotchas
worth remembering (both bit earlier passes at this suite):

- A constraint violation aborts the rest of that Postgres transaction
  (`SQLSTATE 25P02` on any further statement) — don't reuse one `uow` across
  a deliberately-failing `Create` and a must-succeed one; split into two
  tests, each with its own `BeginRollback`.
- `transaction_categories` and similar reference tables carry migration-seeded
  rows (see `migrations/00006`) — assert containment/exclusion of the IDs
  your test created, never an exact total count.
- `time.Time` round-trips as an equal instant but not necessarily an
  identical `*time.Location` value — compare via `.Equal()`, not
  `assert.Equal` on the whole struct.

See `server/TESTING.md` for the full strategy.
