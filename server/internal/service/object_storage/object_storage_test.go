package object_storage

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"

	"jww/internal/domain"
)

// recordingStore captures the refs each method receives, so Namespaced's key
// rewriting can be asserted without a real backend.
type recordingStore struct {
	puts    []Ref
	opens   []Ref
	deletes []Ref
}

func (s *recordingStore) Put(_ domain.AppContext, ref Ref, _ io.Reader, _ string) (Stored, error) {
	s.puts = append(s.puts, ref)
	return Stored{}, nil
}

func (s *recordingStore) Open(_ domain.AppContext, ref Ref) (Stream, error) {
	s.opens = append(s.opens, ref)
	return Stream{Body: io.NopCloser(bytes.NewReader(nil))}, nil
}

func (s *recordingStore) Delete(_ domain.AppContext, ref Ref) error {
	s.deletes = append(s.deletes, ref)
	return nil
}

func keyOf(refs []Ref) []string {
	out := make([]string, len(refs))
	for i, r := range refs {
		out[i] = r.Key
	}
	return out
}

func TestNamespaced_PrefixesEveryKey(t *testing.T) {
	base := &recordingStore{}
	scoped := Namespaced(base, "transaction-attachments")

	_, _ = scoped.Put(domain.AppContext{}, Ref{Key: "abc"}, bytes.NewReader(nil), "text/plain")
	_, _ = scoped.Open(domain.AppContext{}, Ref{Key: "abc", Version: "v1"})
	_ = scoped.Delete(domain.AppContext{}, Ref{Key: "abc"})

	assert.Equal(t, []string{"transaction-attachments/abc"}, keyOf(base.puts))
	assert.Equal(t, []string{"transaction-attachments/abc"}, keyOf(base.opens))
	assert.Equal(t, []string{"transaction-attachments/abc"}, keyOf(base.deletes))
	assert.Equal(t, "v1", base.opens[0].Version, "the version must be preserved untouched")
}

func TestNamespaced_TrimsSurroundingSlashes(t *testing.T) {
	base := &recordingStore{}
	scoped := Namespaced(base, "/avatars/")

	_, _ = scoped.Put(domain.AppContext{}, Ref{Key: "x"}, bytes.NewReader(nil), "")

	assert.Equal(t, "avatars/x", base.puts[0].Key)
}

func TestNamespaced_BlankFolderIsIdentity(t *testing.T) {
	base := &recordingStore{}

	assert.Equal(t, Store(base), Namespaced(base, ""))
	assert.Equal(t, Store(base), Namespaced(base, "/"))
}
