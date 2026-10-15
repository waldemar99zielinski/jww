# internal/infrastructure

**Clean Architecture layer: Frameworks & Drivers.** The outermost layer; depended on by nothing.

**Role.** Parent directory; **no `.go` files of its own**. Contains the concrete I/O infrastructure:

- [`database/`](database/README.md) — pgx pool + concrete Unit of Work.
- [`http_server/`](http_server/README.md) — the HTTP entry point (composition of routes) and `kernel`/`routers`.
- [`scheduler/`](scheduler/README.md) — cron-scheduled jobs (the only package importing the cron library).

## Rules

- Infrastructure may depend on `service` ports and `domain`; `domain`/`service` must never depend on infrastructure.
- Only `cmd/server` imports `database` and `scheduler`; the HTTP routers only ever see services via `kernel.Deps`.

## Testing requirement for this directory

No tests of its own — a parent directory with no code. See `database/` and
`http_server/`'s own READMEs, and `server/TESTING.md`.
