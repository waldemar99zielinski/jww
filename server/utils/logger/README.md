# utils/logger

**Role.** Shared structured logger (`log/slog`) wrapper. Builds a process-wide base handler configured **once**, then derives child loggers carrying component labels and, per request, `domain.AppContext` fields.

## Where to log

**Logging lives at the boundary** — wherever a request, process, or command
actually enters — never inside the layers it calls into. A `service`/`adapter`
function that hits trouble returns an error; it does not log it. The caller
that already has the request/process context logs it once.

| Layer | Logs? | Component label |
|---|---|---|
| `internal/domain` | **Never** — no I/O; importing this package here would cycle (domain is imported *by* logger) | — |
| `internal/service/*` (use cases) | **Almost never** — the default allowlist is `domain` + `service/unit_of_work` + stdlib, so most packages can't import this one at all; propagate a typed error instead. A use case may import it for **`Debug`-only step tracing** (see `transaction_attachment_extraction`): component `service.<pkg>`, `WithContext` on every line, never above `Debug`, never secrets | `service.<pkg>` |
| `internal/adapter/*` (gateway/schema/security) | **Never**, by convention (not a cycle) — let the error bubble to the one caller that logs; logging it here too just duplicates the same failure | — |
| `internal/infrastructure/database` | Yes — migration/startup events | `database` |
| `internal/infrastructure/http_server/kernel` | Yes — request lifecycle (`RequestLogger`, `Recoverer`) | `http_server` |
| `internal/infrastructure/http_server/routers/<resource>` | **Mostly no** — `kernel`'s `RequestLogger` middleware already logs every invocation (`"request completed"`, method/path/status/duration). Declare a package `var log = logger.New().WithComponent("http_server.<resource>")` only when a handler must log something the middleware can't see; today only `transaction_attachment` does, for extraction outcomes (attachment id, duration, failure kind, unwrapped cause) | `http_server.<resource>` (only `http_server.transaction_attachment` today) |
| `cmd/server` | Yes — process lifecycle (startup, listening, shutdown, fatal error) | `cmd.server` |
| `scripts/<name>` | Yes — CLI-style entry points, business events | `scripts.<name>` |

If you're adding code to `internal/service` or `internal/adapter` and feel the
urge to log: that's a sign the information belongs in the returned error (or
the call succeeds and the boundary's existing log line already covers it).

## API

- `Configure(Config) error` — `sync.Once`-guarded; sets level and picks JSON (default) or the pretty console handler (`Pretty`). Later calls are no-ops returning the cached error.
- `New() Logger`, `With(labels ...slog.Attr) Logger`, `Slog() *slog.Logger`.
- `(Logger) WithComponent(name string) Logger` — adds `component=<name>`.
- `(Logger) Slog() *slog.Logger` — base + accumulated labels.
- `(Logger) WithContext(ctx domain.AppContext) *slog.Logger` — base + labels + `flow_id`, `source_type`, `http_endpoint` (when the source type is `http`), `job` (when the source type is `scheduler`), `user_id` (when present).

Config: `LOG_LEVEL` (`info`|`debug`|`warn`|`error`), `LOG_PRETTY`.

## Formats

- **JSON** (`LOG_PRETTY=false`, default) — `slog.NewJSONHandler`, one object per line; for production and machine parsing.
- **Pretty** (`LOG_PRETTY=true`) — [`tint`](https://github.com/lmittmann/tint) (`tint.NewTextHandler`), colorized one-line records for local development. Colors are suppressed (via `go-isatty`) when stdout isn't a terminal, so piped/redirected output stays plain.

## Rules

- Configure **once** at startup (`cmd/server`), before anything logs.
- **Component labels are dot-delimited:** `cmd.server` for the process, `http_server` for the kernel, `database` for migrations, `scheduler` for the jobs runner, `scripts.<name>` for CLI entry points, `http_server.<resource>` for a router that logs, and `service.<pkg>` for a use case's `Debug` trace. Never log in application code without a component.
- Prefer `WithContext(appCtx)` so request attributes attach automatically.
- Declare one package-local `var log = logger.New().WithComponent("...")` per package; don't share another package's logger.
- **Typed attrs, not formatted strings:** `slog.String("user_id", id)`, not `fmt.Sprintf("user_id=%s", id)` baked into the message — keeps fields queryable in the JSON output.
- **Levels:** `Info` for expected lifecycle/business events ("request completed", "migration applied", "account created"); `Warn` for client-caused problems (the request-logging middleware already does this by status, automatically — don't hand-pick a level yourself for that path); `Error` for unexpected/internal failures, always paired with `slog.String("error", err.Error())`; `Debug` for verbose diagnostic detail, off by default (`LOG_LEVEL=debug`).
- **Never log secrets:** passwords, tokens/JWTs, the `Authorization` header value, `JWT_SECRET_KEY`, DB DSNs/connection strings, password hashes. Identifiers (`user_id`, `username`, `flow_id`) are fine — they aren't secrets.
- **Logging is independent of response/return-value assembly.** A function whose job is to build a response body, a DTO, or any other return value must stay pure with respect to logging — it must not also decide to log as a side effect. Log explicitly at the call site (or in middleware) instead of embedding a log call inside an assembler/writer/mapper. `http_server/kernel/respond.go`'s `WriteJSON`/`WriteBadRequest`/`writeError` family is meant to be exactly this kind of pure assembler — don't add new logging calls there, even for a seemingly-reasonable case (e.g. "log every error response"); that belongs in the handler or in middleware instead. (`WriteError`'s existing `httpLog.Slog().Error("internal error", ...)` line predates this rule and is a known, narrow exception — don't use it as a template for more logging in that file.)

Canonical usage (a component that does log, e.g. `scripts/<name>` or `internal/infrastructure/database`):

```go
var log = logger.New().WithComponent("scripts.<name>")
// ...
appCtx := someAppContext()
log.WithContext(appCtx).Info("account created", "username", account.Username)
```

Routers don't currently follow this pattern — see "Where to log" above.

## How to extend

- New `AppContext`-derived field: extend `appContextAttrs` in `logger.go` (keep attrs minimal/stable).
- New log level/format: change `parseLevel` / `Configure`; do not create ad-hoc `slog.New` handlers in feature code.

## Avoid

- Ad-hoc `slog.New(...)` handlers in features — always use the shared base.
- Adding a request field to logs outside `WithContext`/`appContextAttrs`.
- Making `internal/domain` import this package (would cycle: logger → domain already).

## Testing requirement for this directory

No tests yet. If you add logic here beyond config/wiring (e.g. a new
`parseLevel` branch, a new `appContextAttrs` field), add a white-box
(`package logger`) stdlib test for it — redirecting `os.Stdout` does **not**
work for testing this package's own output, because `base` is a
process-global `slog.Logger` built once; see
[`kernel`'s testing requirement](../../internal/infrastructure/http_server/kernel/README.md#testing-requirement-for-this-directory)
for the same gotcha hit (and worked around) from the consumer side. Prefer
testing pure functions (`parseLevel`, `appContextAttrs`) directly over
trying to capture log output.
