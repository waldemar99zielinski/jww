# internal/service/currency_exchange_rate

**Role.** Create/read use cases for exchange-rate records. **No UoW manager** — every method takes/forwards the optional `uow`, because each op is a single repo call; callers supply the transaction.

## Files — `currency_exchange_rate.go`

- `Repository`: `Create(...)`, `GetByCurrencies(ctx, from, to domain.Currency, asOf time.Time, uow...)`, `List(ctx, uow...)`.
- `Service{ repository }`; `NewService(repository)` (two-arg-less constructor — no manager).
- `Create(ctx, from, to domain.Currency, rate domain.ExchangeRate, uow ...)` — stamps `RecordedAt: time.Now().UTC()`, `Validate()`, repo `Create`.
- `GetByCurrencies(...)` — nil → `domain.NewCurrencyExchangeRateNotFoundError(from.Code, to.Code)`.
- `List`: passthrough.

## Rules

- `asOf` is forwarded to the repo for "latest rate at or before this instant" (gateway: `ORDER BY recorded_at DESC LIMIT 1`).
- Timestamps come from `time.Now().UTC()` here, never the handler.

## Avoid

- Orchestrating multiple services (resolve currencies + write rate) — that belongs in `workflow.CurrencyExchangeRateWorkflow`. Imports: `domain`, `unit_of_work`, `time`.

## Testing requirement for this directory

**Location:** `tests/` subdirectory (black-box, imports this package's
exported API — see `server/TESTING.md#conventions`).

Black-box (`package currency_exchange_rate_test`), testify, a local
hand-written fake `Repository` — see `server/TESTING.md`. No `uowManager`
here, so no Restore/Finish assertions needed (that's the caller's job).
Needs: `Create`'s validation failures (`from == to`, non-positive rate)
never reach the repo; `GetByCurrencies`'s `(nil, nil)` →
`domain.NewCurrencyExchangeRateNotFoundError` translation; `List` is a
passthrough.
