# internal/infrastructure/database

**Clean Architecture layer: Frameworks & Drivers.** Implements the `unit_of_work` port declared by the use-case layer.

**Role.** Wires the Postgres connection pool and implements the concrete **Unit of Work** on top of a pgx transaction. The only package that reaches the DB via pgx.

## Files

- `connection.go` (package doc) — `NewPool(ctx context.Context, cfg Config) (*pgxpool.Pool, error)`: `pgxpool.New` then `Ping`; closes the pool and wraps the error if ping fails.
- `migrate.go` — `Migrate(ctx context.Context, pool *pgxpool.Pool) error`: calls `migrator.Up` against the embedded `migrations.FS`, applying pending migrations. Called once at startup right after `NewPool`.
- `unit_of_work.go` — `PostgresUnitOfWorkManager` + `NewPostgresUnitOfWorkManager(pool)`; `Create(ctx domain.AppContext)` begins a tx (`pool.Begin(ctx.Context())`); `PostgresUnitOfWork{ tx pgx.Tx }` implements `Finish` (commit), `Restore` (rollback), `GetValue()` (returns the raw `pgx.Tx`).
- `config.go` — `Config{ URL string \`env:"DATABASE_URL" bind:"required"\` }`; `LoadConfig()`.

## Rules

- The UoW ports are defined in `internal/service/unit_of_work`; this package only implements them.
- `Create` returns the **interface** `unit_of_work.UnitOfWork`, never the concrete type.
- `ctx domain.AppContext` is threaded through so cancellation/context propagates.
- `Migrate` is the only migration entry point; it runs `internal/infrastructure/migrator` against the embedded `migrations.FS` and shares the `schema_migrations` table with `cmd/migrate`.

## How to extend

- New persistence config: add tagged fields to `Config`.
- New concrete repositories belong in `internal/adapter/gateway`, not here.
- New UoW backend: add a manager + UoW here implementing the same ports.
- New schema change: add a numbered SQL file under `migrations/` (never here).

## Avoid

- Leaking `pgx` types past the `GetValue() interface{}` boundary.
- Starting transactions in handlers — services/workflows own demarcation.
- Imports beyond pgx/stdlib, `migrator`, `migrations`, `domain`, `service/unit_of_work`, `utils/{configuration,logger}`.

## Testing requirement for this directory

No tests of its own. `NewPool`/`Migrate`/`PostgresUnitOfWorkManager` are
exercised for real by every consumer of
[`internal/tests/pgtest`](../../tests/pgtest/README.md) — its own test
suite, every `internal/adapter/gateway` integration test (`TestMain` calls
`pgtest.Start`, which calls `NewPool` then `Migrate`), and `server/tests/e2e`'s
suite (same, plus the boot smoke test exercising this package exactly as
`cmd/server`'s real `run()` does). If you change this package's behavior,
verify it by running those suites (`go test -tags=integration
./internal/adapter/gateway/...` and `go test -tags=e2e ./tests/e2e/...`,
both need Docker) rather than adding a test here.
