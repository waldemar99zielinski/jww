// Package object_storage defines the persistence-agnostic outbound port for
// blob storage, with zero knowledge of what actually backs it (S3, MinIO, a
// local filesystem — see internal/adapter/storage for the S3-compatible
// implementation). It mirrors service/unit_of_work: a small, shared,
// cross-cutting infrastructure port rather than a resource-specific one, so
// several independent purposes (transaction attachments today, others later)
// can reuse one implementation and one connection.
package object_storage

import (
	"errors"
	"io"
	"strings"

	"jww/internal/domain"
)

// ErrNotFound is returned (wrapped) by an implementation when the requested
// object or version does not exist.
var ErrNotFound = errors.New("object not found")

// StorageTransactionAttachmentsNamespace is the folder used for transaction attachment objects in
// the shared bucket. It is stable: changing it relocates every stored
// attachment object.
const StorageTransactionAttachmentsNamespace = "transaction-attachments"

// Ref identifies one stored object. Key is a caller-defined, opaque path within
// the store (the Store adds no namespace of its own). Version is empty for a
// Put and the exact stored version for an Open/Delete of a versioned object.
type Ref struct {
	Key     string
	Version string
}

// Stored is the store-side result of a successful Put.
type Stored struct {
	Version string
	ETag    string
}

// Stream is a readable object. Body is closed by the caller.
type Stream struct {
	Body     io.ReadCloser
	ByteSize int64
}

// Store is the outbound object-storage port. Implementations live in
// internal/adapter/*; no provider SDK type crosses this boundary. Delete must
// be idempotent when the object (or version) is already absent.
type Store interface {
	Put(ctx domain.AppContext, ref Ref, body io.Reader, mediaType string) (Stored, error)
	Open(ctx domain.AppContext, ref Ref) (Stream, error)
	Delete(ctx domain.AppContext, ref Ref) error
}

// Namespaced wraps a Store so every key is scoped under "<folder>/". It is how
// independent purposes share one underlying Store (and one bucket) while keeping
// their objects clearly separated — e.g. transaction attachments use the
// "transaction-attachments" folder, a future feature would use its own. Keys
// passed through are namespace-relative (a bare name); a blank/"/"-only folder
// returns the store unchanged, and surrounding slashes are trimmed.
//
// Define stable namespace constants in this package and apply Namespaced at
// wiring time.
func Namespaced(store Store, folder string) Store {
	folder = strings.Trim(folder, "/")
	if folder == "" {
		return store
	}
	return namespacedStore{store: store, prefix: folder + "/"}
}

type namespacedStore struct {
	store  Store
	prefix string
}

func (n namespacedStore) Put(ctx domain.AppContext, ref Ref, body io.Reader, mediaType string) (Stored, error) {
	ref.Key = n.prefix + ref.Key
	return n.store.Put(ctx, ref, body, mediaType)
}

func (n namespacedStore) Open(ctx domain.AppContext, ref Ref) (Stream, error) {
	ref.Key = n.prefix + ref.Key
	return n.store.Open(ctx, ref)
}

func (n namespacedStore) Delete(ctx domain.AppContext, ref Ref) error {
	ref.Key = n.prefix + ref.Key
	return n.store.Delete(ctx, ref)
}
