# internal/infrastructure/http_server/web

**Role.** Owns the optional production React application boundary. This package does not import the
parent `http_server` package and does not know how API routes are mounted.

## Files

- `config.go` — `Config{StaticDir, APIBaseURL}`, loaded from `WEB_STATIC_DIR` and
  `WEB_API_BASE_URL`. An empty static directory means API-only mode. Validation checks the build
  directory, `index.html`, and the optional public HTTP(S) API origin before startup.
- `handler.go` — serves existing Vite files, generates non-cacheable runtime config, and provides
  the `Accept: text/html` BrowserRouter fallback. Static files otherwise use `http.FileServer`'s
  standard behavior.
- `security.go` — a private wrapper installed by `New` that adds CSP, cross-origin isolation,
  clickjacking and MIME protections, referrer/device policy, and HSTS to web responses. The
  configured API origin is allowed for API calls and attachment media, but never for scripts.

## Composition

`cmd/server` loads `web.Config` and injects it separately from `http_server.Config`:

```go
srv, err := httpserver.New(deps, cfg.Http, cfg.Web)
```

The parent server owns the route boundary:

```text
/api, /api/*  -> API router
everything else -> web.Handler
```

Keep API routing out of this package and web filesystem behavior out of the parent server. Keep
optional performance tuning such as compression or custom asset caching at the deployment edge.
