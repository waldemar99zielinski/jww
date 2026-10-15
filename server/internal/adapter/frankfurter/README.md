# internal/adapter/frankfurter

**Clean Architecture layer: Interface Adapter (outbound/driven).** Implements the `RateSource` port.

**Role.** Implements `currency_exchange_rate.RateSource` against the [Frankfurter](https://frankfurter.dev) API — the **only** place that speaks HTTP to an external rate provider. Anonymous, no API key. Owns `frankfurter.Config` (`FRANKFURTER_*`).

## Files

- `frankfurter.go` (package doc; the main file) — `Client` + `NewClient(cfg)`; `Rates(ctx, currencies, asOf)` (the matrix orchestration) and the private single-base `rates(...)` HTTP request.
- `config.go` — `Config{ BaseURL, TimeoutSeconds }` (`FRANKFURTER_BASE_URL` default `https://api.frankfurter.dev`; `FRANKFURTER_TIMEOUT_SECONDS` default `15`); `LoadConfig()`.
- `response.go` — the `rateResponse` wire DTO (`json.Number` rate) and `mapRates(...)`, which turns one base's decoded rows into flatten pairs + skips.

## Behavior

- `Rates` makes **one request per currency as base** (`/v2/rates?base=…&quotes=…[&date=…]`), collecting all N×(N−1) directed pairs directly from the provider — no row is ever derived by computing `1/rate`. A zero `asOf` omits `&date=` (latest day).
- Every requested quote the provider omits (absent from the array, or an empty array) becomes a `RateSkip{From, To, Reason}` — never dropped silently. A non-200 status or a decode failure aborts the whole fetch.
- The provider's echoed date becomes `RecordedAt` at midnight UTC — the observation date, not the fetch time.

## Rules

- **Never decode the rate as `float64`.** The JSON `rate` field is a `json.Number`; its text goes straight to `domain.ParseExchangeRate` (fixed-point 1e-4).
- **Adapters don't log.** Return the error / the skips and let `cmd/server` (via the scheduled job closure) log them once.
- Imports: stdlib, `domain`, `service/currency_exchange_rate`, `utils/configuration`. Nothing else internal.

## Testing requirement for this directory

**Location:** `tests/` subdirectory (black-box, imports this package's
exported API — see `server/TESTING.md#conventions`).

Black-box (`package frankfurter_test`) against `httptest.NewServer` — never a
mocked `http.Client`. Cover: one batch per currency requested with the target
`date`; a zero `asOf` omits `date`; readable pairs come back as
`domain.CurrencyExchangeRateFlatten`; an omitted quote and an empty array each
become a `RateSkip` per requested quote; the echoed date becomes `RecordedAt`;
`500` and malformed JSON both error; a 5-decimal rate parses to the right
fixed-point value.
