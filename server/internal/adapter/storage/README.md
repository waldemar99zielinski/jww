# internal/adapter/storage

**Clean Architecture layer: Interface Adapter (outbound/driven).** Implements the shared `object_storage.Store` port against private S3-compatible object storage.

**Role.** The **only** place AWS SDK types appear. It is deliberately **purpose-agnostic**: it adds no namespace and knows nothing about attachments (or any other caller), so one `S3Store` can back any number of features that pass their own object keys. It streams bytes to and from a private bucket and returns `object_storage` result structs; everything inward of it deals in `io.Reader`/`io.ReadCloser`.

## Files

- `config.go` — `Config` (bucket/region/endpoint/path-style/credentials) + `LoadConfig()`. One connection shared by every purpose; no per-feature prefix is baked in. `Validate()` fails fast on a half-configured static credential pair or a malformed endpoint.
- `s3_store.go` — `S3Store` + `NewS3Store(ctx, cfg)`.
  - Uses **AWS SDK Go v2** with `feature/s3/transfermanager` for streaming uploads (single `PutObject` below the multipart threshold, otherwise a multipart upload; concurrency 1–2, so memory ≈ part size × concurrency) and the raw S3 client for streaming downloads and exact-version deletes.
  - `BaseEndpoint` + `UsePathStyle` make it portable to MinIO and other providers.
  - Checksum calculation/validation is set to "when required" so third-party stores are not sent algorithms they may not implement.
  - Keys are used **verbatim**; the client adds no namespace. Purposes scope the shared store with [`object_storage.Namespaced`](../../service/object_storage/README.md) at wiring time (e.g. transaction attachments to `transaction-attachments/`).
  - `Delete` removes the exact stored `VersionID` (a plain delete only writes a delete marker on a versioned bucket) and is idempotent when the object/version is absent.
  - `isObjectNotFound` maps `NoSuchKey`/`NotFound` to `object_storage.ErrNotFound`.

## Rules

- Implements the existing port ([`internal/service/object_storage`](../../service/object_storage/README.md)); never defines one. Add `var _` assertions only if the repo starts doing so elsewhere (it doesn't).
- **Never creates the bucket.** Deployment (or local init) owns bucket creation and the private policy; the server fails fast on invalid config instead.
- **No logging** and no business logic — return errors; the boundary logs them once.
- No byte-size limit is configured here either; each purpose decides its own policy.

## Avoid

- Importing another adapter subpackage, `internal/infrastructure/*`, or `cmd`.
- Logging credentials, object keys, versions, or filenames.
- Baking a single feature's folder into the generic client.
- `float` sizes; string-interpolated keys.

## Testing requirement for this directory

**Location:** `tests/` subdirectory (black-box, imports this package's
exported API — see `server/TESTING.md#conventions`).

Black-box (`package storage_test`), **integration** (`//go:build
integration`), a real MinIO container via
[`internal/tests/s3test`](../../tests/s3test/README.md) — never a mocked S3
client. One container per test binary (shared `TestMain`). Run with
`go test -tags=integration ./internal/adapter/storage/...` (needs Docker).
Set `JWW_TEST_MINIO_IMAGE` to a mirror if the default `pgsty/minio` tag is
unreachable. Needs: a binary round trip; a zero-byte object; an arbitrary
content type; idempotent delete; a missing object mapping to
`ErrNotFound`; and a cancelled context failing.
