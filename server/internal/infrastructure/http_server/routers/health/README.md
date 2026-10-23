# routers/health

Public process health HTTP adapter:

- `GET /api/v1/health` calls `health.Service.CheckHealth` and returns `200` with
  `{"data":{"status":"ok"}}`. It performs no dependency I/O, so it is pure
  process liveness.
- `GET /api/v1/readiness` calls `health.Service.CheckReadiness`. It checks
  PostgreSQL, OpenAI, object storage, and Frankfurter; PostgreSQL is the only
  hard requirement. Success returns `200 {"data":{"status":"ok"}}`; failure
  returns the standard sanitized `503` error envelope. The response never
  includes dependency names, statuses, latencies, or errors.
- Each unavailable dependency, or successful check at/above
  `READINESS_LATENCY_WARN_MS` (default 1000ms), is logged by
  `health.Service` at WARN under `component=service.health`. Detailed status
  stays server-side.

Health is dependency-free liveness; readiness owns all dependency checks.
Responses are non-cacheable. Neither probe is logged by `kernel.RequestLogger`
(see `isHealthRoute`); only readiness warnings above still reach the log.

## Testing requirement for this directory

Keep handler tests under `tests/` and verify public status and standard JSON
response envelopes. Build a real `health.Service` over local fake checkers and
verify health is always `ok` without invoking dependency checkers, readiness is
`ok` for ready/degraded reports, unavailable required dependencies return `503`,
and neither response leaks dependency details. Readiness policy behavior is
covered by `internal/service/health` tests. Decode JSON rather than asserting
raw body strings.
