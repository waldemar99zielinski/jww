package transaction_test

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"jww/internal/adapter/schema"
	"jww/internal/domain"
)

func seedTransaction(t *testing.T, fx *testFixtures, name string) int64 {
	t.Helper()
	tx, err := fx.transactionRepo.Create(domain.AppContext{}, domain.TransactionTypeExpense, domain.Money(500),
		fx.currency, fx.category, fx.user, domain.TransactionSourceManual, time.Now(), name, "")
	require.NoError(t, err)
	return tx.ID
}

func attachmentMultipartRequest(t *testing.T, path, filename, content string) *http.Request {
	t.Helper()
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, err := writer.CreateFormFile("file", filename)
	require.NoError(t, err)
	_, err = io.WriteString(part, content)
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	req := httptest.NewRequest(http.MethodPost, path, &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req
}

func TestCreateTransactionWithAttachmentIDs(t *testing.T) {
	t.Run("links the draft atomically", func(t *testing.T) {
		r, fx := newTestRouter(t)
		draft := fx.attachmentRepo.seedDraft()
		fx.attachmentStorage.objects[draft.StorageKey] = []byte("bytes")

		body := `{"type":"expense","amount":1000,"currency_code":"` + fx.currency.Code +
			`","category_id":"` + fx.category.ID.String() + `","user_id":"` + fx.user.ID.String() +
			`","source":"manual","name":"Groceries","attachment_ids":["` + draft.ID.String() + `"]}`
		req := httptest.NewRequest(http.MethodPost, "/v1/transactions", bytes.NewBufferString(body))
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		require.Equal(t, http.StatusCreated, rec.Code)
		var resp schema.TransactionResponse
		decodeData(t, rec, &resp)
		require.Len(t, resp.Attachments, 1)
		require.NotNil(t, resp.Attachments[0].TransactionID)
		assert.Equal(t, resp.ID, *resp.Attachments[0].TransactionID)

		linked := fx.attachmentRepo.byID[draft.ID]
		require.NotNil(t, linked.TransactionID)
		assert.Equal(t, resp.ID, *linked.TransactionID)
	})

	t.Run("duplicate ids are a 400", func(t *testing.T) {
		r, fx := newTestRouter(t)
		id := uuid.NewString()
		body := `{"type":"expense","amount":1000,"currency_code":"` + fx.currency.Code +
			`","category_id":"` + fx.category.ID.String() + `","user_id":"` + fx.user.ID.String() +
			`","source":"manual","name":"Groceries","attachment_ids":["` + id + `","` + id + `"]}`
		req := httptest.NewRequest(http.MethodPost, "/v1/transactions", bytes.NewBufferString(body))
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("unknown draft is a 404", func(t *testing.T) {
		r, fx := newTestRouter(t)
		body := `{"type":"expense","amount":1000,"currency_code":"` + fx.currency.Code +
			`","category_id":"` + fx.category.ID.String() + `","user_id":"` + fx.user.ID.String() +
			`","source":"manual","name":"Groceries","attachment_ids":["` + uuid.NewString() + `"]}`
		req := httptest.NewRequest(http.MethodPost, "/v1/transactions", bytes.NewBufferString(body))
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}

func TestAttachAttachmentHandler(t *testing.T) {
	t.Run("links a draft and is idempotent for the same parent", func(t *testing.T) {
		r, fx := newTestRouter(t)
		txID := seedTransaction(t, fx, "existing")
		draft := fx.attachmentRepo.seedDraft()

		path := "/v1/transactions/" + itoa(txID) + "/attachments/" + draft.ID.String()

		req := httptest.NewRequest(http.MethodPut, path, nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNoContent, rec.Code)

		rec = httptest.NewRecorder()
		r.ServeHTTP(rec, httptest.NewRequest(http.MethodPut, path, nil))
		assert.Equal(t, http.StatusNoContent, rec.Code, "repeating against the same parent is a no-op")
	})

	t.Run("a draft linked elsewhere is a 409", func(t *testing.T) {
		r, fx := newTestRouter(t)
		txID := seedTransaction(t, fx, "existing")
		draft := fx.attachmentRepo.seedDraft()
		other := int64(999)
		draft.TransactionID = &other
		fx.attachmentRepo.byID[draft.ID] = draft

		req := httptest.NewRequest(http.MethodPut, "/v1/transactions/"+itoa(txID)+"/attachments/"+draft.ID.String(), nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusConflict, rec.Code)
	})

	t.Run("unknown parent is a 404", func(t *testing.T) {
		r, fx := newTestRouter(t)
		draft := fx.attachmentRepo.seedDraft()
		req := httptest.NewRequest(http.MethodPut, "/v1/transactions/424242/attachments/"+draft.ID.String(), nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}

func TestNestedAttachmentUploadHandler(t *testing.T) {
	r, fx := newTestRouter(t)
	txID := seedTransaction(t, fx, "existing")

	req := attachmentMultipartRequest(t, "/v1/transactions/"+itoa(txID)+"/attachments", "receipt.pdf", "pdf-bytes")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusCreated, rec.Code)
	var resp schema.TransactionAttachmentResponse
	decodeData(t, rec, &resp)
	require.NotNil(t, resp.TransactionID)
	assert.Equal(t, txID, *resp.TransactionID)
	assert.Equal(t, int64(len("pdf-bytes")), resp.ByteSize)
}

func TestNestedAttachmentListHandler(t *testing.T) {
	r, fx := newTestRouter(t)
	txID := seedTransaction(t, fx, "existing")
	draft := fx.attachmentRepo.seedDraft()
	draft.TransactionID = &txID
	fx.attachmentRepo.byID[draft.ID] = draft

	req := httptest.NewRequest(http.MethodGet, "/v1/transactions/"+itoa(txID)+"/attachments", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var resp []schema.TransactionAttachmentResponse
	decodeData(t, rec, &resp)
	require.Len(t, resp, 1)
	assert.Equal(t, draft.ID, resp[0].ID)
}

func TestNestedAttachmentDownloadAndDeleteRequireParentMatch(t *testing.T) {
	r, fx := newTestRouter(t)
	txID := seedTransaction(t, fx, "existing")
	draft := fx.attachmentRepo.seedDraft()
	other := int64(777)
	draft.TransactionID = &other
	fx.attachmentRepo.byID[draft.ID] = draft

	base := "/v1/transactions/" + itoa(txID) + "/attachments/" + draft.ID.String()

	for _, tc := range []struct {
		method string
		path   string
	}{
		{http.MethodGet, base + "/download"},
		{http.MethodDelete, base},
	} {
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, httptest.NewRequest(tc.method, tc.path, nil))
		assert.Equalf(t, http.StatusNotFound, rec.Code, "%s %s should 404 on a parent mismatch", tc.method, tc.path)
	}
}

func TestNestedAttachmentDownloadAndDeleteHappyPath(t *testing.T) {
	r, fx := newTestRouter(t)
	txID := seedTransaction(t, fx, "existing")

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, attachmentMultipartRequest(t, "/v1/transactions/"+itoa(txID)+"/attachments", "receipt.pdf", "pdf-bytes"))
	require.Equal(t, http.StatusCreated, rec.Code)
	var uploaded schema.TransactionAttachmentResponse
	decodeData(t, rec, &uploaded)

	dl := httptest.NewRecorder()
	r.ServeHTTP(dl, httptest.NewRequest(http.MethodGet, "/v1/transactions/"+itoa(txID)+"/attachments/"+uploaded.ID.String()+"/download", nil))
	require.Equal(t, http.StatusOK, dl.Code)
	assert.Equal(t, "pdf-bytes", dl.Body.String())

	del := httptest.NewRecorder()
	r.ServeHTTP(del, httptest.NewRequest(http.MethodDelete, "/v1/transactions/"+itoa(txID)+"/attachments/"+uploaded.ID.String(), nil))
	assert.Equal(t, http.StatusNoContent, del.Code)
}

func itoa(id int64) string {
	return strconv.FormatInt(id, 10)
}
