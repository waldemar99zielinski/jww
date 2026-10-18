# internal/service/transaction_attachment

**Clean Architecture layer: Use Cases.** Defines the attachment ports; knows nothing about HTTP, SQL, or S3.

**Role.** The attachment use cases: upload a file as an unattached draft (or straight onto a transaction), open/download, list, delete, and the lock-and-link operations a workflow uses to attach drafts atomically with transaction creation. Bytes live in blob storage behind the shared [`object_storage.Store`](../object_storage/README.md) port; metadata lives behind the `Repository` port.

## Files

- `repository.go` — metadata `Repository` port: `Create`, identity-only `GetByID`, filtered `Find`, `LockByIDs`, `LinkToTransaction`, `ListByTransactionIDs`, `Delete`.
- `filter.go` — `Filter` query object carrying the required attachment id plus explicit `DraftOnly` or `TransactionID` constraints; the Postgres gateway applies those constraints in SQL.
- `transaction_attachment.go` — storage namespace, orphan reporter, and service.
  - `object_storage.Store` (from [`../object_storage`](../object_storage/README.md)) — the shared, purpose-agnostic blob port this service consumes; it is not defined here.
  - `object_storage.StorageNamespace = "transaction-attachments"` is this purpose's folder inside the shared bucket. Wiring wraps the shared store with `object_storage.Namespaced(store, object_storage.StorageNamespace)`, so the service sees a store already scoped to that folder and uses plain UUID keys.
  - `Service` + `NewService(repository, storage, reportOrphan, grantTokens...)`.
  - `UploadDraft` / `UploadForTransaction` generate an opaque key (the attachment UUID), stream the body through a counting reader (bytes are measured, never buffered whole), and compensate by deleting the object if the metadata write fails. Because the injected store is namespaced, the physical path is `"<bucket>/transaction-attachments/<uuid>"` while `storage_key` (namespace-relative) is just the UUID.
  - `GetDraft`/`OpenDraft`/`DeleteDraft` operate on unattached drafts only; a linked attachment is not found. `GetForTransaction`/`OpenForTransaction`/`DeleteForTransaction` authorize through the parent.
  - `LockByIDs`/`Link`/`ListByTransactionIDs` are hooks a workflow calls inside its own unit of work. `DeleteAllForTransaction` deletes each object's bytes and metadata and forwards the workflow UoW so all metadata rows commit with the parent transaction deletion.
  - `IssueDraftGrant` issues a purpose-bound capability after draft-scoped authorization. `IssuePreviewGrants` batches preview capabilities for attachments already authorized by the transaction workflow. A preview is offered for `image/jpeg|png|webp|gif|avif|heic|heif|dng` (and `image/x-adobe-dng`) plus `application/pdf`; DNG and (outside Safari) HEIC are not renderable by browsers, so the client degrades to a download affordance when the inline image fails. `ResolveGrant` rechecks current metadata and object version; `OpenGranted` streams a full object or validated byte range through the server.

## Rules

- **Naming:** package directories use `snake_case` with an underscore for multi-word names, and the Go package name matches the directory exactly — this package is `transaction_attachment` (not `transactionattachment`). Same convention as `currency_exchange_rate` and `transaction_category`. See [`../README.md`](../README.md#shared-conventions).
- **Namespace via the shared store, not the service.** The base `Store` is folder-agnostic; each purpose owns a namespace constant and is scoped with `object_storage.Namespaced` at wiring. Don't bake the folder into keys here.
- Imports only `domain`, `unit_of_work`, `object_storage`, stdlib, and `uuid` — never the AWS SDK, pgx, `net/http`, or `utils/logger`.
- **The service never logs.** `reportOrphan` (`OrphanReporter`) is an injected callback so the boundary (`cmd/server`) can log one high-severity orphan event; it must never receive the object key or filename.
- **Object first, metadata second.** Upload streams to storage, then inserts metadata; a metadata failure deletes the just-written object. Delete removes the exact stored version before the row.
- Media types are never allowlisted and byte size is never capped (a deliberate product decision — see the design's risks).
- A missing object with a live metadata row is an `AttachmentIntegrityError` (500), not a 404.

## Avoid

- Logging, AWS SDK types, SQL, or HTTP.
- Reordering object/metadata operations (bytes must never be orphaned silently).

## Testing requirement for this directory

**Location:** `tests/` subdirectory (black-box, imports this package's
exported API — see `server/TESTING.md#conventions`).

Black-box (`package transaction_attachment_test`), testify, hand-written fake
`Repository` + `object_storage.Store` (no mocking library) — see
`server/TESTING.md`. Needs: upload success with an exact measured byte count
and an opaque key; storage failure → `ErrorKindUnavailable` with no metadata;
metadata failure → object deleted (compensation) and, when that also fails,
exactly one orphan report; a linked attachment invisible to the standalone
get/open/delete; a missing object → `ErrorKindInternal`; delete ordering
(object before row) and row preservation on a storage failure; parent/child
mismatch → not found.
