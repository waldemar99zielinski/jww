# routers/transaction_category

**Role.** Transaction-category CRUD + filtered listing (protected only).

## Routes (all `protected`)

| Method | Path | Handler |
|---|---|---|
| POST | `/v1/transaction-categories` | `createCategoryHandler` |
| GET | `/v1/transaction-categories` | `listCategoriesHandler` (query: `type` required, `active_only` optional) |
| GET | `/v1/transaction-categories/{categoryID}` | `getCategoryHandler` |
| PUT | `/v1/transaction-categories/{categoryID}` | `updateCategoryHandler` |
| DELETE | `/v1/transaction-categories/{categoryID}` | `deleteCategoryHandler` |

`Mount(_, protected, deps)`.

## Notes

- `categoryID` is a UUID (`uuid.Parse`).
- Create/update decode `schema.TransactionCategoryCreateRequest`/`UpdateRequest` + `Validate`.
- `listCategoriesHandler`: `type` validated via `domain.TransactionCategoryType.Valid()` (uses shared `kernel.ErrInvalidType`); `active_only` parsed with `strconv.ParseBool` (default `true`); calls `deps.TransactionCategoryService.ListByType`.

## Testing requirement for this directory

**Location:** `tests/` subdirectory (black-box, imports this package's
exported API — see `server/TESTING.md#conventions`).

`httptest` + testify, `transactioncategory.NewService` built from a local
hand-written fake `Repository` (no real Postgres) — see `server/TESTING.md`.
Every handler needs its happy path plus its specific error path: bad
JSON/failed `Validate()` → 400 (create/update), not-found → 404
(get/update/delete), invalid `type` → 400 (list).
