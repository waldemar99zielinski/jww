# internal/tests

**Role.** Parent directory for test-support packages scoped to testing
*internal* code — currently just [`pgtest`](pgtest/README.md), the
testcontainers-backed Postgres helper `internal/adapter/gateway`'s
integration tests (and `server/tests/e2e`) depend on. Compare
`server/tests/`, which holds whole-system suites that test the composed
server rather than one internal package.

## Contents

- [`pgtest/`](pgtest/README.md) — boots a disposable Postgres container,
  applies migrations, and provides per-test transaction-rollback isolation.

## Rules

- Nothing here ships in a production binary; everything is consumed only
  by `_test.go` files (directly, or via a tagged non-test package like
  `pgtest` itself — see its own README for why it isn't a `_test.go` file).
- A new internal test-support package (scoped to one concern, like `pgtest`
  is to "a real Postgres") gets its own subdirectory here.

## Testing requirement for this directory

No tests of its own — a parent directory with no code. See `pgtest/`'s own
README, and `server/TESTING.md`.
