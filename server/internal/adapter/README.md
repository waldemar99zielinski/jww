# internal/adapter

**Clean Architecture layer: Interface Adapters.** Implements inbound/outbound ports; sits between use cases and frameworks.

**Role.** Parent directory; **no `.go` files of its own**. Adapters are the driven/secondary side: they implement ports that are defined in `internal/service`. Five subpackages:

- [`gateway/`](gateway/README.md) — Postgres repositories (pgx + Squirrel).
- [`storage/`](storage/README.md) — private S3-compatible object storage (AWS SDK v2) implementing the shared `object_storage.Store` port.
- [`schema/`](schema/README.md) — HTTP request/response JSON shapes + validation.
- [`security/`](security/README.md) — Argon2 password hasher, JWT token service, `security.Config`.
- [`frankfurter/`](frankfurter/README.md) — exchange-rate provider client (`currency_exchange_rate.RateSource` over `net/http`).

## Rules

- **Ports live in `internal/service`, not here.** A new adapter implements an interface that already exists in the service layer; it does not define the port.
- Implicit (structural) interface satisfaction — no `var _ Port = (*Impl)(nil)` assertions.
- Adapters may import `internal/domain`, the service package whose port they implement, `service/unit_of_work`, and third-party libs.
- **Forbidden:** importing another adapter subpackage, `internal/infrastructure/*`, or `cmd`.
- Do not add files directly under `internal/adapter/`; put them in a subpackage (or a new subpackage for a genuinely new technology).
- **Adapters don't log**, by convention (importing `utils/logger` here wouldn't cycle, but do it anyway and you get the same failure logged twice). Return the error and let the one caller at the HTTP/CLI boundary log it. See `utils/logger`'s "Where to log".
- **Test style differs by subpackage's need for I/O**: `schema`/`security` are pure (no I/O) and get black-box unit tests (`package x_test`, no mocking); `gateway` genuinely needs Postgres and gets black-box integration tests instead — never a mocked `DBTX`/`pgx.Tx`. See each subpackage's own "Testing requirement" section and [`server/TESTING.md`](../../TESTING.md).
