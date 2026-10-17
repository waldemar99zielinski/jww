package transaction_test

import (
	"bytes"
	"io"
	"time"

	"github.com/google/uuid"

	"jww/internal/domain"
	"jww/internal/service/object_storage"
	"jww/internal/service/transaction_attachment"
	"jww/internal/service/unit_of_work"
)

// fakeAttachmentRepo is a minimal in-memory transaction_attachment.Repository
// for the handler tests, which drive the real service over it.
type fakeAttachmentRepo struct {
	byID    map[uuid.UUID]domain.TransactionAttachment
	created []domain.TransactionAttachment
}

func newFakeAttachmentRepo() *fakeAttachmentRepo {
	return &fakeAttachmentRepo{byID: map[uuid.UUID]domain.TransactionAttachment{}}
}

func (r *fakeAttachmentRepo) Create(_ domain.AppContext, a domain.TransactionAttachment, _ ...unit_of_work.UnitOfWork) error {
	r.byID[a.ID] = a
	r.created = append(r.created, a)
	return nil
}

func (r *fakeAttachmentRepo) GetByID(_ domain.AppContext, id uuid.UUID, _ ...unit_of_work.UnitOfWork) (*domain.TransactionAttachment, error) {
	a, ok := r.byID[id]
	if !ok {
		return nil, nil
	}
	return &a, nil
}

func (r *fakeAttachmentRepo) Find(_ domain.AppContext, filter transaction_attachment.Filter, _ ...unit_of_work.UnitOfWork) (*domain.TransactionAttachment, error) {
	a, ok := r.byID[filter.ID]
	if !ok || filter.DraftOnly && a.TransactionID != nil {
		return nil, nil
	}
	if filter.TransactionID != nil && (a.TransactionID == nil || *a.TransactionID != *filter.TransactionID) {
		return nil, nil
	}
	return &a, nil
}

func (r *fakeAttachmentRepo) LockByIDs(_ domain.AppContext, ids []uuid.UUID, _ ...unit_of_work.UnitOfWork) ([]domain.TransactionAttachment, error) {
	var out []domain.TransactionAttachment
	for _, id := range ids {
		if a, ok := r.byID[id]; ok {
			out = append(out, a)
		}
	}
	return out, nil
}

func (r *fakeAttachmentRepo) LinkToTransaction(_ domain.AppContext, transactionID int64, ids []uuid.UUID, _ ...unit_of_work.UnitOfWork) error {
	for _, id := range ids {
		a, ok := r.byID[id]
		if !ok || a.TransactionID != nil {
			continue
		}
		txID := transactionID
		a.TransactionID = &txID
		r.byID[id] = a
	}
	return nil
}

func (r *fakeAttachmentRepo) ListByTransactionIDs(_ domain.AppContext, ids []int64, _ ...unit_of_work.UnitOfWork) ([]domain.TransactionAttachment, error) {
	want := make(map[int64]struct{}, len(ids))
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

func (r *fakeAttachmentRepo) Delete(_ domain.AppContext, a domain.TransactionAttachment, _ ...unit_of_work.UnitOfWork) error {
	delete(r.byID, a.ID)
	return nil
}

// fakeAttachmentStorage is a minimal in-memory object_storage.Store.
type fakeAttachmentStorage struct {
	objects map[string][]byte
	deleted []string
	failPut bool
}

func newFakeAttachmentStorage() *fakeAttachmentStorage {
	return &fakeAttachmentStorage{objects: map[string][]byte{}}
}

func (s *fakeAttachmentStorage) Put(_ domain.AppContext, ref object_storage.Ref, body io.Reader, _ string) (object_storage.Stored, error) {
	if s.failPut {
		return object_storage.Stored{}, io.ErrUnexpectedEOF
	}
	b, err := io.ReadAll(body)
	if err != nil {
		return object_storage.Stored{}, err
	}
	s.objects[ref.Key] = b
	return object_storage.Stored{Version: "v1", ETag: "etag-v1"}, nil
}

func (s *fakeAttachmentStorage) Open(_ domain.AppContext, ref object_storage.Ref) (object_storage.Stream, error) {
	b, ok := s.objects[ref.Key]
	if !ok {
		return object_storage.Stream{}, object_storage.ErrNotFound
	}
	return object_storage.Stream{Body: io.NopCloser(bytes.NewReader(b)), ByteSize: int64(len(b))}, nil
}

func (s *fakeAttachmentStorage) Delete(_ domain.AppContext, ref object_storage.Ref) error {
	s.deleted = append(s.deleted, ref.Key)
	delete(s.objects, ref.Key)
	return nil
}

// seedDraft adds an unattached attachment to the repo.
func (r *fakeAttachmentRepo) seedDraft() domain.TransactionAttachment {
	a := domain.TransactionAttachment{
		ID:               uuid.New(),
		OriginalFilename: "receipt.pdf",
		MediaType:        "application/pdf",
		ByteSize:         7,
		StorageKey:       uuid.NewString(),
		StorageVersion:   "v1",
		CreatedAt:        time.Now().UTC(),
	}
	r.byID[a.ID] = a
	return a
}
