# routers/transaction_attachment

**Role.** The standalone attachment resource: upload, get, download, and delete for **unattached drafts** (protected only). Every ID operation requires `transaction_id IS NULL`, so a linked attachment is a 404 here and must be reached through its transaction.

## Routes (all `protected`)

| Method | Path | Handler |
|---|---|---|
| POST | `/v1/transaction-attachments` | `uploadDraftHandler` |
| GET | `/v1/transaction-attachments/{attachmentID}` | `getDraftHandler` |
| GET | `/v1/transaction-attachments/{attachmentID}/content` | `downloadDraftHandler` |
| DELETE | `/v1/transaction-attachments/{attachmentID}` | `deleteDraftHandler` |

`Mount(_, protected, deps)`. The nested, transaction-scoped attachment routes live in the sibling
[`transaction`](../transaction/README.md) router instead.

## Notes

- Attachment IDs are UUIDs: parse with `uuid.Parse(chi.URLParam(r, "attachmentID"))`.
- Upload is **streaming multipart/form-data** via `kernel.ReadSingleMultipartFile`: exactly one file part named `file`. `Finish()` is called after the bytes are stored to prove no extra part followed; because the bytes are already written, a `Finish()` failure compensates by deleting the just-created draft before returning `400`.
- Upload clears the inherited read deadline (`http.NewResponseController(w).SetReadDeadline(time.Time{})`) after auth/validation; download clears the write deadline before streaming.
- Download is the one documented exception to the JSON envelope: bytes on success, JSON errors only before streaming (`kernel.WriteDownload`).
- The route is exempt from the global body cap; see `kernel.LimitRequestBody`.

## Testing requirement for this directory

**Location:** `tests/` subdirectory (black-box) — see `server/TESTING.md`.

`httptest` + testify, a real `transaction_attachment.Service` built from
local hand-written fake `Repository`/`object_storage.Store`. Needs: draft upload →
201 with `transaction_id: null`; a second file part → 400 with the draft
compensated away; storage failure → 503; get/download/delete happy paths and
exact bytes; a linked attachment → 404 on every ID route; malformed id → 400.
