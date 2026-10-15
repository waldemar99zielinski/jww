# internal/tests/pgtest

**Role.** Test-only helper that boots a disposable `postgres:17-alpine`
container (via testcontainers-go), applies every embedded migration, and
exposes utilities for Postgres-backed tests. Nothing outside a `_test.go`
file imports this package, and it never ships in a production binary. See
[`server/TESTING.md`](../../../TESTING.md) for the strategy this implements.

## API

- `Start(ctx) (*Pool, error)` — launches a fresh container, opens a pool,
  runs `database.Migrate`. One container per test *binary* (i.e. per
  package) — call once, not per test.
- `(*Pool) Close(ctx)` — closes the pool and terminates the container.
- `AppContext(ctx) domain.AppContext` — minimal non-HTTP `AppContext`
  (context + fresh flow id), for tests not driven by an `*http.Request`.
- `BeginRollback(tb, pool, appCtx) unit_of_work.UnitOfWork` — starts a real
  Postgres transaction and registers `tb.Cleanup` to roll it back
  unconditionally. Pass the returned `UnitOfWork` to every gateway call in a
  test so its writes are invisible to every other test sharing the
  container.

## Rules

- Build-tag every file that imports this package `//go:build integration`
  (gateway tests) or `//go:build e2e` (the `server/tests/e2e` suite) — both
  need Docker.
- One container per test *binary*, not per test — `Start` takes ~1-2s; call
  it once per package (e.g. from a shared `TestMain`), not from every
  subtest.
- Isolation is via `BeginRollback`'s transaction rollback, not `TRUNCATE` —
  this reuses the `unit_of_work` port the codebase already has instead of
  adding a second reset mechanism.
- `server/tests/e2e` needs its own, real, *committed* data — an HTTP request can't
  share a test's rolled-back transaction, since the server decides its own
  commit points. E2E tests call `Start`/`Close` directly and manage their own
  container lifecycle, not `BeginRollback`.

## Testing requirement for this directory

**Location:** alongside the source, not a `tests/` subdirectory — `pgtest_test.go`
is white-box (`package pgtest`), reaching this package's unexported state to
verify it directly. See `server/TESTING.md#conventions`.

This package's own behavior (container boots, migrations apply,
`BeginRollback` actually isolates writes) is covered by `pgtest_test.go`
(tag `integration`). Any change to `Start`, `Close`, or `BeginRollback` must
keep (or extend) that coverage — this is the one piece every other
Postgres-backed test in the repo depends on.

## Avoid

- Importing this package from anything that isn't a `_test.go` file.
- Sharing one container across packages — each package's test binary gets
  its own.
- Adding `TRUNCATE`-based reset here — `BeginRollback` already solves
  isolation for gateway tests; e2e tests should prefer unique identifiers
  over resetting shared state.
