# internal/infrastructure/http_server

**Clean Architecture layer: Frameworks & Drivers (HTTP entry / composition).** Assembles `kernel` (shared HTTP adapters), `routers` (resource handlers), and `utils` (dependency-free HTTP utilities) into a running server.

**Role.** The HTTP **composition root**. Builds the chi router, attaches cross-cutting middleware, and hands each resource subpackage a **public** router and a **protected** router (pre-guarded with `RequireAuth`) so each resource decides where each route belongs. Also owns `http_server.Config`.

> This package imports `kernel` + every `routers/*`. Nothing here may be imported by `kernel` or `routers/*` — that would be an import cycle.

## Directory ownership

- [`kernel/`](kernel/README.md) — shared HTTP application plumbing: dependencies, middleware, request context, decoding, response writing, multipart handling, and HTTP error mapping. Code here may depend on domain and service types.
- [`routers/`](routers/README.md) — resource-specific route registration and handlers. Put endpoint behavior here, grouped by resource.
- [`utils/`](utils/README.md) — small, resource-agnostic, dependency-free HTTP utilities. These helpers may be imported by `kernel` and routers, so they must not import either of them.
- [`web/`](web/README.md) — optional React build config, validation, static/SPA handler, and browser
  security middleware. It has no dependency on the parent package or API router.
- Root package (`server.go`, `config.go`) — compose the listener, API router, optional injected web
  handler, middleware, and top-level `/api` dispatch.

When deciding where code belongs:

1. If it implements one resource's endpoint behavior, put it in `routers/<resource>`.
2. If it coordinates shared request/response behavior or needs `Deps`, domain, or service types, put it in `kernel`.
3. If it is a small pure helper that only understands HTTP-level values such as paths and has no application dependencies, put it in `utils`.
4. If it wires packages together or configures the running HTTP server, keep it in the root package.

## Files

- `server.go` (package doc)
  - `New(deps, cfg, webCfg) (*http.Server, error)` — receives HTTP and web configuration
    separately, builds the complete handler, and applies fixed connection timeouts.
  - `NewHandler(deps, webCfg) (http.Handler, error)` — returns the API router in API-only mode, or
    creates the injected, security-wrapped web handler and dispatches `/api` in
    `newApplicationHandler`. API middleware remains owned by `NewRouter`.
  - `NewRouter(deps *kernel.Deps) http.Handler`:
    ```go
    r := chi.NewRouter()
    r.Use(kernel.RequestLogger)   // outermost: sees final status
    r.Use(kernel.Recoverer)
    r.Use(kernel.LimitRequestBody(deps.MaxBodyBytes)) // caps every body before it reaches a handler (except the two attachment-upload routes)
    r.Use(cors.Handler(cors.Options{ AllowedOrigins: deps.CorsOrigins, /* ... */ }))

    r.Route("/api", func(api chi.Router) {
        public := api
        protected := api.With(kernel.RequireAuth(deps)) // same tree + auth
        auth.Mount(public, protected, deps)
        // ... every resource
    })
    ```
- `config.go` — HTTP listener and API middleware settings only; `Addr()`, `CorsOriginsList()`.
- `routers/health` mounts public `GET /api/v1/health` (dependency-free liveness) and `GET
  /api/v1/readiness` (PostgreSQL is required; OpenAI, object storage, and Frankfurter are optional).
  Readiness responses expose only `ok` or the standard sanitized `503`; unavailable or slow checks
  are logged at WARN by `health.Service` under `service.health`.
- `routing_verify_test.go` — table test asserting public vs protected status codes. Extend it when auth expectations change.
- `utils/path.go` — matches route-like path patterns and extracts named path segments; see [`utils/`](utils/README.md).

## Rules

- Middleware order matters: `RequestLogger` outermost, then `Recoverer`, then `LimitRequestBody`, then CORS; all `Use` before `Route`.
- `public` and `protected` share **one** chi tree (`With` copies the tree and prepends middleware). Pass both to every `Mount`.
- This is the only importer of `routers/*`.
- Keep `utils` dependency-free within the HTTP server: `kernel` and routers may import it, but it must not import them or the root package.

## How to add a resource router

1. Create `routers/<resource>/` with `router.go` (`Mount`) and one file per handler — see [`routers/`](routers/README.md).
2. Add any new service/workflow to `kernel.Deps` and `cmd/server/deps.go`.
3. Register `<resource>.Mount(public, protected, deps)` inside `r.Route("/api", ...)` here.
4. Add request/response shapes to `internal/adapter/schema`.

## Avoid

- **`"*"` in `cors.Options.AllowedMethods`** — go-chi/cors expands `*` only for `AllowedOrigins` and `AllowedHeaders`. For methods it matches literally, so no real method is allowed and `Access-Control-Allow-Origin` is never written. List methods explicitly.
- **Registering the same method+path twice** (e.g. on both `public` and `protected`) — chi panics at startup.
- Defining handlers or routes here; constructing services/repos here.
- Attaching middleware after routes are mounted.
- Importing this package from `kernel` or any router.

## Testing requirement for this directory

**Location:** alongside the source, not a `tests/` subdirectory — these are
white-box (`package http_server`), reaching this package's unexported
router wiring. See `server/TESTING.md#conventions`.

`cors_test.go` and `routing_verify_test.go` (both white-box, `package
http_server`) cover CORS and the public/protected split directly — extend
`routing_verify_test.go` whenever a route's auth expectations change. Deeper
middleware behavior is [`kernel`](kernel/README.md#testing-requirement-for-this-directory)'s
test suite; per-handler behavior is each [`routers/<resource>`](routers/README.md#testing-requirement-for-this-directory)'s.
`server_handler_test.go` locks down API/web dispatch. Tests alongside the files under `web/` cover
startup validation, static/SPA routing, and browser security policy. See `server/TESTING.md`.
