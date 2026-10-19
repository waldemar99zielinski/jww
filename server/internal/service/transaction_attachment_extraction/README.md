# internal/service/transaction_attachment_extraction

**Clean Architecture layer: Use Cases.** Defines the extraction ports; knows nothing about HTTP, SQL, S3, or any vision provider.

**Role.** The draft-only extraction use case: load one or more unattached image drafts, validate the complete group, mint one short-lived content grant per image, compose absolute content URLs, hand the ordered URLs to a provider-agnostic `Extractor` in one call, and map its combined structured result into a `Prefill`. Nothing is persisted and **no attachment bytes flow through the server's own request path** — the third-party provider fetches the bytes directly from the grant URLs.

The draft-only rule (missing or already-linked attachment → not found) is enforced entirely by the injected `AttachmentSource` port, which is satisfied structurally by [`transaction_attachment`](../transaction_attachment/README.md)'s draft lookup. This package never re-implements it.

## Files

- `extractor.go` — the provider-agnostic ports and result types a concrete adapter (for example [`../../adapter/openai`](../../adapter/openai)) implements:
  - `Extractor` — `ExtractTransactionPrefill(ctx, TransactionPrefillInput) (*TransactionPrefillResult, error)`. It receives ordered absolute content URLs plus the user's active category names split by type, never bytes.
  - `AttachmentSource` — `DraftAttachment(ctx, id)` (the draft-only lookup) and `ContentGrant(ctx, id)` (a short-lived, single-attachment capability token). Satisfied structurally by the `transaction_attachment` service or a small adapter wired in `cmd/server`; no `var _ Port` assertion.
  - `PrefillField[T]`, `PrefillSubTransaction`, `TransactionPrefillInput`, `TransactionPrefillResult` — the wire-agnostic extraction shape, mirroring the client's `TransactionPrefillDto`/`PrefillField<T>`. Confidence is an optional 0–100 score for every scalar and line item. `PrefillField`, `PrefillSubTransaction`, and `TransactionPrefillResult` implement `fmt.Stringer` so that logging the result (see `Extract`) prints one readable line — `TransactionPrefillResult{type=expense (confidence 90), amount=24.90, sub_transactions=[Milk 2% (dairy): 3.49]}` — in the **pretty** handler; the JSON handler ignores `String()` and marshals the struct fields (Go field names).
- `transaction_attachment_extraction.go` — `Prefill` (the service-level result) and `Service` + `NewService(extractor, source, config)` (extractor first, then source).
  - `ExtractMany(ctx, attachmentIDs)` rejects empty, duplicate, zero, or over-10 ID lists, validates every draft and image media type before minting any grant, then mints one grant and URL per ID in request order and calls `Extractor.ExtractTransactionPrefill` exactly once after loading active expense and income categories. Extractor failures become `domain.NewAttachmentExtractionFailedError(err)` (503); successful output maps to one `Prefill` containing all source IDs. `Extract(ctx, attachmentID)` remains a one-element compatibility wrapper.
- `config.go` — `Config{PublicBaseURL}` only (plain struct, `env:"PUBLIC_API_BASE_URL"`, required). No `LoadConfig` here: loading happens in `cmd/server`, keeping this package's imports within the service-layer allowlist.

## Rules

- **Naming:** package directories use `snake_case` with an underscore for multi-word names, and the Go package name matches the directory exactly — this package is `transaction_attachment_extraction` (not `transactionattachmentextraction`). Same convention as `transaction_attachment`, `currency_exchange_rate`, and `transaction_category`. See [`../README.md`](../README.md#shared-conventions).
- **Draft-only.** Extraction is only meaningful against an unattached draft, so the source port's `DraftAttachment` must delegate to the existing draft lookup (`GetDraft`) — a linked attachment is not found. The service does not attempt to detach, relink, or bypass that rule.
- **Images only.** The media check uses `mime.ParseMediaType` and accepts only `image/*`. This is deliberately stricter than `transaction_attachment.previewableMediaType`, which also permits `application/pdf`: PDF is out of scope for extraction v1 (file/document content parts are far less uniformly supported across OpenAI-compatible backends than image input), so a PDF draft is a 400 here even though it previews elsewhere. The check runs before minting a grant so an unsupported file never even reaches the provider.
- **The provider fetches the bytes, not us.** The extractor is handed absolute URLs to *our* content endpoint carrying short-lived per-image grants; the service never opens the object store, streams, or base64-encodes files. This keeps extraction on the same content path attachments already use and means the vision adapter never sees attachment keys or versions.
- **Absolute URL required.** `PublicBaseURL` must be the API's public origin — the provider resolves it from outside the cluster. In local development it must be a publicly reachable (typically tunneled) URL; `http://localhost` is not reachable by the provider and extraction will fail with a 503.
- **One error kind for every provider failure.** Timeout, transport error, non-2xx/rate-limit, content-filter refusal, and malformed/absent structured output all map to `AttachmentExtractionFailedError` (503) with a generic message; the underlying cause is available via `Unwrap` for boundary logging.
- Imports only `domain`, `unit_of_work` (when needed), stdlib (`fmt`, `mime`, `net/url`, `strings`, `time`), `uuid`, and `utils/logger` (**Debug tracing only** — see below). Never another service package, `adapter/*`, `infrastructure/*`, pgx, or `net/http`.
- **Extraction traces at `Debug`; the boundary owns the outcome.** The service logs counts and outcomes under component `service.transaction_attachment_extraction`, never grant tokens or content URLs. The HTTP boundary owns Info/Warn/Error summaries.

## Privacy

- **Third-party data exposure.** The receipt bytes leave the system boundary to the configured vision provider for every extraction call: the short-lived, single-attachment grant token is shared with that third party so it can fetch that one attachment. The grant is minted per call and scoped to one attachment; no credential, key, or other attachment is exposed. This is a user-visible privacy change worth surfacing in the product UI.
- The provider API key and `PublicBaseURL` are server-only config; the key never reaches the client, and the adapter never logs attachment bytes, filenames, or the raw model response.

## Avoid

- Logging above `Debug`, SQL, HTTP, or provider-specific types. Never log the grant token or the composed content URL.
- Opening/storing bytes here, allowing PDF (or any non-image) through, or checking media type *after* minting a grant or calling the extractor.
- Re-implementing the draft-only lookup instead of delegating to the source port.
- Reaching into another service package, `cmd/server`'s config loader, or a concrete adapter — the extractor and source stay injected ports.

## Testing requirement for this directory

**Location:** `tests/` subdirectory (black-box, imports this package's exported API — see [`server/TESTING.md#conventions`](../../../TESTING.md#conventions)).

Black-box (`package transaction_attachment_extraction_test`), testify, hand-written fake `Extractor` + fake `AttachmentSource` (no mocking library) — see [`server/TESTING.md`](../../../TESTING.md). Cover list validation, full-group validation before grants, ordered per-image URLs in one extractor call, non-image rejection, draft/grant/provider failures, and mapping the one combined result into `Prefill`.
