# tests

**Role.** Parent directory for top-level, whole-system test suites —
things that test the composed server, not any one internal package, so
they don't belong under `internal/`. Compare `internal/tests/`, which
holds test-support packages scoped to internal integration tests (e.g.
`pgtest`).

## Contents

- [`e2e/`](e2e/README.md) — the server's own end-to-end suite (real
  Postgres, real adapters, real HTTP, plus one `cmd/server` boot smoke
  test).
- [`utils/`](utils/README.md) — generic helpers shared across suites here
  (HTTP test client, test-deps wiring, OS-level utilities).

## Rules

- A new whole-system suite gets its own subdirectory here, not a new
  top-level directory next to `cmd`/`internal`/`migrations`.
- Helpers reused by more than one suite under here live in `utils/`;
  helpers specific to one suite stay in that suite's own package.
- Nothing under `tests/` is imported by production code (`cmd/`,
  `internal/`) — the dependency only ever points the other way.

## Testing requirement for this directory

No tests of its own — a parent directory with no code. See `e2e/` and
`utils/`'s own READMEs, and `server/TESTING.md`.
