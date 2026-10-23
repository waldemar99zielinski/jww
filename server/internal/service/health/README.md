# internal/service/health

Owns both process health operations:

- `CheckHealth` always reports `ok` and performs no dependency I/O. A running
  process stays live even when an external service is unavailable.
- `CheckReadiness` runs every dependency check and reports `ready`, `degraded`,
  or `not_ready`, plus whether the instance should receive traffic.

`database` is required: its failure makes readiness `not_ready`. `openai`,
`object_storage`, and `frankfurter` are optional capabilities: failures mark
readiness degraded while leaving the instance ready so core finance operations
remain available.

`config.go` owns `Config{ LatencyWarnMS }` (`READINESS_LATENCY_WARN_MS`,
default `1000`). `CheckReadiness` logs a WARN per unavailable dependency, or an
available dependency at/above the threshold, under `component=service.health`.
This is the documented logging exception for a service package.

Both public methods accept the caller's `domain.AppContext`, matching the
service-layer convention; `CheckHealth` deliberately ignores it. The service
owns the small `Checker` port; outbound adapters implement it without importing
HTTP concerns. Checks run concurrently using the app context's underlying
`context.Context`, and reports retain constructor order. Each result records the
latency of its adapter call. Errors are retained for internal use but must never
be serialized to clients.

## Testing requirement for this directory

Use local `CheckFunc` fakes. Verify `CheckHealth` is always `ok` and never calls
a checker, then cover all-available, optional failure, and required failure
behavior for `CheckReadiness` with no real I/O. `Config.LatencyWarnMS` has a
1000ms fallback for zero values.
