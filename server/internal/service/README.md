# internal/service

**Clean Architecture layer: Use Cases.** Defines the ports (interfaces) the app needs; knows nothing about HTTP or SQL.

**Role.** Use cases. Each resource subpackage is one module that defines **its own** repository port interface, scoped to exactly what it needs. `workflow/` composes several services for multi-service operations. Nothing here knows HTTP, SQL, or pgx.

## Packages

- [`auth/`](auth/README.md) — login + token verification.
- [`currency/`](currency/README.md), [`currency_exchange_rate/`](currency_exchange_rate/README.md), [`transactioncategory/`](transactioncategory/README.md), [`transaction/`](transaction/README.md), [`user/`](user/README.md) — CRUD-style use cases.
- [`unit_of_work/`](unit_of_work/README.md) — the persistence-agnostic transaction port.
- [`workflow/`](workflow/README.md) — orchestration across services; owns transaction demarcation.

## Imports

- A resource service may import **only** `internal/domain` + `internal/service/unit_of_work` (+ stdlib, `uuid`, `time`).
- **Forbidden:** other service packages, `adapter/*`, `infrastructure/*`, pgx, `net/http`, `utils/logger`. Composing services is the job of `workflow/`.
- **Services never log.** Return a typed error (see `domain`'s `DomainError`); the HTTP/CLI boundary that called you already logs with full request/process context. See `utils/logger`'s "Where to log".

## Shared conventions

- **Ports live here, not in adapters.** One concrete Postgres gateway satisfies several ports implicitly (structural typing) — no `var _ Port` assertions.
- **Port signature:** `ctx domain.AppContext` first; `uow ...unit_of_work.UnitOfWork` **last and variadic on every method**.
- `Service` struct + `NewService(...)`; include `uowManager unit_of_work.Manager` only when the service owns transactions (multi-step Update/Delete).
- Build the domain struct in the service, set defaults (`uuid.New()`, `time.Now().UTC()`, `Active: true`), call `Validate()` **before** I/O, then call the repo.
- Translate repo `(nil, nil)` into `domain.New...NotFoundError(...)`.
- Services are process-lifetime singletons, constructed once in `cmd/server/deps.go` and exposed via `kernel.Deps`.

## Delegating vs owning a transaction

- **`Create` / `Get*` / `List`:** accept `uow ...` and forward it — the caller (or the service's own caller) owns the transaction.
- **`Update` / `Delete`:** signature has **no `uow` param**; create their own via `uowManager`, `Restore` on every error, `Finish` on success, with the existence check inside the transaction.

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

- [ ] Minimal `<X>Repository` interface in the package (don't share across services).
- [ ] `ctx domain.AppContext` first, `uow ...unit_of_work.UnitOfWork` last on every method.
- [ ] `Service` + `NewService(...)`; add `uowManager` only if it owns transactions.
- [ ] Build entity, set defaults, `Validate()`, call repo.
- [ ] `(nil, nil)` → `domain.New...NotFoundError`.
- [ ] Update/Delete use the own-UoW template.
- [ ] Wire in `cmd/server/deps.go` and `kernel.Deps`.
- [ ] Add a `<name>_test.go` following this directory's testing requirement.
