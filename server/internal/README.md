# internal

**Role.** Parent directory; **no `.go` files of its own**. Holds the four Clean Architecture layers. The **Dependency Rule** applies: source dependencies point inward only — nothing in an inner layer may import an outer layer, directly or transitively.

```
Frameworks & Drivers   infrastructure/{database,http_server}   (outermost)
Interface Adapters     adapter/*
Use Cases              service/*
Entities               domain                                (innermost)
```

| Layer | Directories | Contains | Imports allowed |
|---|---|---|---|
| **Entities** | [`domain/`](domain/README.md) | structs, `Validate()`, `Money`/`ExchangeRate`, typed errors, enums, `AppContext` | stdlib + `uuid` only |
| **Use Cases** | [`service/`](service/README.md) | per-resource operations + their repository **ports**; `workflow/` orchestration + transaction demarcation | `domain`, `service/unit_of_work` (workflow also composes sibling services) |
| **Interface Adapters** | [`adapter/`](adapter/README.md) | Postgres repositories (`gateway`), HTTP DTOs (`schema`), security primitives (`security`) | `domain`, the service port it implements, third-party libs |
| **Frameworks & Drivers** | [`infrastructure/`](infrastructure/README.md) | DB pool + concrete UoW, HTTP server assembly + kernel/routers | services, adapters, `utils/*` |

## Rules

- Never import "upwards" (e.g. a service importing an adapter or `cmd`).
- Never import `cmd` from anywhere. `cmd/server` is the only composition root.
- **Ports are declared inward, implemented outward:** `service/*` declares `Repository`/`PasswordHasher`/`TokenService`/`Manager`; `adapter/*` and `infrastructure/database` implement them.
- Shared cross-cutting helpers live in `utils/`, not here.

## Testing requirement for this directory

No tests of its own — a parent directory with no code. Each layer's own
README has a "Testing requirement" section with that layer's bounded rule;
the full strategy is `server/TESTING.md`.
