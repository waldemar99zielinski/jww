# jww (Joint Wallet Watch) — server

Backend for Joint Wallet Watch, a personal finance Telegram bot + Mini App, written in Go with a Clean Architecture layering: plain domain structs, use-case services, adapters for Postgres and HTTP, and infrastructure packages wired together by `cmd/server`, the composition root.

> **Agents: read the `README.md` in the directory you are about to edit before writing code.**
> Every package has one, describing that layer's rules and the expected implementation shape.

## Stack

- **Router:** [chi](https://github.com/go-chi/chi) (+ `go-chi/cors`)
- **Database:** [pgx](https://github.com/jackc/pgx) (`pgxpool`), SQL built with [Squirrel](https://github.com/Masterminds/squirrel) — no ORM
- **Migrations:** [goose](https://github.com/pressly/goose), embedded in the binary and applied on server startup (the CLI still works for manual up/down/status)
- **Auth:** `golang-jwt/jwt/v5`, `golang.org/x/crypto/argon2` (Argon2id)
- **UUIDs:** `google/uuid`
- **Money/rates:** `domain.Money` / `domain.ExchangeRate`, plain `int64` fixed-point types

## Architecture — dependency direction (one way only)

```
cmd/server                              composition root — imports everything, imported by nothing
└─ internal/infrastructure/http_server  HTTP entry: New/NewRouter + Config
     ├─ kernel                          Deps, middleware, request context, respond helpers
     └─ routers/<resource>              Mount(public, protected, deps) + one handler per file
└─ internal/infrastructure/database     pgxpool + concrete UnitOfWork
└─ internal/adapter/{gateway,schema,security}
└─ internal/service/{...} + workflow    use cases; each service owns its repository port
└─ internal/domain                      plain structs + Validate(); imports nothing internal
└─ utils/{configuration,logger}
```

Hard rules:

- `domain` imports nothing internal.
- `service/<resource>` imports only `domain` + `service/unit_of_work`.
- `adapter/*` implements service ports; never imports `infrastructure/*` or `cmd`.
- `kernel` imports services/workflows/domain/logger, but **never** `routers/*` or the root `http_server` package.
- `routers/*` import `kernel`, but **never** each other, the root `http_server` package, or `cmd`.
- Root `http_server` imports `kernel` + all `routers/*`; only `cmd/server` imports it.
- `cmd/server` is the **only** place that constructs concrete adapters/repositories.

## Clean Architecture — where does code go?

**The Dependency Rule: source dependencies point inward only.** Nothing in an
inner layer may import an outer layer, directly or transitively. The use-case
layer defines the interfaces (ports) it needs; outer layers implement them
(dependency inversion).

```
Frameworks & Drivers   cmd/server, internal/infrastructure/{database,http_server}, utils/*   (outermost)
Interface Adapters     internal/adapter/*, http_server/kernel, http_server/routers
Use Cases              internal/service/*, internal/service/workflow
Entities               internal/domain                                                     (innermost)
```

| Clean Architecture layer | Directory | Put here | Never here |
|---|---|---|---|
| **Entities** (enterprise rules) | `internal/domain` | structs, `Validate()`, `Money`/`ExchangeRate`, typed errors, enums, `AppContext` | I/O, SQL, HTTP, logging, any other layer import |
| **Use Cases** (application rules) | `internal/service/<resource>`, `internal/service/workflow` | a resource's operations + its repository **port**; multi-service orchestration; transaction demarcation | HTTP/JSON, pgx, sibling-service imports (only `workflow` composes services) |
| **Interface Adapters** | `internal/adapter/{gateway,schema,security}`, `http_server/kernel`, `http_server/routers` | controllers/handlers, presenters/DTOs, repository implementations, port adapters, middleware | importing `cmd`, other adapters, or infrastructure |
| **Frameworks & Drivers** | `cmd/server`, `internal/infrastructure/{database,http_server}`, `utils/*` | composition root, DB pool, HTTP server assembly, env binder, logger | business logic |

### Decision guide — "where does my new code go?"

| You are adding… | Put it in… |
|---|---|
| A fact/invariant about an entity (amount > 0, code is 3 letters) | `internal/domain` |
| A new operation ("archive a transaction") | `internal/service/<resource>` (or `internal/service/workflow` if it needs other services) |
| A new DB query / row mapping | `internal/adapter/gateway` (implement the service's port) |
| A new request/response JSON shape | `internal/adapter/schema` |
| A new HTTP endpoint | `http_server/routers/<resource>` + register in `http_server/server.go` |
| A new env var / config | the package it configures + aggregate in `cmd/server/config.go` |
| A cross-cutting HTTP concern (auth, logging) | `http_server/kernel`, attached in `http_server/server.go` |
| Wiring a new dependency | `cmd/server/deps.go` |

**Ports live in the Use Case layer.** `Repository`, `PasswordHasher`,
`TokenService`, `Manager` are declared in `internal/service/*`; `gateway`,
`security`, and `database` implement them. A new adapter implements an existing
port — it never defines the port.

See each directory's README for detail. Top-level guides: [`internal/domain`](internal/domain/README.md), [`internal/service`](internal/service/README.md), [`internal/adapter`](internal/adapter/README.md), [`internal/infrastructure/http_server`](internal/infrastructure/http_server/README.md), [`cmd/server`](cmd/server/README.md).

## Public vs protected routes

Every resource exposes `Mount(public, protected chi.Router, deps *kernel.Deps)`.
`public` and `protected` share one chi routing tree; `protected` just has
`kernel.RequireAuth` prepended. Each resource registers each endpoint on exactly
one of the two (`public.Post(...)` vs `protected.Get(...)`).

- Register a route on **exactly one** router — chi panics at startup if the same
  method+path is registered twice.
- Authenticated requests get the resolved `domain.User` stored on the request
  context, so `kernel.RequestContext(r).User()` and the logs (`user_id`) see it.

## Design notes

- **`Transaction.ID` is a plain `int64`** (Postgres `bigint` identity). Category/User/Currency use `uuid.UUID`; `Currency` is keyed by its 3-letter `code`.
- **Money/rate fields use `domain.Money`/`domain.ExchangeRate`** — `int64` fixed-point (cents, and 1e-4 units). `+`/`-`/`<`/`==` work directly; `ExchangeRate.Convert` is the one nontrivial op. JSON encodes them as bare numbers; Postgres `NUMERIC` round-trips via text, never `float64`.
- **Date-only fields** use `schema.Date` (`"YYYY-MM-DD"`), not `time.Time`.
- **Standard JSON contract**: every 2xx body is `{"data": ...}` (`kernel.WriteJSON`/`WriteOK`/`WriteCreated`, or `kernel.WriteNoContent` for 204); every non-2xx body is `{"error": {"code", "message", "details?}}` (`kernel.WriteError`/`WriteBadRequest`). `details` carries field-level validation failures.
- **No request-validation library**: each request schema has `Validate() error`, called by hand in the handler before the service. Failures → `400 {"error": {"code": "bad_request", "message": "..."}}`.
- **Auth**: missing/malformed `Authorization` header → `401` (via `domain.AuthenticationError`) with `WWW-Authenticate: Bearer`; `/v1/auth/login` is public, everything else is protected. New users are created out of band via [`scripts/signup`](scripts/signup/README.md) or [`scripts/bootstrap`](scripts/bootstrap/README.md).
- **Config lives next to what it configures** (`http_server.Config`, `database.Config`, `security.Config`, `logger.Config`), each loaded via `utils/configuration.Load[T]()`; `cmd/server/config.go` aggregates them.
- **Error flow**: gateway returns `(nil, nil)` for absent rows → service raises a `domain.DomainError` carrying a `domain.ErrorKind` → `kernel.WriteError` maps the kind (validation→400, unauthorized→401, not found→404, conflict→409, internal/unknown→500) and writes the standard error envelope.
- **Transaction demarcation** lives in services (`Update`/`Delete`) and workflows (multi-service ops), never in handlers.
- **Logging lives at the boundary only** — `cmd/server`, `internal/infrastructure/{database,http_server}`, `scripts/*`. Never in `internal/domain`/`internal/service`/`internal/adapter`: those return errors, the boundary logs them once with full request context. Logging must also stay independent of response/return-value assembly — `http_server/kernel/respond.go`'s `Write*` helpers never gain a logging side effect. See [`utils/logger`](utils/logger/README.md).

## Running locally

```bash
cp .env.example .env   # fill in DATABASE_URL (postgres:// DSN) and JWT_SECRET_KEY
go run ./cmd/server
```

### Hot reload

With [Air](https://github.com/air-verse/air) installed (`go install github.com/air-verse/air@latest`), run:

```bash
air
```

`server/.air.toml` rebuilds on `.go` changes, rebuilds on `.sql` changes (migrations are embedded at build time), and restarts on `.env` edits. Air sends `SIGINT` on rebuild so the server shuts down gracefully.

## Running migrations

Migrations are embedded in the binary (`migrations/embed.go`) and applied automatically on startup:
`run()` calls `database.Migrate(ctx, pool)` right after opening the pool, so `go run ./cmd/server`
brings the schema up to date before the first request is served. Versions are tracked in the same
`goose_db_version` table the CLI uses, so both paths stay in sync.

For manual control (status, rollback, creating a new file) use the goose CLI:

```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
export GOOSE_DRIVER=postgres
export GOOSE_DBSTRING="$DATABASE_URL"
goose -dir migrations up
goose -dir migrations status
goose -dir migrations down   # undo the most recent migration
```

## Tests

Full strategy, conventions, and per-layer rationale: **[`TESTING.md`](TESTING.md)**.
Every package's own README has a "Testing requirement for this directory"
section with the bounded rule for that package specifically.

```bash
go test ./...                         # fast tests only — no Docker needed
go test -tags=integration ./...       # + gateway tests, needs Docker running
go test -tags=e2e ./...               # + full-stack journeys + boot smoke test, needs Docker
go test -tags="integration e2e" ./... # everything
go test -race ./...                   # fast tests, race detector
```

- **`domain`/`service`/`adapter/{security,schema}`:** fast, no I/O. Services
  fake their own minimal `Repository` port by hand (no mocking library) —
  see [`internal/service`](internal/service/README.md#testing-requirement-for-this-directory).
- **`adapter/gateway`:** integration-tagged, real Postgres via
  [`internal/tests/pgtest`](internal/tests/pgtest/README.md)
  (testcontainers-go), never a mocked `DBTX` — a fake would only verify the
  mock, not the query. Isolated per-test via a rolled-back transaction, not
  `TRUNCATE`.
- **`http_server/kernel` and `routers/<resource>`:** `httptest`-based, real
  services over local hand-written fake repositories — no real Postgres.
- **`tests/e2e/`:** full-stack HTTP journeys (real Postgres, real adapters, no
  fakes) plus one test that boots the actual `cmd/server` binary. A handful
  of critical flows, not an endpoint-by-endpoint enumeration — see
  `TESTING.md`'s note on how this relates to `client/web`'s Playwright E2E.
