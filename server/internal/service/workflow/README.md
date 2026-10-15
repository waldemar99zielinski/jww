# internal/service/workflow

**Clean Architecture layer: Use Cases (orchestration).** Coordinates several use-case services behind one operation.

**Role.** Use-case orchestration across multiple services. Workflows exist where an operation must resolve referenced resources (currency by code, category/user by id) and/or span several service calls atomically. They are the **only** packages that compose sibling services and the **only** place in the service layer that owns transaction demarcation for multi-service operations.

## Files

- `currency_exchange_rate_workflow.go` — `CurrencyExchangeRateWorkflow{ currencyService, exchangeRateService, rateSource, uowManager }`. Own-UoW: resolve `from`/`to` currencies, then `exchangeRateService.Create(..., active)`, `Finish`. Also owns the daily sync: `SyncDaily(ctx)` → `Sync(ctx, yesterday)`, which lists every currency, fetches the full directed matrix from the `RateSource` port, filters to known codes, writes the batch with one `UpsertMany` (**no UoW** — one call), and returns `SyncResult{Currencies, Fetched, Inserted, Updated, Skipped, Warnings}`. It is base-agnostic (the base flag is never consulted).
- `transaction_workflow.go` — `TransactionWorkflow{ currencyService, categoryService, userService, transactionService, uowManager }`. `Create` resolves currency/category/user then creates the transaction in one UoW. `List(ctx, TransactionListFilter)` resolves optional raw ids/codes into a `transaction.Filter`, validates, then lists.
- `signup.go` — `SignupWorkflow{ userService }`; thin wrapper over `userService.Create` (single service call, no explicit UoW).
- `summary.go` — `SummaryWorkflow{ currencyService, userService, transactionService, currencyExchangeRateService, uowManager }`. Resolves currency/user, normalizes dates, lists transactions in one UoW, converts each into the target currency, builds and validates `domain.Summary`.

## Rules

- Depend on **concrete** `*x.Service` pointers, not interfaces (only services define ports).
- **A single service call needs no unit of work.** If an operation delegates to exactly one service call (no multi-call resolution, no multi-service composition), call it with no `uow` and don't create one — wrapping one statement in a transaction only adds a pointless round-trip. See `CurrencyExchangeRateWorkflow.Sync` (one `UpsertMany`) vs `Create` (resolve `from`, resolve `to`, then create).
- Own the transaction for **multi-call** operations: `active, err := w.uowManager.Create(ctx)`; pass `active` into every service call that accepts an optional `uow`; `_ = active.Restore(ctx)` on every error; `active.Finish(ctx)` once at the end.
- Only call service methods that **accept/forward `uow`** (`Create`/`GetBy*`/`List`) inside a workflow's UoW. A service's `Update`/`Delete` creates its own UoW — **never** call them inside a workflow's UoW (would nest).
- Translate raw request identifiers (codes, ids) into resolved `domain` structs before calling services; services never resolve FKs.
- Timestamps/date normalization live here or in services, never in handlers.
- Passthrough methods are added so every request for a resource still enters through one workflow type.

## How to add a workflow

1. Add `<resource>_workflow.go`, struct with the concrete services it needs + `uowManager` (omit only if genuinely single-service).
2. `New...Workflow(...)`; wire it into `kernel.Deps` and `cmd/server/deps.go`.
3. Multi-call methods: create `active`, resolve references, delegate the write to one service via `active`, `Restore` on error, `Finish` on success.
4. For raw-id request params, define a workflow-local filter/params type (mirroring `TransactionListFilter`).

## Avoid

- Composing services from a plain (non-workflow) service.
- Skipping `Restore` on intermediate errors.
- Importing `adapter/*`, `infrastructure/*`, pgx, `net/http`, or `http_server/kernel` (kernel imports workflows → cycle).
- Per-request workflow construction (they're process-lifetime singletons).

## Testing requirement for this directory

**Location:** `tests/` subdirectory (black-box, imports this package's
exported API — see `server/TESTING.md#conventions`).

Black-box (`package workflow_test`), testify assert+require, no real I/O —
see `server/TESTING.md`.

- **Workflows hold concrete `*x.Service` pointers, not ports**, so you can't
  hand a workflow a fake service directly the way a service test fakes its
  repository. Build the *real* service (e.g. `currency.NewService(...)`)
  wired to a small hand-written fake `Repository`, one layer down — still
  zero I/O, just one layer less "fake." `fakes_test.go` in this package has
  one fake per repository port (`fakeCurrencyRepo`, `fakeCategoryRepo`,
  `fakeUserRepo`, `fakeTransactionRepo`, `fakeExchangeRateRepo`) plus
  `fakeUoW`/`fakeUoWManager`, which record whether `Restore`/`Finish` were
  called.
- **Every error branch inside a workflow's own `uowManager.Create()` block
  needs a test asserting `Restore` fired** (and `Finish` did not) — this is
  the exact class of bug (a missing `Restore` after a mid-loop conversion
  failure in `summary.go`) an earlier audit found by inspection, not by
  test. `summary_test.go`'s "regression" subtest exists specifically so that
  class of bug fails CI instead of waiting for another audit.
- **Assert resolution order where the workflow depends on it**, e.g.
  `TransactionWorkflow.Create` must reject an unknown currency before ever
  touching the category or user repos — assert the later fakes' call counts
  stayed zero, not just that an error came back.
- A workflow with only one service call and no explicit `uowManager`
  (`SignupWorkflow`) doesn't need `Restore`/`Finish` assertions — just prove
  it forwards to the service and surfaces that service's own errors.
- `CurrencyExchangeRateWorkflow.Sync` uses a hand-written `fakeRateSource`
  (one fake per port, like the repositories). Cover: the whole matrix lands
  through exactly one `UpsertMany` and opens **no** unit of work; a source
  error writes nothing; an `UpsertMany` error propagates; fewer than two
  currencies short-circuits; a skipped pair, an empty day, and an unconfigured
  code each add exactly one warning instead of being dropped silently; `Sync`
  forwards the supplied date unchanged and `SyncDaily` asks for yesterday in
  UTC.
