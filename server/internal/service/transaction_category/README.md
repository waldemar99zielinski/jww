# internal/service/transactioncategory

**Role.** CRUD use cases for transaction categories. Owns transactions for `Update`/`Delete`.

## Files

- `repository.go` — `Repository`: `Create`, `GetByID(ctx, uuid.UUID, uow...)`, `ListByType(ctx, type domain.TransactionCategoryType, activeOnly bool, uow...)`, `Update`, `Delete`.
- `transaction_category.go` — `Service{ repository, uowManager }`; `NewService`.
- `Create(ctx, categoryType domain.TransactionCategoryType, name string, sortOrder int, uow...)` — assigns `ID: uuid.New()`, `Active: true` (defaults belong in the **service**, not domain/handler), `Validate()`, repo create.
- `GetByID` — nil → `domain.NewTransactionCategoryNotFoundError(id)`.
- `ListByType(ctx, type, activeOnly, uow...)` — passthrough; `activeOnly` is a plain bool argument (no filter struct).
- `Update`/`Delete` — own-UoW template, no `uow` param.

## Rules

- Follow the shared service conventions ([parent README](../README.md)).
- `TransactionCategory` uses `uuid.UUID` ids.

## Avoid

- Moving `uuid.New()` / `Active: true` defaults into the domain or handlers.
- Importing sibling services (imports: `domain`, `unit_of_work`, `uuid`).

## Testing requirement for this directory

**Location:** `tests/` subdirectory (black-box, imports this package's
exported API — see `server/TESTING.md#conventions`).

Black-box (`package transactioncategory_test`), testify, a local
hand-written fake `Repository` + `unit_of_work.Manager`/`UnitOfWork` — see
`server/TESTING.md`. Needs: `Create` sets `Active: true`/a fresh `ID` and
never calls the repo on a validation failure; `GetByID`'s `(nil, nil)` →
`NotFoundError`; `ListByType` filters by both type and `activeOnly`;
`Update`/`Delete` assert `Restore`-on-every-error-path and
`Finish`-exactly-once-on-success.
