# internal/service/unit_of_work

**Role.** The persistence-agnostic transaction port shared by services and workflows. It knows nothing about Postgres; the concrete tx lives in `internal/infrastructure/database` and is consumed by gateways via `GetValue()`.

## API

```go
type Manager interface {
    Create(ctx domain.AppContext) (UnitOfWork, error)
}

type UnitOfWork interface {
    Finish(ctx domain.AppContext) error   // commit
    Restore(ctx domain.AppContext) error  // rollback
    GetValue() interface{}                // opaque tx handle for adapters
}
```

- Repository methods accept `uow ...UnitOfWork` **optionally**: present → use the underlying transaction; absent → use the default connection.
- `Manager` is a service dependency **only** when an operation spans more than one repository call.

## Rules

- Imports: `internal/domain` only.
- The gateway enforces "at most one uow" and type-asserts `GetValue().(DBTX)`.
- Transaction-demarcation owner pattern: `active, err := mgr.Create(ctx)`; pass `active` into every repo/service call; `_ = active.Restore(ctx)` on any error; `active.Finish(ctx)` once on success.

## Avoid

- Importing pgx or the database package here.
- A repository creating its own UoW.
- Passing more than one UoW.
- Doing business logic in anything consuming `GetValue()`.

## Testing requirement for this directory

No tests of its own — this package is two pure interfaces with no logic to
exercise. The concrete implementation (`internal/infrastructure/database`'s
`PostgresUnitOfWorkManager`) is covered indirectly by every gateway
integration test (each one begins a real transaction via
`pgtest.BeginRollback`, which exercises `Create`/`Restore` for real) and by
every service/workflow unit test's local fake `Manager`/`UnitOfWork`
(which exercises the Restore-on-error/Finish-on-success contract). If this
interface ever grows real logic, it needs its own test.
