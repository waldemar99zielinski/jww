# internal/infrastructure/http_server/kernel

**Clean Architecture layer: Interface Adapters.** The HTTP "controller" layer's shared plumbing (ports/DTOs/middleware); sits between handlers and use cases.

**Role.** The HTTP-layer kernel every resource subpackage depends on: the `Deps` singleton, middleware (`RequestLogger`, `Recoverer`, `RequireAuth`), request-context/flow-id helpers, JSON decode/respond helpers, and shared query errors. It deliberately knows nothing about any specific resource.

> `kernel` may import services/workflows/`domain`/`utils/logger`, but **never** `routers/*`, the root `http_server` package, or `cmd` (`routers → kernel`; importing a router here is a cycle).

## Files

- `deps.go` (package doc) — `httpLog = logger.New().WithComponent("http_server")`; `Deps` holds process-lifetime singletons (`AuthService`, `UserService`, `SignupWorkflow`, `CurrencyService`, `CurrencyExchangeRateWorkflow`, `TransactionCategoryService`, `TransactionService`, `TransactionAttachmentService`, `TransactionWorkflow`, `SummaryWorkflow`, `CorsOrigins`, `MaxBodyBytes`).
- `middleware.go`
  - `RequestLogger` — resolves the flow id, echoes `X-JWW-Flow-ID`, wraps the writer, and after serving logs `"request completed"` with `method`/`path`/`status`/`response_bytes`/`duration`; level from `levelForStatus` (5xx→error, 4xx→warn, else info).
  - `Recoverer` — logs `"panic recovered"` (with stack) via the shared logger and writes the standard `500 {"error":{"code":"internal_error","message":"internal server error"}}`.
  - `LimitRequestBody(maxBytes)` — wraps `r.Body` in `http.MaxBytesReader`; `maxBytes <= 0` falls back to `defaultMaxBodyBytes` (1 MiB), so even a test-built `kernel.Deps{}` gets a cap. An oversized body fails on read, which `DecodeJSON` surfaces as a 400. **The two attachment-upload routes** (`POST /api/v1/transaction-attachments` and `POST /api/v1/transactions/{transactionID}/attachments`, per `isAttachmentUploadRoute`) are exempt — multipart is never exempted globally. Route-pattern matching is delegated to `http_server/utils.MatchPath`.
  - `bearerToken(r) (string, bool)` — parses `Authorization: Bearer <token>`.
  - `RequireAuth(deps)` — missing/malformed header → `WriteError(domain.NewAuthenticationError())`; calls `deps.AuthService.Authenticate(RequestContext(r), token)`; stores the user via `withUser` on success.
- `request_context.go` — headers `X-JWW-Flow-ID` (`headerPrefix = "X-JWW"`); context keys/helpers `withUser`/`userFromContext`, `withFlowID`/`flowIDFromContext`; `newFlowID()` (UUIDv4), `ensureFlowID(r)` (context → header → new); `RequestContext(r) domain.AppContext` builds the per-request `AppContext` (context, HTTP source, user if any, flow id).
- `respond.go` — `DecodeJSON[T]` (disallow unknown fields, exactly one value); success helpers `WriteJSON`/`WriteOK`/`WriteCreated` (all wrap the payload in `{"data": ...}`) and `WriteNoContent`; error helpers `WriteBadRequest`/`WriteValidationError` (400, the latter with field-level `details`) and `WriteError` — the single error→HTTP mapper: `domain.DomainError.Kind()` selects the status and code (`validation`→400, `unauthorized`→401 + `WWW-Authenticate: Bearer`, `not found`→404, `conflict`→409, `internal`/unknown→500, `unavailable`→503 `service_unavailable`), emitted as `{"error":{"code","message","details?"}}`.
- `multipart.go` — `ReadSingleMultipartFile(r)` streams a multipart body via `r.MultipartReader()` (never `ParseMultipartForm`/`ReadForm`): exactly one file part named `file`, a normalized/capped filename (path components and control chars stripped), a media type defaulting to `application/octet-stream`. The returned `*MultipartFile` is an `io.Reader`; call `Finish()` after reading to drain the part and prove no extra part followed.
- `download.go` — `WriteDownload(w, filename, mediaType, byteSize, body)`: the one documented exception to the success envelope (bytes on success), setting `Content-Type`, `Content-Length`, an RFC 6266 `Content-Disposition`, `Cache-Control: private, no-store`, and `X-Content-Type-Options: nosniff`. `attachmentDisposition` builds the ASCII fallback + UTF-8 form.
- `query_errors.go` — `ErrInvalidType`, `ErrInvalidSource` (shared by ≥2 resources; `ErrInvalidType` validates both `TransactionCategoryType` and `TransactionType` query params since the two enums share the same string values).

## Rules

- **Always build context via `RequestContext(r)`** so services and logs share `flow_id`/`http_endpoint`/`user_id`.
- **Error mapping is centralized** in `WriteError`; handlers never switch on error types.
- **Handlers don't log their own invocation.** `RequestLogger` already logs `"request completed"` (method/path/status/duration) for every request; for now that's the only per-request log line, so `routers/<resource>` packages carry no `var log` and no component label of their own.

## How to extend

- **New middleware:** `func Name(next http.Handler) http.Handler`, log via `httpLog.WithContext(RequestContext(r))`, register it in the root `server.go` before routes. Auth-style middleware takes `*Deps` and is applied per-group with `api.With(...)`.
- **New shared HTTP behavior / query error:** add to the relevant kernel file / `query_errors.go`; keep it resource-agnostic and only promote query errors used by ≥2 resources. Put small dependency-free value helpers in [`../utils`](../utils/README.md), not in `kernel`.
- **New dependency:** add a field to `Deps` and populate it in `cmd/server/deps.go`.

## Avoid

- Resource-specific logic or routes here.
- Raw `slog`/`fmt` logging (use `httpLog`/a child logger so flow ids attach).
- Reading/writing context keys from outside the kernel.
- Regenerating flow ids outside `ensureFlowID`.
- Adding logging calls to `respond.go`'s `Write*`/`writeError` family — they're response assemblers and must stay pure (see `utils/logger`'s "Logging is independent of response/return-value assembly"). `WriteError`'s one existing internal-error log predates this rule; don't extend the pattern.

## Testing requirement for this directory

**Location:** alongside the source, not a `tests/` subdirectory — these are
white-box (`package kernel`) and reach unexported helpers like
`bearerToken`/`ensureFlowID`/`levelForStatus`, which a subdirectory package
cannot see. See `server/TESTING.md#conventions`.

White-box (`package kernel`), `httptest` + testify — see `server/TESTING.md`.
No real database or HTTP server needed.

- Every middleware (`RequestLogger`, `Recoverer`, `LimitRequestBody`,
  `RequireAuth`) gets its own `httptest`-based test asserting observable
  HTTP behavior (headers, status, body) — not internal log output. Testing
  "the right log level was picked" means unit-testing `levelForStatus`
  directly (a pure function), not capturing stdout: the shared
  `utils/logger` base handler is process-global, and swapping `os.Stdout`
  after it's been constructed doesn't redirect it, so that approach silently
  tests nothing. See `TestLevelForStatus` / `TestRequestLogger_*` in
  `middleware_test.go`.
- `RequireAuth` needs a real `*auth.Service` (same "concrete struct, not a
  port" situation as `workflow` — see `TESTING.md`): wire it to a
  hand-written fake `auth.UserRepository` (package-local, deliberately
  narrower than `service/user.Repository` — see
  `internal/service/auth/auth.go`), plus the **real**
  `security.Argon2PasswordHasher` and `security.NewJwtTokenService` with a
  throwaway secret. Both are pure in-memory crypto (no real I/O), so the
  test stays fast while still proving `RequireAuth` against genuine JWT
  signature/expiry checks instead of a hand-rolled stand-in for them.
- `bearerToken`, `ensureFlowID`, `RequestContext` are pure/near-pure
  functions — test them directly, table-driven where it fits.
- New middleware added here must ship with a test in the same PR; don't let
  this package accumulate untested files.
