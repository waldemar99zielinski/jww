# internal/domain

**Clean Architecture layer: Entities (innermost).** Imports nothing internal.

**Role.** The dependency-free core: plain data structs, per-entity invariants (`Validate() error`), the `Money`/`ExchangeRate` fixed-point value types, typed domain errors, enums, and the request-scoped `AppContext`. No persistence, no HTTP, no logging.

## Imports

- **Allowed:** stdlib (`context`, `fmt`, `strings`, `time`, `math/bits`, `strconv`), `github.com/google/uuid`.
- **Forbidden:** any other `jww/internal/*`, `net/http`, pgx, logger.

## Key files

- `app_context.go` — `AppContext{ctx,user,requestSource,flowID}` (unexported fields; value-receiver getters, pointer-receiver setters). Carried **by value** everywhere. Built at the HTTP edge by `kernel.RequestContext`.
- `errors.go` — `DomainError` marker interface (unexported `isDomainError()` ⇒ only this package can implement it) exposing `Kind() ErrorKind` (`NotFound`/`Conflict`/`Validation`/`Unauthorized`/`Internal`/`Unavailable`) so the HTTP kernel can map an error to a status without the domain knowing HTTP. Concrete `...NotFoundError` types + `New...NotFoundError(...)`, plus `AuthenticationError` (one message for bad login *and* bad token — anti-enumeration), `AttachmentAlreadyLinkedError` (conflict), `ValidationError` (a cross-item rule the entity can't express), `ObjectStorageUnavailableError` (`Unavailable` → 503, generic message, cause via `Unwrap`), and `AttachmentIntegrityError` (a metadata row whose bytes are gone → 500).
- `money.go` — `type Money int64` (cents), `type ExchangeRate int64` (1e-4). `ExchangeRate.Convert(Money) Money` multiplies in 128 bits (`math/bits`), rounds half away from zero, panics on int64 overflow. Money JSON is a bare **integer cents** number (4127); decimal input is rejected. `ParseMoney` parses display decimals into cents; `ParseExchangeRate` parses rate decimals.
- `enums.go` — `TransactionCategoryType`, `TransactionType`, `TransactionSource`, `SummaryPeriod`; each a `string` with `Valid() bool`.
- Entities: `user.go`, `currency.go`, `currency_exchange_rate.go`, `currency_stats.go`, `transaction.go`, `transaction_attachment.go`, `transaction_category.go`, `summary.go` — each a plain struct with `Validate() error`. `TransactionAttachment.TransactionID` is a `*int64` that is nil while the attachment is an unattached draft; the pointer is the whole state machine.

## Rules

- Entities are **data holders**; construction/mutation lives in services. `Validate()` checks only the entity's own invariants and returns a plain `fmt.Errorf("...")` (fail-fast, first violation).
- "Lookup failed" is a typed `New...NotFoundError(...)`; validation failures are plain errors.
- Never use `float64` for money/rates. Use `Money`/`ExchangeRate`; convert only via `ExchangeRate.Convert`.
- Optionality is expressed with pointer fields (`*SummaryPeriod`, `*time.Time`, `*User`).
- Filenames are `snake_case` of the entity.
- `Validate()` should cover every business-rule `CHECK` constraint in `migrations/` for that table (e.g. `Currency.Code` must be uppercase, matching `currencies_code_uppercase`) — otherwise an invalid value sails through domain validation and only fails later as an opaque DB error. When adding/changing a validated field, check the matching migration for constraints not yet mirrored here.

## How to add a domain type

1. Add `<name>.go`, plain exported struct (fields = other domain types / stdlib / `uuid`).
2. Add `func (x X) Validate() error` returning `fmt.Errorf` on the first violated invariant.
3. Use `Money`/`ExchangeRate` for amounts; `uuid.UUID` for identity, or `int64` when it maps to a Postgres `bigint`.
4. If callers need a "not found" signal, add `XNotFoundError` + constructor here embedding `baseDomainError` with `kind: ErrorKindNotFound`. Pick the `ErrorKind` that matches the semantics (the kernel maps each kind to a status: validation→400, unauthorized→401, not found→404, conflict→409, internal→500).
5. Keep it behavior-free beyond `Validate`, value methods (`String`, JSON), enum `Valid()`.

## Conventions

- Every exported struct gets a one-line doc comment starting with its name (`// X is a ...`); if it has a non-obvious field (e.g. an ID that's a plain `int64` instead of `uuid.UUID`, or a field whose zero value is semantically meaningful), say so there instead of leaving it to `Validate()` or the reader to infer.
- Money/rate math worth more than a glance (overflow paths, rounding) gets a test, not just the happy path — see `money_test.go`'s overflow and half-away-from-zero cases.

## Testing requirement for this directory

**Location:** alongside the source, not a `tests/` subdirectory — these are
white-box (`package domain`) and reach unexported helpers like
`isUpperAlpha`, which a subdirectory package cannot see. See
`server/TESTING.md#conventions`.

- Every `Validate() error` gets a table-driven test (`testify`'s `assert`/`require`): one subtest per violated invariant, plus one happy-path subtest. See `currency_test.go`/`summary_test.go` for the pattern (the latter for a struct with interdependent fields that need more than one base fixture).
- Every enum's `Valid()` gets a test covering a valid value, an invalid value, and the empty string.
- Every new `New...NotFoundError` constructor is covered by `errors_test.go`'s table — assert `errors.As` finds a `DomainError` and `Kind()` is the kind you intended (not just `ErrorKindNotFound` by default).
- Money/rate math worth more than a glance (overflow paths, rounding) gets explicit coverage, not just the happy path — see `money_test.go`.
- Tests stay white-box (`package domain`, not `domain_test`) and use `testify` (`assert`/`require`), matching the existing suite.

## Avoid

- Repository interfaces / factories / persistence methods here.
- Returning typed `DomainError`s from `Validate`.
- Float money, or JSON-quoting money as a string.
- Exporting fields on `AppContext`.
- **Error kinds drive the status:** never return a typed `DomainError` whose `ErrorKind` doesn't match its meaning — `kernel.WriteError` maps every non-auth `DomainError` by its `Kind()`, and an unclassified one defaults to 404.
