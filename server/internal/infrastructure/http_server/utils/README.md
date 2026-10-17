# internal/infrastructure/http_server/utils

**Role.** Small, reusable, resource-agnostic utilities for HTTP-level values. This package is a dependency leaf within `http_server`: `kernel` and resource routers may import it, but it does not import those packages.

## Files

- `path.go`
  - `MatchPath(pattern, path) PathMatch` — matches paths segment by segment. Segments written as `{name}` capture a non-empty value in `PathMatch.Params`; `PathMatch.Matched` reports whether the complete path matched.
  - Matching is exact: missing, empty, additional, or trailing segments do not match.
- `path_params.go`
  - `ParseInt64PathParam(r, name)` parses a named Chi path parameter as an `int64`.
  - `ParseUUIDPathParam(r, name)` parses a named Chi path parameter as a UUID.
- `path_test.go` — covers literal matching, multiple parameter extraction, and rejected path shapes.

Example:

```go
result := utils.MatchPath(
    "/api/v1/transactions/{transactionID}/attachments",
    r.URL.Path,
)
if result.Matched {
    transactionID := result.Params["transactionID"]
    // ...
}
```

## What belongs here

- Pure helpers for parsing, matching, or normalizing HTTP-specific values.
- Helpers useful to more than one HTTP package, or needed by `kernel` without belonging to middleware/request/response orchestration.
- Code whose inputs and outputs are ordinary values and that requires no application state.

## What does not belong here

- Middleware, request context, response encoding, error-to-status mapping, multipart processing, or dependencies — put these in [`../kernel`](../kernel/README.md).
- Route registration and endpoint behavior — put these in [`../routers/<resource>`](../routers/README.md).
- Server construction, CORS setup, or package wiring — keep these in the root [`http_server`](../README.md) package.
- Domain rules, service calls, repositories, logging, configuration, or resource-specific constants.

## Dependency rules

- Prefer the Go standard library; add no domain, service, adapter, `kernel`, router, or root `http_server` imports.
- Keep helpers deterministic and free of I/O, global state, and side effects.
- Do not turn this package into a catch-all. A helper that is only meaningful to one resource stays with that resource.
- Export only utilities intended for use by another HTTP package.

## Testing requirement for this directory

Keep tests alongside the source as `*_test.go`. Use table-driven unit tests where appropriate and cover successful values, malformed input, empty values, and boundary shapes. These utilities should not require an HTTP server, database, services, or other external dependencies.
