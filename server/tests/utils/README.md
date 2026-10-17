# tests/utils

**Role.** Generic helpers shared across the suites under `server/tests/`
(currently just [`tests/e2e`](../e2e/README.md)) — HTTP test driving, real
test-dependency wiring, and small OS-level utilities. This is a plain,
importable package (no `_test.go` files, no build tag), not a test package
itself: Go can't import identifiers from another package's `_test.go`
files, so anything meant to be reused across test packages has to live in
an ordinary package instead — the same reasoning behind
[`internal/tests/pgtest`](../../internal/tests/pgtest/README.md).

## API

- `Envelope[T]` — mirrors `kernel.Envelope`/`ErrorEnvelope` loosely enough
  to decode either `{"data": ...}` or `{"error": {...}}` without importing
  the kernel package's unexported internals.
- `Client` (`BaseURL string`, `HTTP *http.Client`) + `NewClient(baseURL,
  httpClient)` — `(*Client) Do(t, method, path, token, body, v) int` sends a
  request and decodes the envelope into `v`, returning the status code;
  `(*Client) Login(t, username, password) string` logs in and returns the
  access token. `(*Client) UploadFile(t, path, token, filename, contentType,
  content) (*http.Response, []byte)` posts one streaming multipart `file`
  part and returns the raw response; `(*Client) Download(t, path, token) (int,
  []byte, http.Header)` gets the one bytes-not-envelope endpoint.
  `DecodeBody(t, resp, v)` is the shared raw decode step.
- `BuildTestDeps(ctx, pool *pgxpool.Pool, jwtSecret string, storageConfig
  storage.Config) (*kernel.Deps, error)` — wires every
  repository/service/workflow against `pool`, mirroring
  `cmd/server/deps.go`'s `buildDeps` (including the S3 attachment storage
  adapter, pointed at the caller's MinIO via `storageConfig`). Duplicated
  rather than imported because `buildDeps` is unexported in `package main` —
  keep the two in sync if that wiring changes; see `cmd/server/README.md`.
- `FreePort() (int, error)` — asks the OS for an ephemeral TCP port and
  releases it immediately, for a test that needs to bind its own server.
- `ModuleRoot(t testing.TB) string` — finds the directory containing
  `go.mod` by searching upward from the *caller's* source file. Prefer this
  over hardcoding a relative path like
  `filepath.Dir(filepath.Dir(thisFile))` — that kind of depth-counting
  breaks silently the next time a test package moves (as `tests/e2e`
  already has once).

## Testing requirement for this directory

No tests of its own — every function here is exercised transitively by
every consumer under `tests/` (currently `tests/e2e`'s full suite). If you
add non-trivial logic (not just a thin wrapper), add a white-box test for
it directly, the same way `internal/tests/pgtest` tests itself.

## Avoid

- Journey-specific or endpoint-specific logic here — that belongs in the
  consuming suite (e.g. `tests/e2e`'s own journey test files).
- Importing this package from a `_test.go` file in a package that isn't
  under `tests/` — if `internal/adapter/gateway` needs a similar helper,
  that's `internal/tests/pgtest`'s job, not this package's (different
  concerns: `pgtest` is Postgres-container lifecycle, this is
  HTTP-client/deps-wiring/OS-level test driving).
