# internal/service/transaction

**Role.** CRUD/list use cases for transactions, plus the reusable `Filter` query object. Owns transactions for `Update`/`Delete`.

## Files

- `transaction.go` — `Repository`; `Service{ repository, uowManager }`; `NewService`.
  - `Create` takes **explicit scalar fields** (not a `Transaction`) because the DB generates `id`/`created_at`, and returns the persisted `*domain.Transaction`.
  - `GetByID(ctx, int64, uow...)` — nil → `domain.NewTransactionNotFoundError(id)`.
  - `List(ctx, *Filter, uow...)` — passthrough.
  - `Update`/`Delete` — own-UoW template, no `uow` param.
- `filter.go` — `Filter{ User, Category *domain.TransactionCategory, Type, Currency, Source, CreatedFrom, CreatedTo }` (all optional pointers) with `Validate()` covering range ordering and `Category.Type == Type`.

## Rules

- `Transaction.ID` is `int64` (Postgres `bigint`), not a UUID.
- `Filter` is a **service-layer query object** referencing already-resolved **domain structs**; workflows resolve raw ids/codes into these.
- `Create` builds a `Transaction` for `Validate()`, then calls the repo with `time.Now().UTC()` and scalar columns; returns the repo-created entity so generated `id`/`created_at` propagate.

## How to extend

- New filter field: add a pointer to `Filter`, handle it in the gateway, and extend `Filter.Validate()` for cross-field rules only.

## Avoid

- Resolving foreign keys here (workflow concern).
- Fetching related entities just to filter — the gateway JOINs.
- Imports beyond `domain`, `unit_of_work`, `time`, `fmt`.

## Testing requirement for this directory

**Location:** `tests/` subdirectory (black-box, imports this package's
exported API — see `server/TESTING.md#conventions`).

Black-box (`package transaction_test`), testify, a local hand-written fake
`Repository` + `unit_of_work.Manager`/`UnitOfWork` — see `server/TESTING.md`.
Needs: `Create`'s validation failure (e.g. `Category.Type != Type`) never
reaches the repo; `GetByID`'s `(nil, nil)` → `NotFoundError`; `List` passes
the `*Filter` through unchanged (`assert.Same`); `Update`/`Delete` assert
`Restore`-on-every-error-path and `Finish`-exactly-once-on-success.
