# cmd/migrate

**Clean Architecture layer: Frameworks & Drivers (composition root; outermost).**

**Role.** The manual CLI for the migrations embedded in `jww/migrations` — the same files
`internal/infrastructure/database.Migrate` applies automatically at server startup. Lets an operator
run/reverse/inspect migrations without starting the whole server.

## Files

- `main.go` — `run(args)` configures the logger, loads `database.Config` (reads `DATABASE_URL` straight
  from the real process environment via `utils/configuration` — no `.env` parsing here; export the var
  or source your `.env` yourself before running), opens a pool, then dispatches on `args[0]`:
  - `up` — `migrator.Up`
  - `down [n]` — `migrator.DownN`, `n` defaults to `1`
  - `status` — `migrator.Status`, prints `pending` / `applied` / `APPLIED (checksum mismatch!)` per
    migration
  - `baseline` — `migrator.Baseline`, see `migrations/README.md#cutting-over-an-existing-database`

## How to use

```bash
export DATABASE_URL=postgres://...
go run ./cmd/migrate up
go run ./cmd/migrate status
go run ./cmd/migrate down       # reverses the single most-recently-applied migration
go run ./cmd/migrate down 3     # reverses the last 3
go run ./cmd/migrate baseline   # one-time cutover of an already-up-to-date database, see migrations/README.md
```

## Rules

- Same startup shape as `cmd/server` minus dotenv (config → logger → pool), minus the HTTP server and
  scheduler — this binary never builds `kernel.Deps`.
- `main()` prints errors to stderr and exits 1; no retries, no partial-success handling beyond what
  `migrator` itself returns.

## Avoid

- Importing anything from `cmd/server` — binaries don't import each other (`cmd/README.md`).
- Adding business logic here — this package only wires `migrator` to a real `DATABASE_URL`.

## Testing requirement for this directory

No tests of its own — thin wiring over `internal/infrastructure/migrator`, which is exercised for real
by `migrations/`'s and `internal/infrastructure/database/`'s testing requirements. If you change the
subcommand dispatch, verify manually against a local Postgres (`go run ./cmd/migrate status` etc.)
rather than adding a test here.
