# tests/e2e

**Role.** The server's own end-to-end suite — real Postgres, real adapters,
real HTTP, no fakes anywhere. Tests the composed system, not any one
internal package, so it lives under `server/tests/` rather than under
`internal/`. See [`server/TESTING.md`](../../TESTING.md)'s "End-to-end"
section for the full rationale, and [`tests/utils`](../utils/README.md) for
the generic helpers this suite (and any future one under `tests/`) shares.

**How this relates to `client/web`'s E2E:** that suite uses Playwright
against the real backend for a handful of critical *UI* journeys
(login/protected routes, add-transaction round trip, category CRUD) — see
`client/web/TESTING.md`. This suite is a different, narrower concern:
proving the API's own multi-endpoint contract is correct *without* a
browser or the web client. The two must not grow to cover the same ground —
Playwright proves the UI and API work together through a browser; this
proves the HTTP API is internally correct on its own.

## Two tiers

1. **Journey tests (primary, in-process)** — `auth_journey_test.go`,
   `crud_journey_test.go`, `summary_journey_test.go`,
   `attachment_journey_test.go`. `TestMain` (`e2e_main_test.go`) starts one
   `pgtest` Postgres container **and one `s3test` MinIO container** for the
   whole package, builds real `kernel.Deps` (`utils.BuildTestDeps`, duplicating
   `cmd/server/deps.go`'s `buildDeps` — see that package's README for why),
   wraps `http_server.NewRouter(deps)` in `httptest.NewServer`, and seeds one
   user directly through the service layer (there's no public signup HTTP
   endpoint). Everything below the HTTP layer is 100% real; a bug in
   Squirrel SQL, JWT signing, S3 streaming, or service orchestration shows up
   here exactly like in production. What it doesn't exercise is `cmd/server`'s
   own `main()`.
2. **Boot smoke test (secondary, out-of-process)** — `boot_smoke_test.go`'s
   `TestCmdServerBoots`. Builds the actual `cmd/server` binary, runs it as a
   subprocess against the same `pgtest` database, the same MinIO instance, and
   an ephemeral port (`utils.FreePort`), polls until it responds, sends
   `SIGTERM`, and asserts it exits cleanly. This is the one place that proves
   `main.go`'s own wiring — config loading, `database.Migrate` running at
   startup, storage-adapter construction, graceful shutdown — works as
   deployed. Kept to exactly one test; the journeys above already cover
   business behavior.

## Data isolation

Unlike gateway tests (one rolled-back transaction per test via
`pgtest.BeginRollback`), journeys here can't share a transaction with the
server — each HTTP request commits for real, inside whatever transaction
the service itself opens. Instead: the whole package gets one disposable
container (torn down after the test binary exits, so nothing outlives the
run), and each journey uses a unique currency code (`ZZA`/`ZZB`/…) rather
than asserting exact list counts, since migration `00006` seeds 20
reference categories and other journeys may run in the same container.

## Testing requirement for this directory

Keep this suite small — 2-4 journeys covering the riskiest or most
representative flows (auth, a CRUD+relationship round trip, the summary
aggregation), not an endpoint-by-endpoint enumeration. That level of detail
belongs in gateway integration tests and router `httptest` tests, which are
far cheaper to run and to extend. Add a journey here only when a flow
spans multiple resources in a way no single lower-level test suite would
catch. The one `TestCmdServerBoots` test is a fixed cost, not something to
duplicate — don't add a second binary-spawning test for a different
scenario; extend the existing one if `main.go`'s own behavior needs more
coverage. Generic (non-journey-specific) helpers belong in `tests/utils`,
not here — see that package's own testing requirement.

## How to run

```bash
go test -tags=e2e ./tests/e2e/...   # needs Docker running
```
