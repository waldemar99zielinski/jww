# routers/currency

**Role.** Currency CRUD (protected only).

## Routes (all `protected`)

| Method | Path | Handler |
|---|---|---|
| POST | `/v1/currencies` | `createCurrencyHandler` |
| GET | `/v1/currencies` | `listCurrenciesHandler` |
| GET | `/v1/currencies/{code}` | `getCurrencyHandler` |
| PUT | `/v1/currencies/{code}` | `updateCurrencyHandler` |
| DELETE | `/v1/currencies/{code}` | `deleteCurrencyHandler` |

`Mount(_, protected, deps)` — `public` is unused.

## Notes

- Create/update decode `schema.CurrencyCreateRequest`/`CurrencyUpdateRequest` + `Validate`.
- Get/delete read `chi.URLParam(r, "code")`.
- **Update** fetches the existing entity, copies it, mutates allowed fields (`Name`, `IsBase`), then calls `Update` — never build a fresh entity from the request.
- Delete does a `GetByCode` then `Delete`; returns `204` via `kernel.WriteNoContent`.

## Testing requirement for this directory

**Location:** `tests/` subdirectory (black-box, imports this package's
exported API — see `server/TESTING.md#conventions`).

`httptest` + testify, `currency.NewService` built from a local hand-written
fake `Repository` (no real Postgres) — see `server/TESTING.md`. Every
handler needs its happy path plus its specific error path: bad JSON/failed
`Validate()` → 400 (create/update), not-found → 404 (get/update/delete).
Generic error-mapping is `kernel/respond_test.go`'s job, not this one's.
