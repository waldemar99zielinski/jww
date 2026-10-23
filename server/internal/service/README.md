# internal/service

**Clean Architecture layer: Use Cases.** Defines the ports (interfaces) the app needs; knows nothing about HTTP or SQL.

**Role.** Use cases. Each resource subpackage is one module that defines **its own** repository port interface, scoped to exactly what it needs. `workflow/` composes several services for multi-service operations. Nothing here knows HTTP, SQL, or pgx.

## Packages

- [`auth/`](auth/README.md) — login + token verification.
- [`currency/`](currency/README.md), [`currency_exchange_rate/`](currency_exchange_rate/README.md), [`transaction_category/`](transaction_category/README.md), [`transaction/`](transaction/README.md), [`transaction_attachment/`](transaction_attachment/README.md), [`user/`](user/README.md), [`webauthn_credential/`](webauthn_credential/README.md) — CRUD-style use cases.
- [`health/`](health/README.md) — concurrent required/optional dependency health aggregation for readiness.
- [`unit_of_work/`](unit_of_work/README.md) — the persistence-agnostic transaction port.
- [`object_storage/`](object_storage/README.md) — the persistence-agnostic blob-storage port (shared by any feature that stores files).
- [`workflow/`](workflow/README.md) — orchestration across services; owns transaction demarcation.

## Imports

- A resource service may import **only** `internal/domain` + `internal/service/unit_of_work` (+ stdlib, `uuid`, `time`).
- **Forbidden:** other service packages, `adapter/*`, `infrastructure/*`, pgx, `net/http`. Composing services is the job of `workflow/`.
- **Services never log**, with one documented exception: `service/health` owns dependency readiness diagnostics and imports `utils/logger` to WARN on unavailable/slow checks. Otherwise return a typed error (see `domain`'s `DomainError`); the HTTP/CLI boundary that called you already logs with full request/process context. See `utils/logger`'s "Where to log".

## Shared conventions

- **Directory/package naming:** a multi-word package directory uses `snake_case` with an underscore (`currency_exchange_rate`, `transaction_category`, `transaction_attachment`), and the Go package name matches the directory exactly (so the import path and the identifier you call are spelled the same way). Never concatenate the words (`transactionattachment`).
- **Ports live here, not in adapters.** One concrete Postgres gateway satisfies several ports implicitly (structural typing) — no `var _ Port` assertions.
- **Repository ports live in `repository.go`; filter/query objects live in `filter.go`.** Keep service behavior and construction in the resource-named service file.
- **Port signature:** `ctx domain.AppContext` first; `uow ...unit_of_work.UnitOfWork` **last and variadic on every method**.
- `Service` struct + `NewService(...)`; include `uowManager unit_of_work.Manager` only when the service owns transactions (multi-step Update/Delete).
- Build the domain struct in the service, set defaults (`uuid.New()`, `time.Now().UTC()`, `Active: true`), call `Validate()` **before** I/O, then call the repo.
- Translate repo `(nil, nil)` into `domain.New...NotFoundError(...)`.
- Services are process-lifetime singletons, constructed once in `cmd/server/deps.go` and exposed via `kernel.Deps`.

## Delegating vs owning a transaction

- **`Create` / `Get*` / `List`:** accept `uow ...` and forward it — the caller (or the service's own caller) owns the transaction.
- **`Update` / standalone `Delete`:** normally create their own UoW via `uowManager`, `Restore` on every error, and `Finish` on success, with the existence check inside the transaction. A delete that must compose with other resource changes may accept `uow ...`; when supplied, it forwards that UoW and leaves finish/restore to the workflow owner.

## Filter rules

- A resource owns exactly one filter type in its own `filter.go`; workflows and transport packages reuse it and never define mirror filter structs.
- Filter fields name the exact queried field. For related resources, flatten the key into an explicit scalar such as `UserID`, `CategoryID`, `FromCurrencyCode`, or `ToCurrencyCode`; never put a nested domain struct (`User`, `Currency`, etc.) in a filter.
- If several fields of one related resource become queryable, expose each separately (for example `CategoryID` and `CategoryName`) rather than introducing a `Category` field.
- Use pointers where the zero value is meaningful and must be distinguishable from “not supplied”; a blank string may represent “not supplied” when blank is not a valid value.
- Range bounds use explicit `...From` / `...To` names and inclusive SQL predicates unless the field documents otherwise. `Validate()` rejects inverted ranges and other cross-field contradictions.
- The owning service validates a non-nil filter before repository I/O. A validation failure must not call the repository.
- Repository implementations apply filters in the data-store query, never by loading all rows and filtering in memory. A nil filter means no filtering.
- Adding a filter field requires service validation/pass-through coverage and adapter integration coverage proving the generated query behavior.

## Testing requirement for this directory

No mocking library — a service's `Repository` port is already the minimal
interface this package needs, so a table-driven test implements it directly
as a small hand-written fake (a struct holding the fields/funcs the test
cases exercise) in that package's `_test.go` file, black-box
(`package <name>_test`), testify assert/require. Fake `PasswordHasher`/
`TokenService`/etc. the same way. This is also why ports stay scoped to one
service instead of being shared: a shared interface would force every test
to fake methods it doesn't use. This isn't just a convention — **every
method added or changed here needs test coverage following this pattern**,
including: a happy path; a validation failure proving the repo was never
called; the repo `(nil, nil)` → `NotFoundError` translation where
applicable; and, for `Update`/`Delete`, `Restore`-on-every-error-path and
`Finish`-exactly-once-on-success assertions via a local fake
`unit_of_work.Manager`/`UnitOfWork`. See each subpackage's own README and
`server/TESTING.md`.

## Add a new service — checklist

- [ ] Minimal `<X>Repository` interface in `repository.go` (don't share across services).
- [ ] Filter/query objects, when needed, in `filter.go`.
- [ ] `ctx domain.AppContext` first, `uow ...unit_of_work.UnitOfWork` last on every method.
- [ ] `Service` + `NewService(...)`; add `uowManager` only if it owns transactions.
- [ ] Build entity, set defaults, `Validate()`, call repo.
- [ ] `(nil, nil)` → `domain.New...NotFoundError`.
- [ ] Update/Delete use the own-UoW template.
- [ ] Wire in `cmd/server/deps.go` and `kernel.Deps`.
- [ ] Add a `<name>_test.go` following this directory's testing requirement.
