package transaction_attachment_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"jww/internal/domain"
	"jww/internal/service/object_storage"
	"jww/internal/service/transaction_attachment"
	"jww/internal/service/unit_of_work"
)

func appCtx() domain.AppContext {
	var ctx domain.AppContext
	ctx.SetContext(context.Background())
	return ctx
}

type fakeRepo struct {
	byID      map[uuid.UUID]domain.TransactionAttachment
	createErr error
	deleteErr error
	created   []domain.TransactionAttachment
	deleted   []uuid.UUID
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{byID: map[uuid.UUID]domain.TransactionAttachment{}}
}

func (r *fakeRepo) Create(_ domain.AppContext, a domain.TransactionAttachment, _ ...unit_of_work.UnitOfWork) error {
	if r.createErr != nil {
		return r.createErr
	}
	r.byID[a.ID] = a
	r.created = append(r.created, a)
	return nil
}

func (r *fakeRepo) GetByID(_ domain.AppContext, id uuid.UUID, _ ...unit_of_work.UnitOfWork) (*domain.TransactionAttachment, error) {
	a, ok := r.byID[id]
	if !ok {
		return nil, nil
	}
	return &a, nil
}

func (r *fakeRepo) Find(_ domain.AppContext, filter transaction_attachment.Filter, _ ...unit_of_work.UnitOfWork) (*domain.TransactionAttachment, error) {
	a, ok := r.byID[filter.ID]
	if !ok || filter.DraftOnly && a.TransactionID != nil {
		return nil, nil
	}
	if filter.TransactionID != nil && (a.TransactionID == nil || *a.TransactionID != *filter.TransactionID) {
		return nil, nil
	}
	return &a, nil
}

func (r *fakeRepo) LockByIDs(_ domain.AppContext, ids []uuid.UUID, _ ...unit_of_work.UnitOfWork) ([]domain.TransactionAttachment, error) {
	var out []domain.TransactionAttachment
	for _, id := range ids {
		if a, ok := r.byID[id]; ok {
			out = append(out, a)
		}
	}
	return out, nil
}

func (r *fakeRepo) LinkToTransaction(_ domain.AppContext, transactionID int64, ids []uuid.UUID, _ ...unit_of_work.UnitOfWork) error {
	for _, id := range ids {
		if a, ok := r.byID[id]; ok && a.TransactionID == nil {
			txID := transactionID
			a.TransactionID = &txID
			r.byID[id] = a
		}
	}
	return nil
}

func (r *fakeRepo) ListByTransactionIDs(_ domain.AppContext, ids []int64, _ ...unit_of_work.UnitOfWork) ([]domain.TransactionAttachment, error) {
	want := map[int64]struct{}{}
	for _, id := range ids {
		want[id] = struct{}{}
	}
	var out []domain.TransactionAttachment
	for _, a := range r.byID {
		if a.TransactionID != nil {
			if _, ok := want[*a.TransactionID]; ok {
				out = append(out, a)
			}
		}
	}
	return out, nil
}

func (r *fakeRepo) Delete(_ domain.AppContext, a domain.TransactionAttachment, _ ...unit_of_work.UnitOfWork) error {
	if r.deleteErr != nil {
		return r.deleteErr
	}
	r.deleted = append(r.deleted, a.ID)
	delete(r.byID, a.ID)
	return nil
}

type fakeStorage struct {
	objects   map[string][]byte
	putErr    error
	openErr   error
	deleteErr error
	putCalls  int
	deleted   []string
}

func newFakeStorage() *fakeStorage {
	return &fakeStorage{objects: map[string][]byte{}}
}

func (s *fakeStorage) Put(_ domain.AppContext, ref object_storage.Ref, body io.Reader, _ string) (object_storage.Stored, error) {
	s.putCalls++
	if s.putErr != nil {
		return object_storage.Stored{}, s.putErr
	}
	b, err := io.ReadAll(body)
	if err != nil {
		return object_storage.Stored{}, err
	}
	s.objects[ref.Key] = b
	return object_storage.Stored{Version: "v1", ETag: "etag"}, nil
}

func (s *fakeStorage) Open(_ domain.AppContext, ref object_storage.Ref) (object_storage.Stream, error) {
	if s.openErr != nil {
		return object_storage.Stream{}, s.openErr
	}
	b, ok := s.objects[ref.Key]
	if !ok {
		return object_storage.Stream{}, object_storage.ErrNotFound
	}
	return object_storage.Stream{Body: io.NopCloser(bytes.NewReader(b)), ByteSize: int64(len(b))}, nil
}

func (s *fakeStorage) Delete(_ domain.AppContext, ref object_storage.Ref) error {
	if s.deleteErr != nil {
		return s.deleteErr
	}
	s.deleted = append(s.deleted, ref.Key)
	delete(s.objects, ref.Key)
	return nil
}

type orphanCall struct {
	id  uuid.UUID
	err error
}

func newService(repo *fakeRepo, storage *fakeStorage) (*transaction_attachment.Service, *[]orphanCall) {
	var orphans []orphanCall
	svc := transaction_attachment.NewService(repo, storage, func(_ domain.AppContext, id uuid.UUID, err error) {
		orphans = append(orphans, orphanCall{id: id, err: err})
	})
	return svc, &orphans
}

func TestService_UploadDraft(t *testing.T) {
	t.Run("streams, counts bytes, and stores metadata with a null transaction", func(t *testing.T) {
		repo, storage := newFakeRepo(), newFakeStorage()
		svc, _ := newService(repo, storage)

		attachment, err := svc.UploadDraft(appCtx(), "receipt.pdf", "application/pdf", strings.NewReader("hello world"))

		require.NoError(t, err)
		assert.Equal(t, int64(len("hello world")), attachment.ByteSize)
		assert.Nil(t, attachment.TransactionID)
		assert.Equal(t, "receipt.pdf", attachment.OriginalFilename)
		assert.Equal(t, "v1", attachment.StorageVersion)
		assert.NotContains(t, attachment.StorageKey, "receipt.pdf", "the key must be opaque")
		assert.Len(t, repo.created, 1)
	})

	t.Run("storage failure surfaces unavailable and writes no metadata", func(t *testing.T) {
		repo, storage := newFakeRepo(), newFakeStorage()
		storage.putErr = errors.New("connection refused")
		svc, _ := newService(repo, storage)

		_, err := svc.UploadDraft(appCtx(), "a.pdf", "application/pdf", strings.NewReader("x"))

		require.Error(t, err)
		var domainErr domain.DomainError
		require.ErrorAs(t, err, &domainErr)
		assert.Equal(t, domain.ErrorKindUnavailable, domainErr.Kind())
		assert.Empty(t, repo.created)
	})

	t.Run("metadata failure compensates by deleting the object", func(t *testing.T) {
		repo, storage := newFakeRepo(), newFakeStorage()
		repo.createErr = errors.New("insert failed")
		svc, orphans := newService(repo, storage)

		_, err := svc.UploadDraft(appCtx(), "a.pdf", "application/pdf", strings.NewReader("x"))

		require.Error(t, err)
		assert.Len(t, storage.deleted, 1, "the just-written object must be deleted")
		assert.Empty(t, *orphans)
	})

	t.Run("a failed compensation reports exactly one orphan", func(t *testing.T) {
		repo, storage := newFakeRepo(), newFakeStorage()
		repo.createErr = errors.New("insert failed")
		storage.deleteErr = errors.New("delete failed")
		svc, orphans := newService(repo, storage)

		_, err := svc.UploadDraft(appCtx(), "a.pdf", "application/pdf", strings.NewReader("x"))

		require.Error(t, err)
		require.Len(t, *orphans, 1)
	})
}

func TestService_UploadForTransaction(t *testing.T) {
	repo, storage := newFakeRepo(), newFakeStorage()
	svc, _ := newService(repo, storage)

	attachment, err := svc.UploadForTransaction(appCtx(), 12, "a.pdf", "application/pdf", strings.NewReader("abc"))

	require.NoError(t, err)
	require.NotNil(t, attachment.TransactionID)
	assert.Equal(t, int64(12), *attachment.TransactionID)
}

func TestService_GetDraft(t *testing.T) {
	repo, storage := newFakeRepo(), newFakeStorage()
	svc, _ := newService(repo, storage)

	t.Run("missing draft is not found", func(t *testing.T) {
		_, err := svc.GetDraft(appCtx(), uuid.New())
		var domainErr domain.DomainError
		require.ErrorAs(t, err, &domainErr)
		assert.Equal(t, domain.ErrorKindNotFound, domainErr.Kind())
	})

	t.Run("a linked attachment is invisible to standalone access", func(t *testing.T) {
		txID := int64(1)
		linked := seedAttachment(repo)
		linked.TransactionID = &txID
		repo.byID[linked.ID] = linked

		_, err := svc.GetDraft(appCtx(), linked.ID)
		require.Error(t, err)
	})
}

func TestService_OpenDraft(t *testing.T) {
	t.Run("success returns the bytes", func(t *testing.T) {
		repo, storage := newFakeRepo(), newFakeStorage()
		svc, _ := newService(repo, storage)
		attachment, err := svc.UploadDraft(appCtx(), "a.pdf", "application/pdf", strings.NewReader("payload"))
		require.NoError(t, err)

		_, stream, err := svc.OpenDraft(appCtx(), attachment.ID)
		require.NoError(t, err)
		defer stream.Body.Close()
		got, err := io.ReadAll(stream.Body)
		require.NoError(t, err)
		assert.Equal(t, "payload", string(got))
	})

	t.Run("a missing object is an integrity fault, not a 404", func(t *testing.T) {
		repo, storage := newFakeRepo(), newFakeStorage()
		svc, _ := newService(repo, storage)
		attachment := seedAttachment(repo) // row present, object absent

		_, _, err := svc.OpenDraft(appCtx(), attachment.ID)
		var domainErr domain.DomainError
		require.ErrorAs(t, err, &domainErr)
		assert.Equal(t, domain.ErrorKindInternal, domainErr.Kind())
	})
}

func TestService_DeleteDraft(t *testing.T) {
	t.Run("deletes the object before the row", func(t *testing.T) {
		repo, storage := newFakeRepo(), newFakeStorage()
		svc, _ := newService(repo, storage)
		attachment := seedAttachment(repo)
		storage.objects[attachment.StorageKey] = []byte("x")

		require.NoError(t, svc.DeleteDraft(appCtx(), attachment.ID))
		assert.Equal(t, []string{attachment.StorageKey}, storage.deleted)
		assert.Equal(t, []uuid.UUID{attachment.ID}, repo.deleted)
	})

	t.Run("a storage failure preserves the row", func(t *testing.T) {
		repo, storage := newFakeRepo(), newFakeStorage()
		storage.deleteErr = errors.New("nope")
		svc, _ := newService(repo, storage)
		attachment := seedAttachment(repo)

		err := svc.DeleteDraft(appCtx(), attachment.ID)
		require.Error(t, err)
		assert.Empty(t, repo.deleted)
		_, stillThere := repo.byID[attachment.ID]
		assert.True(t, stillThere)
	})
}

func TestService_DeleteForTransaction_RequiresParentMatch(t *testing.T) {
	repo, storage := newFakeRepo(), newFakeStorage()
	svc, _ := newService(repo, storage)
	txID := int64(5)
	other := int64(6)
	attachment := seedAttachment(repo)
	attachment.TransactionID = &txID
	repo.byID[attachment.ID] = attachment

	require.Error(t, svc.DeleteForTransaction(appCtx(), other, attachment.ID))
	require.NoError(t, svc.DeleteForTransaction(appCtx(), txID, attachment.ID))
}

func TestService_DeleteAllForTransaction(t *testing.T) {
	t.Run("deletes both object bytes and metadata", func(t *testing.T) {
		repo, storage := newFakeRepo(), newFakeStorage()
		svc, _ := newService(repo, storage)
		a1 := seedAttachment(repo)
		a2 := seedAttachment(repo)
		transactionID := int64(7)
		a1.TransactionID = &transactionID
		a2.TransactionID = &transactionID
		repo.byID[a1.ID] = a1
		repo.byID[a2.ID] = a2
		storage.objects[a1.StorageKey] = []byte("a")
		storage.objects[a2.StorageKey] = []byte("b")

		require.NoError(t, svc.DeleteAllForTransaction(appCtx(), transactionID))
		assert.ElementsMatch(t, []string{a1.StorageKey, a2.StorageKey}, storage.deleted)
		assert.ElementsMatch(t, []uuid.UUID{a1.ID, a2.ID}, repo.deleted)
		assert.NotContains(t, repo.byID, a1.ID)
		assert.NotContains(t, repo.byID, a2.ID)
	})

	t.Run("stops on storage failure without deleting that metadata", func(t *testing.T) {
		repo, storage := newFakeRepo(), newFakeStorage()
		svc, _ := newService(repo, storage)
		attachment := seedAttachment(repo)
		transactionID := int64(7)
		attachment.TransactionID = &transactionID
		repo.byID[attachment.ID] = attachment
		storage.deleteErr = errors.New("nope")

		err := svc.DeleteAllForTransaction(appCtx(), transactionID)
		require.Error(t, err)
		assert.Empty(t, repo.deleted)
		assert.Contains(t, repo.byID, attachment.ID)
	})
}

func seedAttachment(repo *fakeRepo) domain.TransactionAttachment {
	a := domain.TransactionAttachment{
		ID:               uuid.New(),
		OriginalFilename: "receipt.pdf",
		MediaType:        "application/pdf",
		ByteSize:         1,
		StorageKey:       uuid.NewString(),
		StorageVersion:   "v1",
		CreatedAt:        time.Now().UTC(),
	}
	repo.byID[a.ID] = a
	return a
}
