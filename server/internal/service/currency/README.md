# internal/service/currency

**Role.** CRUD use cases for the `Currency` entity. Owns transactions for multi-step `Update`/`Delete`.

## Files

- `repository.go` — `Repository`: `Create`, `GetByCode`, `List`, `Update`, `Delete` (all `ctx`, `uow...`).
- `currency.go` — `Service{ repository, uowManager }`; `NewService(repo, uowManager)`.
- `Create(ctx, code, name string, isBase bool, uow ...)` — build `domain.Currency`, `Validate()`, repo `Create`, return `&currency`.
- `GetByCode(ctx, code, uow ...)` — nil → `domain.NewCurrencyNotFoundError(code)`.
- `List(ctx, uow ...)` — passthrough.
- `Update(ctx, currency)` / `Delete(ctx, currency)` — own-UoW template (existence check inside, `Restore` on error, `Finish` on success); **no `uow` param**.

## Rules

- Follow the shared service conventions ([parent README](../README.md)): `ctx` first, `uow...` last, defaults set in the service, `Validate()` before I/O.
- `Currency` is keyed by its natural key `code` (3-letter ISO), not a surrogate id.

## Avoid

- Cross-entity rules (e.g. "one base currency" invariants that need other rows) — those belong in a workflow or the DB.
- Importing sibling services. Imports are `domain` + `unit_of_work` only.

## Testing requirement for this directory

**Location:** `tests/` subdirectory (black-box, imports this package's
exported API — see `server/TESTING.md#conventions`).

Black-box (`package currency_test`), testify, a local hand-written fake
`Repository` + `unit_of_work.Manager`/`UnitOfWork` — see `server/TESTING.md`
and [`internal/service/README.md`](../README.md#testing). Every method needs
a happy-path test; `Create`/`Update` need a validation-failure case proving
the fake repo was never called; `GetByCode` needs the `(nil, nil)` →
`domain.NewCurrencyNotFoundError` translation asserted via `errors.As`;
`Update`/`Delete` need `Restore`-on-every-error-path and
`Finish`-exactly-once-on-success assertions.

## Known gap

`currencies_single_base_uq` (migration `00003`) only enforces **at most** one
base currency; "at least one must exist" is documented there as an
application-layer invariant, but nothing in this service (or elsewhere)
currently enforces it — `Update`/`Delete` will happily leave zero base
currencies. If that invariant matters, it belongs in this service's
`Update`/`Delete` (check inside the existing transaction before committing).
