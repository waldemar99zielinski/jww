# routers/currency_exchange_rate

**Role.** Create/get/list currency exchange rates (protected only). Writes go through `CurrencyExchangeRateWorkflow` (enforces cross-currency rules); reads too.

## Routes (all `protected`)

| Method | Path | Handler |
|---|---|---|
| POST | `/v1/currency-exchange-rates` | `createExchangeRateHandler` |
| GET | `/v1/currency-exchange-rates` | `getExchangeRateHandler` (query: `from_currency_code`, `to_currency_code`, `as_of`) |
| GET | `/v1/currency-exchange-rates/all` | `listExchangeRatesHandler` (optional query: `from_currency_code`, `to_currency_code`, `recorded_from`, `recorded_to`) |

`Mount(_, protected, deps)`.

## Notes

- Create decodes `schema.CurrencyExchangeRateCreateRequest` then calls `Workflow.Create`.
- Get parses `as_of` via `schema.ParseDate`.
- List builds the shared `currency_exchange_rate.Filter`; recorded range bounds use RFC3339 timestamps.
- **Routing hazard:** `/v1/currency-exchange-rates` and `/v1/currency-exchange-rates/all` are distinct chi patterns; don't later add a `{param}` route that would shadow `/all`.

## Testing requirement for this directory

**Location:** `tests/` subdirectory (black-box, imports this package's
exported API — see `server/TESTING.md#conventions`).

`httptest` + testify, a real `CurrencyExchangeRateWorkflow` built from
local hand-written fake repositories (no real Postgres) — see
`server/TESTING.md`. Needs: create happy path and its validation-failure
400; get happy path, missing-required-query-param 400, and not-found 404;
list happy path. The workflow's own cross-currency rules are its own test
suite's job, not this one's.
