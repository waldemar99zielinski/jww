# internal/service/currency_exchange_rate

**Role.** Create/read use cases for exchange-rate records. **No UoW manager** — every method takes/forwards the optional `uow`, because each op is a single repo call; callers supply the transaction.

## Files

- `repository.go` — `Repository`: `Create(...)`, `UpsertMany(ctx, []domain.CurrencyExchangeRateFlatten, uow...) (UpsertCounts, error)`, `GetByCurrencies(ctx, from, to domain.Currency, asOf time.Time, uow...)`, `List(ctx, filter, uow...)`.
- `filter.go` — `Filter{ FromCurrencyCode, ToCurrencyCode, RecordedFrom, RecordedTo }`; related currencies are represented by explicit codes, and `Validate()` rejects an inverted recorded-at range.
- `currency_exchange_rate.go` — `RateSource`, result types, and `Service{ repository }`; `NewService(repository)` (two-arg-less constructor — no manager).
- `Create(ctx, from, to domain.Currency, rate domain.ExchangeRate, uow ...)` — stamps `RecordedAt: time.Now().UTC()`, `Validate()`, repo `Create`. Left untouched, so the manual HTTP endpoint keeps its "now" timestamp and single-row semantics.
- `UpsertMany(ctx, []domain.CurrencyExchangeRateFlatten, uow ...)` — normalizes every `RecordedAt` to UTC, `Validate()`s the whole batch (first invalid row aborts before I/O), then hands it to the repo as **one** statement that replaces on a `(from, to, recorded_at)` conflict. An empty batch is a no-op.
- `GetByCurrencies(...)` — nil → `domain.NewCurrencyExchangeRateNotFoundError(from.Code, to.Code)`.
- `List`: validates and forwards the optional filter.

## Ports

- `RateSource.Rates(ctx, currencies []string, asOf time.Time) (RateFetch, error)` — the outbound port the daily sync depends on. Returns every directed pair it can read plus a `RateSkip` for anything the provider omitted. Implemented in `internal/adapter/frankfurter`.
- `RateSkip{From, To, Reason}` / `RateFetch{Rates, Skipped}` — the source's return shape. Skips are **returned**, never swallowed, so the boundary can log them.
- `UpsertCounts{Inserted, Updated}` — how a batch changed the table.

## Rules

- `CurrencyExchangeRateFlatten` (codes only) is the sync/write shape; the nested `domain.CurrencyExchangeRate` stays the read/HTTP shape. Both `Validate()` the same two CHECK constraints.
- `asOf` is forwarded to the repo for "latest rate at or before this instant" (gateway: `ORDER BY recorded_at DESC LIMIT 1`).
- `Create` timestamps come from `time.Now().UTC()` here, never the handler; `UpsertMany` trusts the caller's business-date `RecordedAt`.

## Avoid

- Orchestrating multiple services (resolve currencies + fetch rates) — that belongs in `workflow.CurrencyExchangeRateWorkflow`.
- Logging here (services never log) — return the skips/errors instead.
- Imports: `domain`, `unit_of_work`, `time`.

## Testing requirement for this directory

**Location:** `tests/` subdirectory (black-box, imports this package's
exported API — see `server/TESTING.md#conventions`).

Black-box (`package currency_exchange_rate_test`), testify, a local
hand-written fake `Repository` — see `server/TESTING.md`. No `uowManager`
here, so no Restore/Finish assertions needed (that's the caller's job).
Needs: `Create`'s validation failures (`from == to`, non-positive rate)
never reach the repo; `GetByCurrencies`'s `(nil, nil)` →
`domain.NewCurrencyExchangeRateNotFoundError` translation; `List` is a
passthrough; `UpsertMany` — empty batch is a no-op that never touches the
repo; a happy batch is forwarded and returns the repo's inserted/updated
counts; the first invalid row aborts before any I/O; `RecordedAt` is
normalized to UTC.
