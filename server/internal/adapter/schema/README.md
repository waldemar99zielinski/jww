# internal/adapter/schema

**Clean Architecture layer: Interface Adapter (presentation/DTOs).** Converts between HTTP JSON and domain entities.

**Role.** The HTTP request/response JSON shapes and their hand-rolled validation. The only place that translates domain entities to/from JSON. Pure DTOs — no HTTP handling, no SQL, no business rules beyond field-shape checks.

## Files

One per resource (`auth.go`, `user.go`, `currency.go`, `currency_exchange_rate.go`, `transaction.go`, `transaction_attachment.go`, `transaction_category.go`, `summary.go`) plus `date.go` (`type Date time.Time`, `"YYYY-MM-DD"` JSON + `ParseDate`) and `validate.go` (helpers `minLength`, `exactLength`, `nonNegative`, `firstError`, `noDuplicateUUIDs`, `noNilUUIDs`).

- `transaction.go` — `TransactionCreateRequest` gains optional `attachment_ids []uuid.UUID` (validated for duplicates/empty ids); `TransactionResponse` gains `attachments []TransactionAttachmentResponse`, always non-nil, with no `attachment_count`.
- `transaction_attachment.go` — `TransactionAttachmentResponse` (public fields only; `transaction_id` is `*int64` and null for a draft) + `TransactionAttachmentToResponse`.

## Rules

- **Naming:** `<Entity>CreateRequest`, `<Entity>UpdateRequest`, `<Entity>Response`; converter is a free function `<Entity>ToResponse(d) <Entity>Response`. (`auth` deviates: `LoginRequest` / `TokenResponse` / `NewTokenResponse`.)
- **JSON tags:** always `snake_case`.
- Reuse **domain types** for fields: `domain.Money`, `domain.ExchangeRate`, `domain.<Enum>`, `schema.Date` for date-only, `uuid.UUID` for ids.
- **Validation is a value-receiver method** `func (r XRequest) Validate() error`, called by the handler before the service. No validation library, no struct tags for validation.
- Aggregate field errors with `firstError(...)` and return the first failure; messages are lowercase and name the JSON field.
- Responses **never** leak secrets (e.g. `UserResponse` omits `PasswordHash`). Flatten nested domain fields to scalars (`d.Currency.Code` → `currency_code`).

## How to add a schema

1. Add `<entity>.go`, `package schema`.
2. Define `<Entity>CreateRequest` (and `UpdateRequest` if needed) with snake_case JSON tags.
3. Implement `Validate()`: enum/range checks inline with `fmt.Errorf`; length checks via `firstError(minLength/exactLength/nonNegative, ...)`.
4. Define `<Entity>Response` + `<Entity>ToResponse`.
5. Query params are **not** DTOs — handlers parse them with `schema.ParseDate` / `uuid.Parse` / `domain.<Enum>.Valid()`.

## Testing requirement for this directory

**Location:** `tests/` subdirectory (black-box, imports this package's
exported API — see `server/TESTING.md#conventions`).

No I/O here, so every path is a fast unit test (`package schema_test`,
black-box):

- Every `Validate()` gets a table-driven test: one happy-path case, one
  case per violated invariant.
- Every `<Entity>ToResponse` converter gets a test asserting each exported
  response field actually came from the right domain field (a copy-paste
  mistake that silently drops or mismaps a field is exactly what this
  catches — a fake/mock can't).
- A response that must never leak a secret (`UserResponse` omitting
  `PasswordHash`) gets an explicit JSON-marshal assertion that the field
  name never appears, not just a struct-shape check.
- `schema.Date`: parse/format/JSON round-trip plus malformed-input cases.

See [`server/TESTING.md`](../../../TESTING.md).

## Avoid

- Re-implementing domain invariants beyond shape/range checks.
- Custom validation error structs or field-error maps.
- `time.Time` for date-only fields; `float64` for money.
- Exposing `PasswordHash`; inconsistent JSON casing / `omitempty`.
- Imports other than stdlib (`fmt`, `time`, `strings`), `uuid`, `domain` — no `net/http`, pgx, service, gateway, security, infrastructure.
