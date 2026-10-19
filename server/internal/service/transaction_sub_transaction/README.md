# internal/service/transaction_sub_transaction

**Clean Architecture layer: Use Cases.** Defines the line-item port; knows nothing about HTTP, SQL, or extraction.

**Role.** The line-item use cases: create, update, and delete one `TransactionSubTransaction` row under its parent transaction, plus a batch `ListByTransactionIDs` used to hydrate a transaction's items on read. It owns only its own rows and their persistence port.

## Files

- `repository.go` — line-item `Repository` port: `Create`, `Update`, `Delete`, identity-only `GetByID`, and batch `ListByTransactionIDs`. Every method takes `uow ...unit_of_work.UnitOfWork` last and variadic, so a caller (normally the transaction workflow) can compose a line-item write with its parent transaction mutation atomically.
- `transaction_sub_transaction.go` — `Service` + `NewService(repository, uowManager)`.
  - `Create` builds the row (assigns a new `uuid.New()` id and `CreatedAt = time.Now().UTC()`), validates it, then persists. A validation failure returns before the repository is called.
  - `GetByID` is an identity-only lookup and translates a repository `(nil, nil)` into `domain.NewTransactionSubTransactionNotFoundError(id)`, mirroring [`transaction_attachment`](../transaction_attachment/transaction_attachment.go)'s draft lookup.
  - `Update` validates first — before any I/O, so an invalid update never reaches the repository or opens a unit of work — then owns a unit of work unless the caller supplied one: it creates its own via `uowManager`, runs the existence check (`GetByID`) inside it, `Restore`s on every error path (including not-found) and `Finish`es on success. A supplied `uow` is forwarded and the finish/restore stays with the caller/owner. A missing row is a not-found.
  - `Delete` follows the same own-unless-supplied template: without a `uow` it creates one, existence-checks inside, `Restore`s on every error path and `Finish`es on success; a supplied `uow` is forwarded. A missing row is a not-found, not a silent no-op.
  - `ListByTransactionIDs` short-circuits an empty id set with `(nil, nil)` and never calls the repository in that case.

## Rules

- **Naming:** package directories use `snake_case` with an underscore for multi-word names, and the Go package name matches the directory exactly — this package is `transaction_sub_transaction` (not `transactionsubtransaction`). Same convention as `transaction_attachment`, `currency_exchange_rate`, and `transaction_category`. See [`../README.md`](../README.md#shared-conventions).
- **Identity.** A line item is keyed by its own `uuid.UUID`; `TransactionID` is a plain `int64` (a Postgres bigint identity column), unlike the UUID-keyed entities. `GetByID`/`Delete`/`Update` look the row up by that id alone.
- **Amounts are non-zero, not positive.** Negative line items represent discounts/coupons; zero is rejected by the domain and the database. The parent transaction amount must still be positive.
- **Amount accumulation is NOT here.** This service never sums amounts, never reads or writes the parent `Transaction.Amount`, and does not import the transaction service. The rule "whenever a transaction has one or more line items, `Transaction.Amount` is derived as Σ(items)" is cross-entity, so it lives in [`workflow.TransactionWorkflow`](../workflow/README.md), which composes this service with the transaction repository inside one unit of work. Keeping it out of here is what keeps this package a pure single-row store.
- Imports only `domain`, `unit_of_work`, stdlib, and `uuid` — never the AWS SDK, pgx, `net/http`, or `utils/logger`.
- **The service never logs.** It returns a typed `domain.DomainError`; the boundary logs.
- Build the entity and set defaults, `Validate()` **before** any I/O, then call the repository.
- Translate a repository `(nil, nil)` into the not-found error.
- `Create`/`Get*`/`List` forward `uow`; `Update`/`Delete` own a unit of work unless one is supplied.

## Avoid

- Logging, SQL, or HTTP.
- Summing items, touching the parent transaction's `amount`, or any cross-entity logic — that is the workflow's job.
- Reaching into another service package (composing services is `workflow/`'s role).

## Testing requirement for this directory

**Location:** `tests/` subdirectory (black-box, imports this package's
exported API — see `server/TESTING.md#conventions`).

Black-box (`package transaction_sub_transaction_test`), testify, hand-written
fake `Repository` + a local fake `unit_of_work.Manager`/`UnitOfWork` (no
mocking library) — see `server/TESTING.md`. Cover: `Create` happy path with an
assigned non-zero id and `CreatedAt`, and a validation failure proving the
repository was never called; `GetByID` translating a repository `(nil, nil)`
into `*domain.TransactionSubTransactionNotFoundError` (kind
`ErrorKindNotFound`); `Update` validating before the existence check, then
`Restore`-on-every-error-path (missing row) and `Finish`-exactly-once-on-success
assertions; `Delete` 404-ing a missing row without deleting and its own
`Restore`/`Finish` behavior; and `ListByTransactionIDs` short-circuiting an
empty id set without a repository call.
