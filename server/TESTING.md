# Testing strategy

**Implemented 2026-10-15.** Every layer below has real tests following this
plan; `go test ./...` / `-tags=integration` / `-tags=e2e` all pass. Each
package's own README has a "Testing requirement for this directory" section
with that package's bounded rule — check there first when adding or
changing a test. This document is now the reference for *why* the strategy
looks like this, not a pending proposal.

## Decisions already made (2026-10-14)

| Decision | Choice | Why |
|---|---|---|
| Integration DB | **testcontainers-go** | Ephemeral `postgres:17-alpine` per test run via the official Go module API, matching the image already used in `development/docker-compose.yml`. No manual setup, hermetic, same behavior on any machine. |
| Assertions | **testify** (`assert`/`require`) | Less boilerplate than hand-rolled `if got != want { t.Fatalf(...) }`. We keep hand-writing fakes — **not** `testify/mock` — so the existing "ports stay scoped, fakes stay local" philosophy in [`internal/service/README.md`](internal/service/README.md#testing) is unchanged. |
| Fast/slow split | **Build tags** | `//go:build integration` for gateway tests, `//go:build e2e` for the full-stack suite — separate tags since they're independently expensive (one Postgres-backed test file vs. booting a whole HTTP server). `go test ./...` runs only fast tests by default. |
| E2E boot strategy | **Both** | In-process `httptest.NewServer` wrapping the real router for journey tests (fast, everything below HTTP is real) + one separate smoke test that spawns the compiled `cmd/server` binary to prove `main.go`'s own wiring (startup migrations, graceful shutdown). |
| CI | **Not in this plan** | No git remote exists yet. This doc defines local commands and the layering strategy only; CI is a follow-up once the repo is pushed somewhere. |

## Philosophy

- **Test pyramid, shaped by the existing Clean Architecture layering.** Most
  tests are fast and need no I/O (domain, service, security, schema). A
  smaller set needs a real Postgres (gateway). Fewer still exercise a full
  HTTP round-trip (handlers). We don't add a layer of tests for something
  already fully covered by the layer below it.
- **Don't mock what you don't own.** We never mock `pgx`, `net/http`, or
  third-party libs. Gateway tests run against a real Postgres; everything
  above the gateway is tested through the port it depends on, with a
  hand-written fake — this already matches `internal/service/README.md`'s
  "no mocking library" rule, just not implemented yet.
- **Fakes stay local to the package that needs them**, same reasoning
  `internal/service/README.md` already gives: a shared fake interface would
  force every test to implement methods it doesn't use. No central
  `mocks`/`testsupport` package full of shared fakes.
- **Test at the boundary that would actually catch the bug.** A validation
  rule belongs in a domain test, not a gateway test. A SQL column-mapping bug
  belongs in a gateway test, not a service test with a fake repo (the fake
  can't catch it).

## The layers

| Layer | Needs real I/O? | Test type | Package style |
|---|---|---|---|
| `internal/domain` | No | Unit | white-box (`package domain`) |
| `internal/service/*` + `workflow` | No — fake the port | Unit | black-box (`package x_test`) |
| `internal/adapter/security` | No | Unit | black-box |
| `internal/adapter/schema` | No | Unit | black-box |
| `internal/adapter/gateway` | **Yes — real Postgres** | Integration | black-box |
| `internal/adapter/storage` | **Yes — real MinIO** | Integration | black-box |
| `internal/infrastructure/http_server/kernel` | No (mostly) | Unit | white-box (already the case) |
| `internal/infrastructure/http_server/routers/<resource>` | No — fake repos under real services | Unit-ish (httptest) | black-box |
| `internal/infrastructure/database` (migrate, UoW) | Yes | Integration | — (exercised by gateway tests) |
| `server/tests/e2e` (whole system) | **Yes — real Postgres, real HTTP** | E2E | `package e2e_test` |
| `scripts/*` | Yes | Later / deferred | — |

"White-box" = internal test package (`package domain`), same as the 5 tests
that already exist. "Black-box" = `_test` package suffix (`package domain_test`
would be wrong here, meant generically — e.g. `package currency_test`),
testing only the public API. I'm proposing black-box for everything new
*except* `kernel`, which already uses white-box to reach unexported helpers
(`bearerToken`, `ensureFlowID`) and should stay that way.

### `internal/domain` — unit, no I/O

Already started (`money_test.go`). Extend with table-driven `Validate()` tests
per entity (one subtest per violated invariant, one happy path), enum
`Valid()` tests, and `ErrorKind` mapping sanity checks. Nothing here needs
testcontainers or fakes — it's the cheapest, highest-value layer to fill in
first.

```go
func TestCurrency_Validate(t *testing.T) {
    tests := []struct {
        name    string
        c       domain.Currency
        wantErr string // substring, "" = no error
    }{
        {"valid", domain.Currency{Code: "USD", Name: "US Dollar"}, ""},
        {"lowercase code rejected", domain.Currency{Code: "usd", Name: "US Dollar"}, "code"},
        // ...
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) { /* ... */ })
    }
}
```

### `internal/service/*` + `workflow` — unit, hand-written fakes

Each resource test builds a minimal fake implementing that package's own
`Repository` port (and `PasswordHasher`/`TokenService` where relevant),
exactly as `internal/service/README.md#testing` already describes. Cover, per
service:

- Happy path (Create/Get/List/Update/Delete).
- Validation failure → repo is never called (assert the fake's call count).
- Repo `(nil, nil)` → translated into the right `domain.New...NotFoundError`.
- `Update`/`Delete`: a fake `unit_of_work.Manager`/`UnitOfWork` that records
  calls, asserting `Restore` fires on every error path and `Finish` fires
  exactly once on success.
- Defaults get set (`Active: true`, a non-zero `CreatedAt`, etc.) — assert
  "set", not an exact clock value, since services call `time.Now().UTC()`
  directly (no injected clock).

**`workflow/` is a wrinkle:** it composes *concrete* service structs
(`*currency.Service`, not an interface), so a workflow test can't swap in a
fake service the way a handler test could. The practical pattern: build the
*real* concrete services, each wired to its own fake repository one level
down, and let the workflow orchestrate those real-but-in-memory services.
Still zero I/O, still fast — just one layer less "fake" than a pure service
test.

### `internal/adapter/security` — unit, no I/O

`Argon2PasswordHasher`: hash/verify round-trip, wrong password rejected, hash
format sanity. `JwtTokenService`: sign/verify round-trip, expired token
rejected, wrong secret rejected, `LoadConfig`'s "bad `JWT_ALGORITHM` fails at
startup" behavior (already fixed in code, needs a test). All fast, no
containers.

### `internal/adapter/schema` — unit, no I/O

Table-driven `Validate()` tests per request DTO (mirror the domain pattern),
plus `<Entity>ToResponse` round-trip tests and `schema.Date` parse/format edge
cases (`ParseDate("")`, malformed input, leap years if relevant).

### `internal/adapter/gateway` — **integration**, real Postgres

This is the one layer that gets a real database, via testcontainers-go. Two
new pieces of test infrastructure:

1. **A small internal test-support helper** (not exported, not a dependency
   for production code) — something like
   `internal/adapter/gateway/gateway_testing_test.go` or a tiny
   `internal/tests/pgtest` package — that, once per `go test` binary run
   (`TestMain`), starts a `postgres:17-alpine` container via
   `testcontainers-go/modules/postgres`, runs `database.Migrate` against it,
   and hands back a `*pgxpool.Pool` for every test in the package to share.
   One container per test *binary*, not per test — starting a container
   takes real time; we don't want 50 of them.

2. **Per-test isolation via the existing `UnitOfWork` abstraction** — no
   `TRUNCATE` needed. Every gateway method already accepts an optional
   `uow ...unit_of_work.UnitOfWork`
   ([`internal/service/unit_of_work`](internal/service/unit_of_work/README.md)).
   So each test:

   ```go
   uow, err := uowManager.Create(appCtx)
   require.NoError(t, err)
   t.Cleanup(func() { _ = uow.Restore(appCtx) }) // always rolls back

   repo := gateway.NewPostgresCurrencyRepository(pool)
   created, err := repo.Create(appCtx, currency, uow)
   // ... assert against created / repo.GetByCode(appCtx, code, uow) ...
   ```

   Every write the test makes lives inside one Postgres transaction that gets
   rolled back in `t.Cleanup`, regardless of pass/fail. Tests can run in
   parallel against the one shared container with zero cross-test
   interference, and we don't have to reason about `TRUNCATE` ordering
   against FK constraints (`transactions → transaction_categories`,
   `transactions → currencies`, etc.). This reuses infrastructure the
   codebase already has instead of adding a new reset mechanism.

Cover, per gateway: the actual SQL round-trips a fake can't — `NUMERIC` ↔
`domain.Money`/`ExchangeRate` precision, `(nil, nil)` on a genuine miss,
unique/FK constraint violations surfacing as the right error, Squirrel query
correctness for every filter combination `transaction`'s `List` supports.

Build tag at the top of every file in this set:

```go
//go:build integration
```

### `internal/infrastructure/http_server/kernel` — mostly unit, no I/O

`RequestLogger`/`Recoverer`/`LimitRequestBody` are pure `http.Handler`
middleware — test with `httptest.NewRequest`/`httptest.NewRecorder` and no
services at all. `RequireAuth` is the one piece that needs a real
`*auth.Service` (concrete struct, same issue as workflows) — build one wired
to a fake `user.Repository` and a real `security.JwtTokenService` with a
throwaway secret, then assert missing/malformed/expired/valid tokens all
behave correctly. `bearerToken`/`ensureFlowID`/`RequestContext` are pure
functions already easy to unit test directly (white-box, same package).

### `internal/infrastructure/http_server/routers/<resource>` — httptest, fakes underneath

This fills the gap the last audit flagged (no handler-level tests existed).
Build `kernel.Deps` from **real services wired to fake repositories** — the
exact same fakes the service-layer tests already wrote, reused one level up.
No real Postgres, no testcontainers, still fast:

```go
deps := &kernel.Deps{
    CurrencyService: currency.NewService(fakeCurrencyRepo, fakeUoWManager),
    // ...
}
req := httptest.NewRequest(http.MethodPost, "/api/v1/currencies", strings.NewReader(body))
rec := httptest.NewRecorder()
currency.Mount(router, router, deps) // or call the handler directly
router.ServeHTTP(rec, req)

assert.Equal(t, http.StatusCreated, rec.Code)
```

One test file per **resource** (not per handler file) covering every
handler's happy path plus the error paths that are resource-specific (bad
JSON → 400, failed `Validate()` → 400, not-found → 404, conflict → 409 where
applicable). Generic error-mapping behavior (`WriteError`'s kind→status
table) is already covered once in `kernel/respond_test.go` — resource tests
don't need to re-prove that.

### `scripts/bootstrap`, `scripts/signup` — later / optional

`cmd/server`'s own wiring is now covered by the E2E boot smoke test above.
`scripts/bootstrap`'s idempotency and `scripts/signup`'s CLI flags are thin
wrappers around already-tested services and only make sense as integration
tests themselves — lower value for now. I'd defer these to a later pass
rather than include them in the first rollout; flag if you disagree.

### End-to-end — real Postgres, real adapters, no fakes anywhere

**How this relates to `client/web`'s E2E:** that suite already uses Playwright
against the real backend (`development/docker-compose.yml`) for a handful of
critical *UI* journeys (login/protected routes, add-transaction round trip,
category CRUD) — see `client/web/TESTING.md`. Server E2E is a different,
narrower concern: proving the API's own multi-endpoint contract is correct
*without* a browser or the web client, so a backend-only change can be
verified on its own and the suite stays runnable with just `go test`. The two
don't duplicate each other — Playwright proves the UI and API work together
through a browser; this proves the HTTP API is internally correct. Neither
suite should grow to cover what the other already does.

**Two tiers, both using the real Postgres from the testcontainers helper
(§ gateway, above) — nothing is faked:**

1. **Journey tests (primary, in-process).** `httptest.NewServer` wrapping
   `http_server.NewRouter(deps)`, where `deps` is built exactly like
   `cmd/server/deps.go` does — real gateways, real `security` adapters, the
   real `PostgresUnitOfWorkManager` — just pointed at the testcontainers
   pool instead of a production `DATABASE_URL`. Fast (no process spawning),
   and everything below the HTTP layer is 100% real: a bug in Squirrel SQL,
   JWT signing, or service orchestration would show up here exactly like in
   production. What it does *not* exercise is `cmd/server`'s own `main()` —
   config loading, `database.Migrate` running at startup, signal handling,
   graceful shutdown.

2. **One boot smoke test (secondary, out-of-process).** Build the actual
   `cmd/server` binary (`go build -o`), run it as a subprocess pointed at the
   test Postgres DSN and an ephemeral port, wait for it to report healthy,
   hit one endpoint, then send it `SIGTERM` and assert it exits cleanly
   within the shutdown timeout. This is the one place that actually proves
   `main.go`'s wiring — including that migrations really do run on startup
   against a schema-less database — works as deployed. Kept to a single test,
   not a journey suite; the journeys above already cover business behavior.

**Suggested journeys for tier 1** (mirrors Playwright's "only a few critical
flows" philosophy — this is deliberately not an endpoint-by-endpoint
enumeration; that's what gateway and handler tests already do):

- **Auth:** log in as a bootstrapped user → `GET /v1/auth/me` succeeds →
  the same request with no/garbage token → `401` with
  `WWW-Authenticate: Bearer`.
- **CRUD + relationships:** create a currency → create a category → create a
  transaction referencing both → fetch it back → list with a filter that
  should match it → delete it → re-fetch → `404`.
- **Summary (the riskiest read path):** create a couple of transactions in
  different currencies → `GET /v1/summary` → assert the aggregated,
  currency-converted numbers are right. This is the one endpoint that
  touches currency, exchange rate, transaction, and user services together,
  so it's the highest-value single journey to get right end-to-end.

**Data isolation is different here than in gateway tests.** Gateway tests
roll back a transaction per test (§ gateway, above); E2E can't do that —
each HTTP request commits for real inside the service it calls, the test is
just a client. Instead: give the whole E2E suite its own container (torn
down after the binary exits, so leftover rows don't outlive the run), have
each journey use unique identifiers (a per-test currency code / username
prefix) rather than asserting exact list counts, and don't run journeys with
`t.Parallel()` against shared seed data (e.g. the seeded transaction
categories) without accounting for what other journeys may have created.

**Location:** `server/tests/e2e/` (`package e2e_test`) — it tests the
composed system, not any one internal package, so it doesn't belong under
`internal/`. Generic helpers it shares with any future suite under
`tests/` (an HTTP test client, `BuildTestDeps`, `FreePort`, `ModuleRoot`)
live in the sibling `server/tests/utils/` package, importable since it's a
plain package rather than a `_test.go` file. Tagged `//go:build e2e` (not
`integration`), so
`go test -tags=integration ./...` (gateway only) and
`go test -tags=e2e ./...` (full-stack journeys) can be run independently —
the journey suite is slower and broader than a single gateway test file, and
someone iterating on one gateway shouldn't have to also boot a full HTTP
server.

## Conventions

- **File location:** black-box tests (`package x_test`) live in a `tests/`
  subdirectory of the package they test, e.g.
  `internal/adapter/gateway/tests/postgres_currency_repository_test.go`
  tests `internal/adapter/gateway`. This is a plain Go package one level
  down that imports the parent package's exported API — nothing special to
  `go test`, which still finds it via `./...`. **Exception:** white-box
  tests (`package domain`, `package kernel`, `package configuration`,
  `package http_server`, `package pgtest`) reach unexported identifiers
  (`isUpperAlpha`, `levelForStatus`, `bearerToken`, …) and Go does not allow
  a subdirectory to see a parent package's unexported names — those test
  files stay alongside the code they test. Don't convert a white-box test to
  black-box just to move it; only do so if the test can honestly be rewritten
  against the public API.
- **Naming:** `Test<Type>_<Method>_<Scenario>`, e.g.
  `TestService_Create_ValidatesBeforeCallingRepo`. Table-driven with
  `t.Run(tt.name, ...)` subtests, matching the existing `money_test.go`/
  `respond_test.go` style.
- **testify scope:** `require` for setup/preconditions that should abort the
  test immediately (`require.NoError(t, err)`); `assert` for the actual
  checks so one test can report every failing assertion instead of stopping
  at the first. No `testify/suite`, no `testify/mock`.
- **Fakes:** one hand-written fake per port, declared in the `_test.go` file
  that needs it (or a `fakes_test.go` in the same package if more than one
  test file shares it) — never a package-wide shared fakes file across
  services.
- **Build tag placement:** `//go:build integration` as the first line,
  blank line, then `package foo_test`. Gateway (and later, any other
  DB-backed) test files only.

## How to run

```bash
go test ./...                        # fast tests only (default, no Docker needed)
go test -tags=integration ./...      # + gateway tests, needs Docker running
go test -tags=e2e ./...              # + full-stack journeys and the boot smoke test, needs Docker
go test -tags="integration e2e" ./... # everything
go test -race ./...                  # fast tests, race detector
```

## Phased rollout (proposed order)

1. **`internal/domain`** — fill in `Validate()` coverage for every entity.
2. **`internal/adapter/security`** + **`internal/adapter/schema`** — fast,
   no new infra, closes two audit-flagged gaps (JWT algorithm fail-fast,
   general schema validation) in one pass.
3. **`internal/service/*`** — the highest-value layer; this is where most
   business-rule regressions would actually be caught.
4. **`internal/service/workflow`** — once the services above have fakes to
   compose.
5. **`internal/adapter/gateway`** — stand up the testcontainers helper once,
   then one test file per repository.
6. **`http_server/kernel`** — middleware unit tests.
7. **`http_server/routers/<resource>`** — handler tests, reusing service-test
   fakes.
8. **`server/tests/e2e`** — the testcontainers-backed journey suite (auth,
   CRUD+relationships, summary) plus the single `cmd/server` boot smoke test.
9. *(deferred)* `scripts/bootstrap`/`scripts/signup` tests, CI.

## What actually happened, vs. the plan above

- **Phased order:** followed roughly as listed, implemented in parallel
  rather than strictly sequentially.
- **Workflow and `RequireAuth`:** the "real service/concrete struct over a
  local fake repo" pattern worked cleanly in practice — no awkwardness,
  including for `TransactionWorkflow`'s multi-service composition and
  `SummaryWorkflow`'s cross-service aggregation. `RequireAuth` turned out
  simpler than expected: `auth.Service` is built from its own small ports
  (`UserRepository`/`PasswordHasher`/`TokenService`), not a concrete
  `kernel.Deps`-style struct, so it got a normal fake-the-port test rather
  than needing the one-level-down pattern.
- **`scripts/bootstrap`/`scripts/signup`:** deferred, as planned.
- **Coverage target:** none set; no numeric gate, matching "every layer has
  the tests described above."
- **Real bugs found by writing these tests** (fixed, not just noted): a
  gateway test reusing one Postgres transaction across a deliberately-failed
  constraint check and a must-succeed insert (the first poisons the
  transaction — split into two tests); a list-count assertion that didn't
  account for migration-seeded reference rows; a `time.Time` struct
  comparison that failed on `*time.Location` internals despite the instants
  being equal; and an E2E test asserting the wrong sign for `Summary.Net` on
  an all-expense scenario (`Net = income − expense` is enforced by
  `domain.Summary.Validate()` itself — the test was wrong, not the code).
