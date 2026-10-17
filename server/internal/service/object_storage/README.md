# internal/service/object_storage

**Clean Architecture layer: Use Cases.** The persistence-agnostic outbound port for blob storage.

**Role.** Like [`service/unit_of_work`](../unit_of_work/README.md), this is a small, **shared, cross-cutting infrastructure port** rather than a resource-specific one: several independent purposes (transaction attachments today, others later) reuse one implementation and one connection. It knows nothing about S3, MinIO, or the filesystem — see [`internal/adapter/storage`](../../adapter/storage/README.md) for the S3-compatible implementation.

## Files

- `object_storage.go` —
  - `Store` — `Put(ctx, Ref, body io.Reader, mediaType) (Stored, error)`, `Open(ctx, Ref) (Stream, error)`, `Delete(ctx, Ref) error` (idempotent if absent).
  - `Ref{Key, Version}` — `Key` is a caller-defined, opaque path; the base `Store` adds no namespace of its own. `Version` is empty for a Put and the exact stored version for an Open/Delete.
  - `Stored{Version, ETag}`, `Stream{Body, ByteSize}`.
  - `ErrNotFound` — the wrapped sentinel an implementation returns for a missing object.
  - `Namespaced(store, folder) Store` — a decorator that scopes every key under `"<folder>/"`, so independent purposes share one underlying store/bucket while keeping their objects in separate folders. This is the mechanism for distinguishing purposes (each defines a folder constant and wrapping happens at wiring). Keys passed through a namespaced store are relative to that folder.

## Rules

- **The base `Store` imposes no namespace.** Stable namespace constants live in this package; apply `Namespaced(store, const)` where the store is injected. Transaction attachments use `StorageNamespace = "transaction-attachments"`. The same store implementation then serves every purpose unchanged.
- Implementations live in `internal/adapter/*`; no provider SDK type crosses this boundary.
- Deleting an absent object must succeed (idempotent) so callers can retry a partial cleanup.
- This port imports only `domain`, `errors`, `io`, and `strings`.

## Avoid

- Provider-specific types or connection settings here.
- A specific feature's folder baked into the base store — that belongs to the feature (via `Namespaced`).

## Testing requirement for this directory

**Location:** alongside the source (white-box, `package object_storage`) —
it reaches no unexported identifiers today, but keeping the port's own test
beside it matches `kernel`/`configuration`'s style.

`object_storage_test.go` covers `Namespaced` with a recording fake `Store`:
every method prefixes its key, the version is preserved untouched, surrounding
slashes are trimmed, and a blank/`"/"` folder returns the store unchanged.
The behavior of a real backend is proven by
[`internal/adapter/storage`](../../adapter/storage/README.md)'s integration
suite. See `server/TESTING.md`.
