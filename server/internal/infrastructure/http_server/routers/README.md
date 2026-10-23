# internal/infrastructure/http_server/routers

**Clean Architecture layer: Interface Adapters.** Handlers are "controllers" that translate HTTP requests into use-case calls.

**Role.** One subpackage per resource, holding its HTTP handlers and its `Mount`. Routers depend on `kernel` (and `adapter/schema`, `domain`, `logger`); they never import each other, the root `http_server` package, `cmd`, or concrete adapters. Only the root `http_server` package imports routers.

## Subpackages

`auth/`, `currency/`, `currency_exchange_rate/`, `health/`, `summary/`, `transaction/`,
`transaction_attachment/`, `transaction_category/`, `user/`, `webauthn/`.

## A router's shape

- `router.go`:
  ```go
  func Mount(public, protected chi.Router, deps *kernel.Deps) {
      public.Post("/v1/<resource>", createHandler(deps))   // unauthenticated
      protected.Get("/v1/<resource>", listHandler(deps))   // authenticated
      // ... register each route on exactly ONE router
  }
  ```
- One file per handler: a factory `func xHandler(deps *kernel.Deps) http.HandlerFunc`.
- Routers carry no logger of their own — `kernel`'s `RequestLogger` middleware already logs every request (`"request completed"`, method/path/status/duration); see `utils/logger`'s "Where to log".

## Handler body order

1. `appCtx := kernel.RequestContext(r)`
2. Decode + `Validate()` (body) **or** parse path/query params.
3. Call a `kernel.Deps` service/workflow with `appCtx`.
4. `kernel.WriteError(w, err)` on failure; on success use `kernel.WriteJSON`/`WriteOK`/`WriteCreated` (all wrap in `{"data": ...}`) or `kernel.WriteNoContent(w)` — never a bare `w.WriteHeader`.

## Rules

- **Register each method+path on exactly one of `public`/`protected`** — chi panics on duplicates. Don't use `Route`/`Mount` to put the same prefix on both routers.
- Use `kernel.RequestContext(r)` (never a bare `context`) so logs/services carry the flow id and user.
- URL params via `chi.URLParam`; transactions use `int64` ids (`strconv.ParseInt`), everything else `uuid.Parse`; dates via `schema.ParseDate`.
- Handler-local sentinel errors are fine; promote to `kernel/query_errors.go` only when shared by ≥2 resources.

## Avoid

- Importing another router package, the root `http_server` package, or any concrete adapter.
- Building services/repos per request (they come from `kernel.Deps`).
- Writing a body after `w.WriteHeader`.

## Add a new resource

1. Add `routers/<resource>/router.go` + one file per handler.
2. Add its service/workflow to `kernel.Deps` and `cmd/server/deps.go`.
3. Register `Mount` in the root `http_server/server.go`.
4. Add request/response shapes in `internal/adapter/schema`.
5. Extend `routing_verify_test.go` if auth expectations change.
6. Add a `<resource>_handlers_test.go` — see each subpackage's own "Testing
   requirement" section and `server/TESTING.md`.

## Testing requirement for this directory

Every resource subpackage needs `httptest`-based tests (`package
<resource>_test`) covering each handler's happy path and its specific error
paths, built from real services wired to local hand-written fake
repositories — no real Postgres. See each subpackage's own README and
`server/TESTING.md`.
