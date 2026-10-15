# routers/transaction

**Role.** Transaction CRUD and filtered listing (protected only).

## Routes (all `protected`)

| Method | Path | Handler |
|---|---|---|
| POST | `/v1/transactions` | `createTransactionHandler` |
| GET | `/v1/transactions` | `listTransactionsHandler` (filters) |
| GET | `/v1/transactions/{transactionID}` | `getTransactionHandler` |
| DELETE | `/v1/transactions/{transactionID}` | `deleteTransactionHandler` |

`Mount(_, protected, deps)`.

## Notes

- **Transaction IDs are `int64`** (Postgres `bigint`): parse with `strconv.ParseInt(chi.URLParam(r, "transactionID"), 10, 64)`, **not** `uuid.Parse`.
- Create decodes `schema.TransactionCreateRequest` then calls `deps.TransactionWorkflow.Create` (the workflow resolves currency/category/user).
- `listTransactionsHandler` builds `workflow.TransactionListFilter`, validating enums with `domain.TransactionType.Valid()`/`domain.TransactionSource.Valid()` and using shared sentinels `kernel.ErrInvalidType`/`kernel.ErrInvalidSource`; parses `created_from`/`created_to` with `time.RFC3339`.
- This is the only router that imports `internal/service/workflow` (for the filter type).

## Testing requirement for this directory

**Location:** `tests/` subdirectory (black-box, imports this package's
exported API — see `server/TESTING.md#conventions`).

`httptest` + testify, a real `TransactionWorkflow`/`TransactionService`
built from local hand-written fake repositories (no real Postgres) — see
`server/TESTING.md`. Needs: create happy path and validation-failure 400;
get happy path and not-found 404; delete happy path and not-found 404; list
with each query-param family (`type`/`source` invalid values → 400,
`created_from`/`created_to` malformed → 400). The workflow's own
currency/category/user resolution order is its own test suite's job, not
this one's.
