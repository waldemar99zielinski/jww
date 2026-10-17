# internal/tests/s3test

**Role.** Boots a disposable MinIO container for tests that need a real
S3-compatible object store — [`internal/adapter/storage`](../../adapter/storage/README.md)'s
integration tests and `server/tests/e2e`. Test-only: nothing outside a test
imports it, and it never ships in a production binary.

## API

- `Start(ctx, bucketName) (*Bucket, error)` — starts MinIO, creates
  `bucketName`, and returns a `Bucket` whose `Config` (`storage.Config`) points
  a real `storage.S3Store` at it. Call once per test binary.
- `(*Bucket) Close(ctx)` — terminates the container; call from `TestMain`
  after `m.Run()`.

## Notes

- The image is pinned (`pgsty/minio:RELEASE.…`, a community rebuild of MinIO —
  MinIO deleted its own `minio/minio` Docker Hub images in September 2026). Set
  `JWW_TEST_MINIO_IMAGE` to a mirror (e.g. an internal registry) when the
  default registry is unreachable.
- Deliberately independent of [`pgtest`](../pgtest/README.md): the two
  containers have different lifecycles, and a test that needs Postgres does not
  necessarily need object storage.

## Testing requirement for this directory

No tests of its own — it is a thin container-lifecycle helper consumed by the
storage integration suite and `server/tests/e2e`; a broken helper fails those
suites outright. See `server/TESTING.md`.
