# scripts/syncrates

**Clean Architecture layer: Frameworks & Drivers (composition root; outermost).** Standalone tool — not part of the server binary.

**Role.** Backfill currency exchange rates for an explicit date range. Each calendar day is handed to the same `CurrencyExchangeRateWorkflow.Sync` the nightly scheduler calls, so the rows are produced by the server's own workflow, gateway, and Frankfurter adapter — no SQL or HTTP is re-implemented here. Unlike [`bootstrap`](../bootstrap/README.md), which seeds a fixed set of currencies, this tool fills in rate history for whatever currencies the database already holds.

## Files

- `syncrates.go` — `Client` + `NewClient(pool, rateSource)`, which wires the real currency service, exchange-rate service, and `CurrencyExchangeRateWorkflow` exactly like `cmd/server/deps.go`. `SyncRange(ctx, from, to, includeWeekends)` is the range loop.
- `syncrates_test.go` — fast unit tests for the range logic (inclusive bounds, UTC calendar days, weekend skipping, failed-day isolation, aggregate counts) using a fake syncer.
- `cmd/main.go` — the command: resolves flags/env, configures the logger and Frankfurter client, opens the pgx pool, runs the range, and logs each day and the totals.

## Usage

From the `server/` directory:

```bash
go run ./scripts/syncrates/cmd \
  -from 2026-01-01 -to 2026-03-31 \
  -database-url postgres://user:pass@localhost:5432/finance
```

…or via environment variables; the flag wins when both are set:

```bash
DATABASE_URL=postgres://user:pass@localhost:5432/finance \
RATES_FROM=2026-01-01 RATES_TO=2026-03-31 \
go run ./scripts/syncrates/cmd
```

Supplying no database URL or no bound fails with an error.

## Parameters

| Flag | Environment variable | Required | Meaning |
|---|---|---|---|
| `-database-url` | `DATABASE_URL` | yes | `postgres://` DSN of the target database. |
| `-from` | `RATES_FROM` | yes | First day to fill, `YYYY-MM-DD` (UTC). |
| `-to` | `RATES_TO` | yes | Last day to fill, inclusive, `YYYY-MM-DD` (UTC). May be today, never later. |
| `-include-weekends` | `RATES_INCLUDE_WEEKENDS` | no (default `false`) | Also fetch Saturdays/Sundays instead of skipping them. |

## Notes

- **Needs at least two currencies** in the database; with fewer, the workflow short-circuits and writes nothing. Seed them first with [`bootstrap`](../bootstrap/README.md) or [`addcurrency`](../addcurrency/).
- **Weekends are skipped by default**: the provider publishes business days only, and a weekend request would just echo the previous business day. Holidays are not detected locally — those days are fetched and simply produce warnings if the provider has no new data.
- **Idempotent**: rates are upserted on `(from, to, recorded_at)`, so re-running a range updates rows in place instead of duplicating them.
- **One failed day does not abort the range** — the error is logged for that day, remaining days still run, and the command exits non-zero with `N of M day(s) failed`.
- The range is fetched one day at a time through the workflow, which makes one provider request per currency per day. Long ranges take a while by design; the process has no fixed deadline and stops cleanly on `SIGINT`/`SIGTERM`.
- Does **not** run migrations — bring the schema up first.
- Respects `FRANKFURTER_BASE_URL` / `FRANKFURTER_TIMEOUT_SECONDS` through `frankfurter.LoadConfig()`.

## Testing requirement for this directory

The range logic is unit-tested with a fake `syncer` (`go test ./scripts/syncrates/...`), because it is real logic beyond flag/env resolution. The database write and provider fetch are already covered by the `workflow`, `gateway`, and `frankfurter` test suites, so this package does not duplicate them with an integration test.
