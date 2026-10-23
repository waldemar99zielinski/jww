# scripts

**Clean Architecture layer: Frameworks & Drivers (composition roots; outermost).** Standalone, run-on-demand tools — **not** part of the server binary.

**Role.** Maintenance programs that operate on the system from outside the HTTP API (bootstrapping accounts, seeding a dev database, repairs). Each tool is a package under `scripts/<tool>/` with its **own** configuration and its command in `scripts/<tool>/cmd/`. Nothing here is wired into or imported by `cmd/server`, and the server never imports anything from `scripts/`.

## Contents

- [`signup/`](signup/README.md) — create a user directly in Postgres (Argon2id-hashed), bypassing the HTTP signup endpoint. Reusable `Client` in the package.
- [`bootstrap/`](bootstrap/README.md) — seed the fixed local test accounts (`julcia`, `wozi`, `baby`, each password = username); idempotent. Reuses `signup`'s `Client`.
- [`syncrates/`](syncrates/README.md) — backfill currency exchange rates for a date range by running the same workflow the nightly scheduler runs, one day at a time; weekends skipped by default.

## Rules

- Each tool is a package under `scripts/<tool>/` with its command `main` in `scripts/<tool>/cmd/`. A tool may import another tool's package (e.g. `bootstrap` uses `signup`).
- Each command resolves what its `run` needs from a flag first, then the matching environment variable, and fails when neither supplies a required value — never wired into `cmd/server`'s aggregate `Config`.
- Reuse the existing adapters/services/workflows (`internal/...`) instead of re-implementing DB access or hashing, so rows stay compatible with the server.
- Run from the `server/` directory: `go run ./scripts/<tool>/cmd`.

## Testing requirement for this directory

`signup`, `bootstrap`, and `addcurrency` are thin wrappers around already-tested
services (`signup`'s `Client` over `user.Service`; `bootstrap` over `signup` and
`addcurrency`), so the value of a dedicated test is lower than for the layers
covered so far. `syncrates` additionally owns range iteration and aggregation,
which is covered by a fast unit test with a fake syncer. If you add real logic
here beyond that (anything that isn't flag/env resolution or a straight service
call), it would need either a unit test like `syncrates`' or an
integration-tagged test via `internal/tests/pgtest`, the same way
`internal/adapter/gateway` and `e2e` use it.
