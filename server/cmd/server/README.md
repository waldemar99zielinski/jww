# cmd/server

**Clean Architecture layer: Frameworks & Drivers (composition root; outermost).** The only place that depends on every layer.

**Role.** The composition root and process entry point. It loads `.env` + config, configures the logger, builds the pgx pool, applies the embedded goose migrations, constructs the whole dependency graph **exactly once**, asks `http_server` for a fully-mounted server, runs it, and shuts down gracefully. This is the only package allowed to know every concrete adapter.

## Files

- `main.go` — `run()` performs, in order: `loadDotEnv(".env")` → `LoadConfig()` → `logger.Configure(cfg.Logger)` → `signal.NotifyContext(SIGINT, SIGTERM)` → `database.NewPool(ctx, cfg.Database)` (`defer pool.Close()`) → `database.Migrate(ctx, pool)` (applies embedded goose migrations) → `buildDeps(...)` → `httpserver.New(deps, cfg.Http)` → serve in a goroutine (select on `errCh` vs `ctx.Done()`) → `srv.Shutdown(10s timeout)`. `main()` logs and `os.Exit(1)` on error.
- `config.go` — `type Config { Http http_server.Config; Database database.Config; Auth security.Config; Logger logger.Config }` and `LoadConfig()` which calls each module's own loader and aggregates.
- `deps.go` — `buildDeps(pool, authConfig, corsOrigins, maxBodyBytes) *kernel.Deps`: builds the UoW manager, hasher, token service, all repositories, all services, all workflows, then fills `kernel.Deps`.
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
