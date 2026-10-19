# routers/transaction_attachment

**Role.** The standalone attachment resource: upload, get, download, grant issuance, and delete for **unattached drafts**. Every draft ID operation requires `transaction_id IS NULL`, so a linked attachment is a 404 here and must be reached through its transaction. It also owns the public, grant-authenticated content proxy.

## Routes

| Access | Method | Path | Handler |
|---|---|---|---|
| protected | POST | `/v1/transaction-attachments` | `uploadDraftHandler` |
| protected | GET | `/v1/transaction-attachments/{attachmentID}` | `getDraftHandler` |
| protected | GET | `/v1/transaction-attachments/{attachmentID}/download` | `downloadDraftHandler` |
| protected | POST | `/v1/transaction-attachments/{attachmentID}/grants` | `issueDraftGrantHandler` |
| protected | POST | `/v1/transaction-attachments/{attachmentID}/extraction` | `extractionHandler` |
| protected | POST | `/v1/transaction-attachments/extraction` | `batchExtractionHandler` |
| protected | DELETE | `/v1/transaction-attachments/{attachmentID}` | `deleteDraftHandler` |
| grant | GET | `/v1/attachment-content?grant=...` | `grantedContentHandler` |

The public route `GET /v1/attachment-content?grant=...` validates the short-lived grant and proxies the exact object version with inline or attachment disposition. It supports one byte range and returns `206`/`416` as appropriate.

`Mount(public, protected, deps)`. The nested, transaction-scoped attachment routes live in the sibling
[`transaction`](../transaction/README.md) router instead.

## Notes

- Attachment IDs use `httputils.ParseUUIDPathParam`.
- Upload is **streaming multipart/form-data** via `kernel.ReadSingleMultipartFile`: exactly one file part named `file`. `Finish()` is called after the bytes are stored to prove no extra part followed; because the bytes are already written, a `Finish()` failure compensates by deleting the just-created draft before returning `400`.
- Upload clears the inherited read deadline (`http.NewResponseController(w).SetReadDeadline(time.Time{})`) after auth/validation; download clears the write deadline before streaming; extraction clears the write deadline after request validation because a vision call may run for minutes.
- Download is the one documented exception to the JSON envelope: bytes on success, JSON errors only before streaming (`kernel.WriteDownload`).
- Grant issuance accepts `{"purpose":"preview"}` or `{"purpose":"download"}` and returns the grant token, not a URL. Preview grants are limited to safe raster image and PDF media types. Granted responses are private/no-store and never expose the S3-compatible endpoint or key.
- The route is exempt from the global body cap; see `kernel.LimitRequestBody`.
- **Only extraction logs above the middleware.** This package declares `log = logger.New().WithComponent("http_server.transaction_attachment")` and `extractionHandler` uses it: `Info "attachment extraction completed"` (attachment id, duration, sub-transaction/warning counts) on success, and `Warn`/`Error` (attachment id, duration, error type, unwrapped cause) on failure — `Warn` for client-caused kinds (not found, conflict, validation, unauthorized), `Error` for the rest. The CRUD routes log nothing of their own; `kernel`'s `RequestLogger` already covers them. The service's own `Debug` step trace is documented in its [README](../../../service/transaction_attachment_extraction/README.md).

## Testing requirement for this directory

**Location:** `tests/` subdirectory (black-box) — see `server/TESTING.md`.

`httptest` + testify, a real `transaction_attachment.Service` built from
local hand-written fake `Repository`/`object_storage.Store`. Needs: draft upload →
201 with `transaction_id: null`; a second file part → 400 with the draft
compensated away; storage failure → 503; get/download/delete happy paths and
exact bytes; a linked attachment → 404 on every ID route; malformed id → 400.
