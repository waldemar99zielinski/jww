# cmd/server

**Clean Architecture layer: Frameworks & Drivers (composition root; outermost).** The only place that depends on every layer.

**Role.** The composition root and process entry point. It loads `.env` + config, configures the logger, builds the pgx pool, applies the embedded migrations, constructs the whole dependency graph **exactly once**, asks `http_server` for a fully-mounted server, runs it, and shuts down gracefully. This is the only package allowed to know every concrete adapter.

## Files

- `main.go` — `run()` performs, in order: `loadDotEnv(".env")` → `LoadConfig()` (including web deployment validation) → `logger.Configure(cfg.Logger)` → `signal.NotifyContext(SIGINT, SIGTERM)` → `database.NewPool(ctx, cfg.Database)` (`defer pool.Close()`) → `database.Migrate(ctx, pool)` (applies embedded migrations, see `internal/infrastructure/migrator`) → `buildDeps(...)` → `startScheduler(cfg.Scheduler, deps.CurrencyExchangeRateWorkflow)` → `httpserver.New(deps, cfg.Http, cfg.Web)` → serve in a goroutine (select on `errCh` vs `ctx.Done()`) → `stopScheduler(10s)` then `srv.Shutdown(10s timeout)`. `main()` logs and `os.Exit(1)` on error.
- `config.go` — aggregates module configs, including separate `http_server.Config` and
  `http_server/web.Config` values, and calls each module's own loader.
- `deps.go` — `buildDeps(ctx, pool, authConfig, frankfurterConfig, storageConfig, corsOrigins, maxBodyBytes, healthConfig) (*kernel.Deps, error)`: builds the UoW manager, hasher, token service, the Frankfurter `RateSource`, the S3 attachment storage adapter, all repositories, services, and workflows, plus the health service over PostgreSQL (required, via `gateway.PostgresHealthChecker`), OpenAI, S3, and Frankfurter checks, then fills `kernel.Deps`. It returns an error because constructing the storage adapter can fail (invalid AWS config).
- `scheduler.go` — `startScheduler(cfg, rateWorkflow) (stop func(context.Context), error)`: builds the exchange-rate sync scheduler and its job closure, starts it when `cfg.Scheduler.Enabled`, and returns a graceful-shutdown `stop`. The job runs `CurrencyExchangeRateWorkflow.SyncDaily`, logs each `SyncResult.Warning` at **WARN** and the run summary at **INFO**, and the `stop` logs a shutdown failure at **WARN**. Keeping it here leaves `main.go` to a couple of lines.
- `dotenv.go` — `loadDotEnv(path)`: hand-rolled parser; real environment always wins; a missing `.env` is not an error.

## Rules

- Fixed startup order: dotenv → config → logger → signal context → DB pool → migrate → deps → server → serve → shutdown.
- Construct every singleton here; never lazily build dependencies in handlers.
- Config is aggregated here; each module owns its own `Config` + `LoadConfig`.
- Only wiring — no business logic.

## How to implement

- **New service/workflow:** construct it in `buildDeps` and add a field to `kernel.Deps`.
- **New module config:** create `Config` + `LoadConfig()` in that module, add a field to `Config` here, call its loader in `LoadConfig`, and document the vars in `.env.example`.
- **New binary:** add `cmd/<name>/` mirroring this structure.

## Avoid

- `os.Getenv` reads scattered in `main` — always go through `LoadConfig`.
- Forgetting `defer pool.Close()` / `defer stop()`.
- Constructing `http.Server` by hand here — `httpserver.New` owns that.

## Testing requirement for this directory

No tests of its own — this package is pure wiring, already covered by what
it wires. `server/tests/e2e`'s boot smoke test (`TestCmdServerBoots`) is the
one place that exercises this exact composition root for real: it builds
this binary, runs it as a subprocess against a real Postgres, confirms it
serves traffic, and confirms graceful shutdown on `SIGTERM`. If you change
`run()`, `buildDeps`, or the startup order, verify with `go test -tags=e2e
./tests/e2e/...` (needs Docker) rather than adding a test here.
`tests/utils.BuildTestDeps` duplicates `buildDeps`'s wiring for the
in-process journey tests (it can't import this `package main`'s unexported
function) — keep the two in sync if this file's wiring changes; see
`server/tests/utils/README.md`.
