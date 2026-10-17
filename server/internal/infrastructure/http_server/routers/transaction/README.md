# routers/transaction

**Role.** Transaction CRUD and filtered listing, plus the transaction-scoped attachment routes (all protected).

## Routes (all `protected`)

| Method | Path | Handler |
|---|---|---|
| POST | `/v1/transactions` | `createTransactionHandler` |
| GET | `/v1/transactions` | `listTransactionsHandler` (filters) |
| GET | `/v1/transactions/{transactionID}` | `getTransactionHandler` |
| DELETE | `/v1/transactions/{transactionID}` | `deleteTransactionHandler` |
| POST | `/v1/transactions/{transactionID}/attachments` | `uploadAttachmentHandler` |
| PUT | `/v1/transactions/{transactionID}/attachments/{attachmentID}` | `attachAttachmentHandler` |
| GET | `/v1/transactions/{transactionID}/attachments` | `listAttachmentsHandler` |
| GET | `/v1/transactions/{transactionID}/attachments/{attachmentID}/content` | `downloadAttachmentHandler` |
| DELETE | `/v1/transactions/{transactionID}/attachments/{attachmentID}` | `deleteAttachmentHandler` |

`Mount(_, protected, deps)`. The standalone draft routes live in the sibling
[`transaction_attachment`](../transaction_attachment/README.md) router.

## Notes

- **Transaction IDs are `int64`** (Postgres `bigint`): parse with `parseTransactionID(r)` (`strconv.ParseInt`), **not** `uuid.Parse`. Attachment IDs are UUIDs.
- Create decodes `schema.TransactionCreateRequest` (which now carries an optional `attachment_ids`) then calls `deps.TransactionWorkflow.Create`, which links those drafts in the same UoW as the insert.
- `getTransactionHandler`/`deleteTransactionHandler` call `TransactionWorkflow.Get`/`Delete` (not `TransactionService` directly) so attachment summaries are hydrated and object cleanup cannot be bypassed.
- Nested attachment handlers require the attachment to belong to the path transaction; a mismatch is 404 (`GetForTransaction`/`DeleteForTransaction`/`OpenForTransaction`). `attachAttachmentHandler` is a 204 no-op for the same parent and a 409 for a different one.
- `listAttachmentsHandler` uses `TransactionWorkflow.Get` (which also proves the parent exists).
- Upload uses `kernel.ReadSingleMultipartFile` and is exempt from the global body cap; download streams via `kernel.WriteDownload`.
- `listTransactionsHandler` builds the shared `transaction.Filter`, validating enums with `domain.TransactionType.Valid()`/`domain.TransactionSource.Valid()` and using shared sentinels `kernel.ErrInvalidType`/`kernel.ErrInvalidSource`; parses `created_from`/`created_to` with `time.RFC3339`.
- With no query filters, list returns every transaction ordered by `created_at DESC`, then `id DESC` for deterministic ties. The same ordering applies to filtered results.
- This is the only router (besides the standalone attachment router) that imports `internal/service/workflow`.

## Testing requirement for this directory

**Location:** `tests/` subdirectory (black-box, imports this package's
exported API — see `server/TESTING.md#conventions`).

`httptest` + testify, real `TransactionWorkflow`/`TransactionService` and a
real `transaction_attachment.Service` built from local hand-written fake
repositories/storage (no real Postgres) — see `server/TESTING.md`. Needs:
create happy path and validation-failure 400; create with `attachment_ids`
links and returns a hydrated response; duplicate/unknown attachment ids →
400/404; attach-later 204, same-parent no-op, other-parent 409; nested upload
201; nested list/download/delete; parent mismatch 404; get/delete happy paths
and not-found 404; list query-param errors. The workflow's own resolution
order is its own test suite's job, not this one's.
