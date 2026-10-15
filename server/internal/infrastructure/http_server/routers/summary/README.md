# routers/summary

**Role.** Read-only summary report (protected only). All math is delegated to `SummaryWorkflow`.

## Routes (all `protected`)

| Method | Path | Handler |
|---|---|---|
| GET | `/v1/summary` | `getSummaryHandler` |

`Mount(_, protected, deps)`.

## Notes

`getSummaryHandler` reads query params and delegates:

- `period_end` — required, `schema.ParseDate`.
- `currency_code` — required (missing → local sentinel `errMissingCurrencyCode`).
- `period` — optional, validated via `domain.SummaryPeriod.Valid()` (else `errInvalidPeriod`).
- `period_start` — optional, `schema.ParseDate`.
- `user_id` — optional, `uuid.Parse`.

Then `deps.SummaryWorkflow.Generate(appCtx, period, periodEnd.Time(), currencyCode, periodStart, userID)` → `schema.SummaryToResponse`.

## Avoid

- Doing any conversion/aggregation here — that lives in `SummaryWorkflow`.
- Handler-local sentinels are fine here; promote to `kernel/query_errors.go` only if shared.

## Testing requirement for this directory

**Location:** `tests/` subdirectory (black-box, imports this package's
exported API — see `server/TESTING.md#conventions`).

`httptest` + testify, real services built from local hand-written fake
repositories (no real Postgres) — see `server/TESTING.md`. This handler is
mostly query-param validation, so focus coverage there: happy path, each
required param missing, each malformed value (`period_end`, `period`,
`user_id`), and the workflow's own `NotFoundError` (unknown currency)
surfacing as 404. The aggregation math itself is `SummaryWorkflow`'s own
test suite's job, not this one's.
