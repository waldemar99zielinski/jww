package transaction_attachment_test

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"jww/internal/adapter/schema"
	"jww/internal/domain"
	"jww/internal/infrastructure/http_server/kernel"
	attachmentrouter "jww/internal/infrastructure/http_server/routers/transaction_attachment"
	"jww/internal/service/object_storage"
	"jww/internal/service/transaction_attachment"
	"jww/internal/service/unit_of_work"
)

// --- fakes -----------------------------------------------------------------

type fakeRepo struct {
	byID    map[uuid.UUID]domain.TransactionAttachment
	created []domain.TransactionAttachment
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{byID: map[uuid.UUID]domain.TransactionAttachment{}}
}

func (r *fakeRepo) Create(_ domain.AppContext, a domain.TransactionAttachment, _ ...unit_of_work.UnitOfWork) error {
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
	return nil
}

func (r *fakeRepo) ListByTransactionIDs(_ domain.AppContext, ids []int64, _ ...unit_of_work.UnitOfWork) ([]domain.TransactionAttachment, error) {
	return nil, nil
}

func (r *fakeRepo) Delete(_ domain.AppContext, a domain.TransactionAttachment, _ ...unit_of_work.UnitOfWork) error {
	delete(r.byID, a.ID)
	return nil
}

type fakeStorage struct {
	objects map[string][]byte
	failPut bool
	deleted []string
}

func newFakeStorage() *fakeStorage { return &fakeStorage{objects: map[string][]byte{}} }

func (s *fakeStorage) Put(_ domain.AppContext, ref object_storage.Ref, body io.Reader, _ string) (object_storage.Stored, error) {
	if s.failPut {
		return object_storage.Stored{}, io.ErrUnexpectedEOF
	}
	b, err := io.ReadAll(body)
	if err != nil {
		return object_storage.Stored{}, err
	}
	s.objects[ref.Key] = b
	return object_storage.Stored{Version: "v1", ETag: "etag"}, nil
}

func (s *fakeStorage) Open(_ domain.AppContext, ref object_storage.Ref) (object_storage.Stream, error) {
	b, ok := s.objects[ref.Key]
	if !ok {
		return object_storage.Stream{}, object_storage.ErrNotFound
	}
	return object_storage.Stream{Body: io.NopCloser(bytes.NewReader(b)), ByteSize: int64(len(b))}, nil
}

func (s *fakeStorage) Delete(_ domain.AppContext, ref object_storage.Ref) error {
	s.deleted = append(s.deleted, ref.Key)
	delete(s.objects, ref.Key)
	return nil
}

// --- helpers ---------------------------------------------------------------

type fixture struct {
	router  http.Handler
	repo    *fakeRepo
	storage *fakeStorage
}

func newFixture() *fixture {
	repo := newFakeRepo()
	storage := newFakeStorage()
	svc := transaction_attachment.NewService(repo, storage, nil)
	deps := &kernel.Deps{TransactionAttachmentService: svc}
	r := chi.NewRouter()
	attachmentrouter.Mount(r, r, deps)
	return &fixture{router: r, repo: repo, storage: storage}
}

func multipartBody(t *testing.T, filename, contentType, content string, extra func(*multipart.Writer)) (*bytes.Buffer, string) {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	header := textproto.MIMEHeader{}
	header.Set("Content-Disposition", `form-data; name="file"; filename="`+filename+`"`)
	header.Set("Content-Type", contentType)
	part, err := w.CreatePart(header)
	require.NoError(t, err)
	_, err = io.WriteString(part, content)
	require.NoError(t, err)
	if extra != nil {
		extra(w)
	}
	require.NoError(t, w.Close())
	return &buf, w.FormDataContentType()
}

func doMultipart(t *testing.T, f *fixture, filename, contentType, content string, extra func(*multipart.Writer)) *httptest.ResponseRecorder {
	t.Helper()
	body, formContentType := multipartBody(t, filename, contentType, content, extra)
	req := httptest.NewRequest(http.MethodPost, "/v1/transaction-attachments", body)
	req.Header.Set("Content-Type", formContentType)
	rec := httptest.NewRecorder()
	f.router.ServeHTTP(rec, req)
	return rec
}

func decodeData(t *testing.T, rec *httptest.ResponseRecorder, v any) {
	t.Helper()
	var body struct {
		Data json.RawMessage `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.NoError(t, json.Unmarshal(body.Data, v))
}

func seedDraft(repo *fakeRepo) domain.TransactionAttachment {
	a := domain.TransactionAttachment{
		ID:               uuid.New(),
		OriginalFilename: "receipt.pdf",
		MediaType:        "application/pdf",
		ByteSize:         3,
		StorageKey:       uuid.NewString(),
		StorageVersion:   "v1",
		CreatedAt:        time.Now().UTC(),
	}
	repo.byID[a.ID] = a
	return a
}

// --- tests -----------------------------------------------------------------

func TestUploadDraftHandler(t *testing.T) {
	t.Run("happy path returns 201 with a null transaction id", func(t *testing.T) {
		f := newFixture()
		rec := doMultipart(t, f, "receipt.pdf", "application/pdf", "pdf-bytes", nil)

		require.Equal(t, http.StatusCreated, rec.Code)
		var resp schema.TransactionAttachmentResponse
		decodeData(t, rec, &resp)
		assert.Nil(t, resp.TransactionID)
		assert.Equal(t, "receipt.pdf", resp.OriginalFilename)
		assert.Equal(t, int64(len("pdf-bytes")), resp.ByteSize)
		assert.NotEqual(t, uuid.Nil, resp.ID)
	})

	t.Run("a second file part is a bad request and leaves nothing behind", func(t *testing.T) {
		f := newFixture()
		rec := doMultipart(t, f, "a.pdf", "application/pdf", "x", func(w *multipart.Writer) {
			part, err := w.CreateFormFile("file", "b.pdf")
			require.NoError(t, err)
			_, _ = part.Write([]byte("y"))
		})

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Empty(t, f.repo.byID, "the created draft must be compensated away")
	})

	t.Run("storage failure is a 503", func(t *testing.T) {
		f := newFixture()
		f.storage.failPut = true
		rec := doMultipart(t, f, "a.pdf", "application/pdf", "x", nil)

		assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
		assert.Empty(t, f.repo.byID)
	})
}

func TestGetDraftHandler(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		f := newFixture()
		draft := seedDraft(f.repo)
		req := httptest.NewRequest(http.MethodGet, "/v1/transaction-attachments/"+draft.ID.String(), nil)
		rec := httptest.NewRecorder()
		f.router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		var resp schema.TransactionAttachmentResponse
		decodeData(t, rec, &resp)
		assert.Equal(t, draft.ID, resp.ID)
	})

	t.Run("linked attachment is 404", func(t *testing.T) {
		f := newFixture()
		draft := seedDraft(f.repo)
		draft.TransactionID = ptrInt64(9)
		f.repo.byID[draft.ID] = draft

		req := httptest.NewRequest(http.MethodGet, "/v1/transaction-attachments/"+draft.ID.String(), nil)
		rec := httptest.NewRecorder()
		f.router.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("malformed id is 400", func(t *testing.T) {
		f := newFixture()
		req := httptest.NewRequest(http.MethodGet, "/v1/transaction-attachments/not-a-uuid", nil)
		rec := httptest.NewRecorder()
		f.router.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func TestDownloadDraftHandler(t *testing.T) {
	t.Run("streams the exact bytes with safe headers", func(t *testing.T) {
		f := newFixture()
		rec := doMultipart(t, f, "receipt.pdf", "application/pdf", "pdf-bytes", nil)
		require.Equal(t, http.StatusCreated, rec.Code)
		var uploaded schema.TransactionAttachmentResponse
		decodeData(t, rec, &uploaded)

		req := httptest.NewRequest(http.MethodGet, "/v1/transaction-attachments/"+uploaded.ID.String()+"/content", nil)
		dl := httptest.NewRecorder()
		f.router.ServeHTTP(dl, req)

		require.Equal(t, http.StatusOK, dl.Code)
		assert.Equal(t, "pdf-bytes", dl.Body.String())
		assert.Equal(t, "application/pdf", dl.Header().Get("Content-Type"))
		assert.Contains(t, dl.Header().Get("Content-Disposition"), "attachment;")
		assert.Equal(t, "nosniff", dl.Header().Get("X-Content-Type-Options"))
	})
}

func TestDeleteDraftHandler(t *testing.T) {
	t.Run("deletes the object and the row", func(t *testing.T) {
		f := newFixture()
		draft := seedDraft(f.repo)
		f.storage.objects[draft.StorageKey] = []byte("x")

		req := httptest.NewRequest(http.MethodDelete, "/v1/transaction-attachments/"+draft.ID.String(), nil)
		rec := httptest.NewRecorder()
		f.router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNoContent, rec.Code)
		assert.NotContains(t, f.repo.byID, draft.ID)
		assert.Equal(t, []string{draft.StorageKey}, f.storage.deleted)
	})

	t.Run("linked attachment is 404 and nothing is deleted", func(t *testing.T) {
		f := newFixture()
		draft := seedDraft(f.repo)
		draft.TransactionID = ptrInt64(9)
		f.repo.byID[draft.ID] = draft

		req := httptest.NewRequest(http.MethodDelete, "/v1/transaction-attachments/"+draft.ID.String(), nil)
		rec := httptest.NewRecorder()
		f.router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
		assert.Empty(t, f.storage.deleted)
	})
}

func ptrInt64(v int64) *int64 { return &v }
