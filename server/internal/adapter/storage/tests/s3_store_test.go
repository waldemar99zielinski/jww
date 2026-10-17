//go:build integration

// Package storage_test exercises the generic S3 object-storage adapter against
// a real MinIO container — never a mocked S3 client. One container is started
// once for the whole package via TestMain.
package storage_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"jww/internal/adapter/storage"
	"jww/internal/domain"
	"jww/internal/service/object_storage"
	"jww/internal/tests/s3test"
)

var (
	bucket *s3test.Bucket
	appCtx domain.AppContext
)

func TestMain(m *testing.M) {
	ctx := context.Background()
	b, err := s3test.Start(ctx, "jww-test")
	if err != nil {
		panic("storage_test: starting MinIO: " + err.Error())
	}
	bucket = b
	appCtx.SetContext(ctx)

	code := m.Run()

	b.Close(ctx)
	os.Exit(code)
}

func newStore(t *testing.T) *storage.S3Store {
	t.Helper()
	s, err := storage.NewS3Store(context.Background(), bucket.Config)
	require.NoError(t, err)
	return s
}

func key() string { return "objects/" + uuid.NewString() }

func TestS3Store_RoundTrip(t *testing.T) {
	s := newStore(t)
	k := key()
	// Every byte value, so an accidental text/encoding assumption would fail.
	payload := make([]byte, 256)
	for i := range payload {
		payload[i] = byte(i)
	}

	stored, err := s.Put(appCtx, object_storage.Ref{Key: k}, bytes.NewReader(payload), "application/octet-stream")
	require.NoError(t, err)
	assert.NotEmpty(t, stored.ETag)

	stream, err := s.Open(appCtx, object_storage.Ref{Key: k, Version: stored.Version})
	require.NoError(t, err)
	defer stream.Body.Close()
	got, err := io.ReadAll(stream.Body)
	require.NoError(t, err)
	assert.Equal(t, payload, got)
}

func TestS3Store_ZeroByteObject(t *testing.T) {
	s := newStore(t)
	k := key()

	_, err := s.Put(appCtx, object_storage.Ref{Key: k}, bytes.NewReader(nil), "application/octet-stream")
	require.NoError(t, err)

	stream, err := s.Open(appCtx, object_storage.Ref{Key: k})
	require.NoError(t, err)
	defer stream.Body.Close()
	got, err := io.ReadAll(stream.Body)
	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestS3Store_ArbitraryContentType(t *testing.T) {
	s := newStore(t)
	k := key()

	// Any claimed media type is accepted; it is stored as object metadata only.
	_, err := s.Put(appCtx, object_storage.Ref{Key: k}, bytes.NewReader([]byte("data")), "application/x-custom; charset=binary")
	require.NoError(t, err)

	stream, err := s.Open(appCtx, object_storage.Ref{Key: k})
	require.NoError(t, err)
	defer stream.Body.Close()
	got, err := io.ReadAll(stream.Body)
	require.NoError(t, err)
	assert.Equal(t, []byte("data"), got)
}

func TestS3Store_DeleteIsIdempotent(t *testing.T) {
	s := newStore(t)
	k := key()
	_, err := s.Put(appCtx, object_storage.Ref{Key: k}, bytes.NewReader([]byte("x")), "text/plain")
	require.NoError(t, err)

	require.NoError(t, s.Delete(appCtx, object_storage.Ref{Key: k}))
	require.NoError(t, s.Delete(appCtx, object_storage.Ref{Key: k}), "deleting an absent object succeeds")
}

func TestS3Store_OpenMissingIsNotFound(t *testing.T) {
	s := newStore(t)

	_, err := s.Open(appCtx, object_storage.Ref{Key: key()})
	require.Error(t, err)
	assert.True(t, errors.Is(err, object_storage.ErrNotFound))
}

func TestS3Store_CancelledContextFails(t *testing.T) {
	s := newStore(t)
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()

	var ctx domain.AppContext
	ctx.SetContext(cancelled)

	_, err := s.Put(ctx, object_storage.Ref{Key: key()}, bytes.NewReader([]byte("x")), "text/plain")
	assert.Error(t, err)
}
