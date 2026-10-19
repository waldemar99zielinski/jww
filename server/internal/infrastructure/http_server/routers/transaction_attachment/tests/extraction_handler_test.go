package transaction_attachment_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"jww/internal/adapter/schema"
	jwtadapter "jww/internal/adapter/security/jwt"
	"jww/internal/domain"
	"jww/internal/infrastructure/http_server/kernel"
	attachmentrouter "jww/internal/infrastructure/http_server/routers/transaction_attachment"
	"jww/internal/service/transaction_attachment"
	transaction_attachment_extraction "jww/internal/service/transaction_attachment_extraction"
)

// extractionBaseURL is the bare origin the router test wires as the API's
// public base URL, so the reader's absolute content URL starts with it.
const extractionBaseURL = "https://api.prefill.test"

// stubExtractionReader is a deterministic transaction_attachment_extraction.Extractor
// that never touches the network and records what it was asked to extract.
type stubExtractionReader struct {
	result  *transaction_attachment_extraction.TransactionPrefillResult
	err     error
	calls   int
	lastReq transaction_attachment_extraction.TransactionPrefillInput
}

func (r *stubExtractionReader) ExtractTransactionPrefill(_ domain.AppContext, req transaction_attachment_extraction.TransactionPrefillInput) (*transaction_attachment_extraction.TransactionPrefillResult, error) {
	r.calls++
	r.lastReq = req
	if r.err != nil {
		return nil, r.err
	}
	return r.result, nil
}

// extractionAttachmentSource adapts the real attachment service to the
// extraction service's AttachmentSource port, mirroring cmd/server's
// draftAttachmentSource.
type extractionAttachmentSource struct {
	attachments *transaction_attachment.Service
}

type staticExtractionCategories struct{}

func (staticExtractionCategories) TransactionCategories(_ domain.AppContext) (expense, income []string, err error) {
	return []string{"Groceries"}, []string{"Salary"}, nil
}

func (s extractionAttachmentSource) DraftAttachment(ctx domain.AppContext, id uuid.UUID) (*domain.TransactionAttachment, error) {
	return s.attachments.GetDraft(ctx, id)
}

func (s extractionAttachmentSource) ContentGrant(ctx domain.AppContext, id uuid.UUID) (string, error) {
	return s.attachments.IssueDraftGrant(ctx, id, transaction_attachment.GrantPurposePreview)
}

type extractionFixture struct {
	router  http.Handler
	repo    *fakeRepo
	storage *fakeStorage
	reader  *stubExtractionReader
}

func newExtractionFixture(reader *stubExtractionReader) *extractionFixture {
	repo := newFakeRepo()
	storage := newFakeStorage()
	grants := jwtadapter.NewAttachmentGrantService(jwtadapter.Config{
		SecretKey: "extraction-handler-test-secret", Algorithm: "HS256", AccessTokenExpireMinutes: 60,
	}, 5*time.Minute)
	attachmentService := transaction_attachment.NewService(repo, storage, nil, grants)
	extractionService := transaction_attachment_extraction.NewService(
		reader,
		extractionAttachmentSource{attachments: attachmentService},
		staticExtractionCategories{},
		transaction_attachment_extraction.Config{PublicBaseURL: extractionBaseURL},
	)
	deps := &kernel.Deps{
		TransactionAttachmentService:           attachmentService,
		TransactionAttachmentExtractionService: extractionService,
	}
	r := chi.NewRouter()
	attachmentrouter.Mount(r, r, deps)
	return &extractionFixture{router: r, repo: repo, storage: storage, reader: reader}
}

// seedImageDraft inserts an unattached image attachment, the only media type
// extraction accepts in v1.
func seedImageDraft(repo *fakeRepo) domain.TransactionAttachment {
	a := domain.TransactionAttachment{
		ID:               uuid.New(),
		OriginalFilename: "receipt.jpg",
		MediaType:        "image/jpeg",
		ByteSize:         10,
		StorageKey:       uuid.NewString(),
		StorageVersion:   "v1",
		CreatedAt:        time.Now().UTC(),
	}
	repo.byID[a.ID] = a
	return a
}

func postExtraction(f *extractionFixture, attachmentID uuid.UUID) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/v1/transaction-attachments/"+attachmentID.String()+"/extraction", nil)
	rec := httptest.NewRecorder()
	f.router.ServeHTTP(rec, req)
	return rec
}

func postBatchExtraction(f *extractionFixture, attachmentIDs []uuid.UUID) *httptest.ResponseRecorder {
	body, _ := json.Marshal(map[string]any{"attachment_ids": attachmentIDs})
	req := httptest.NewRequest(http.MethodPost, "/v1/transaction-attachments/extraction", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	f.router.ServeHTTP(rec, req)
	return rec
}

func TestExtractionHandler(t *testing.T) {
	t.Run("happy path returns a prefill body with scalars and items", func(t *testing.T) {
		confidence := 91.0
		reader := &stubExtractionReader{result: &transaction_attachment_extraction.TransactionPrefillResult{
			Type:         &transaction_attachment_extraction.PrefillField[domain.TransactionType]{Value: domain.TransactionTypeExpense, Confidence: &confidence, Source: "header"},
			Amount:       &transaction_attachment_extraction.PrefillField[domain.Money]{Value: domain.Money(2490)},
			CurrencyCode: &transaction_attachment_extraction.PrefillField[string]{Value: "USD"},
			Name:         &transaction_attachment_extraction.PrefillField[string]{Value: "Corner Market"},
			SubTransactions: []transaction_attachment_extraction.PrefillSubTransaction{
				{Name: "Milk 2%", Label: "dairy", Amount: domain.Money(349)},
				{Name: "Bread", Label: "bakery", Amount: domain.Money(219)},
			},
		}}
		f := newExtractionFixture(reader)
		draft := seedImageDraft(f.repo)

		rec := postExtraction(f, draft.ID)

		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		var resp schema.TransactionPrefillResponse
		decodeData(t, rec, &resp)
		assert.Equal(t, draft.ID, resp.AttachmentID)
		require.NotNil(t, resp.Type)
		assert.Equal(t, domain.TransactionTypeExpense, resp.Type.Value)
		require.NotNil(t, resp.Amount)
		assert.Equal(t, domain.Money(2490), resp.Amount.Value)
		require.NotNil(t, resp.CurrencyCode)
		assert.Equal(t, "USD", resp.CurrencyCode.Value)
		require.NotNil(t, resp.Name)
		assert.Equal(t, "Corner Market", resp.Name.Value)
		require.Len(t, resp.SubTransactions, 2)
		assert.Equal(t, schema.PrefillSubTransactionResponse{Name: "Milk 2%", Label: "dairy", Amount: domain.Money(349)}, resp.SubTransactions[0])
		assert.Equal(t, "Bread", resp.SubTransactions[1].Name)
	})

	t.Run("the reader receives the absolute content URL under /api/v1", func(t *testing.T) {
		reader := &stubExtractionReader{result: &transaction_attachment_extraction.TransactionPrefillResult{}}
		f := newExtractionFixture(reader)
		draft := seedImageDraft(f.repo)

		rec := postExtraction(f, draft.ID)
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

		require.Equal(t, 1, reader.calls)
		require.Len(t, reader.lastReq.ImageURLs, 1)
		got := reader.lastReq.ImageURLs[0]
		wantPrefix := extractionBaseURL + "/api/v1/attachment-content?grant="
		require.True(t, strings.HasPrefix(got, wantPrefix),
			"reader URL %q must start with %q", got, wantPrefix)
		parsed, err := url.Parse(got)
		require.NoError(t, err)
		assert.Equal(t, extractionBaseURL, parsed.Scheme+"://"+parsed.Host)
		assert.Equal(t, "/api/v1/attachment-content", parsed.Path)
		assert.NotEmpty(t, parsed.Query().Get("grant"))
	})

	t.Run("zero proposed items serialize as an empty array, not null", func(t *testing.T) {
		reader := &stubExtractionReader{result: &transaction_attachment_extraction.TransactionPrefillResult{}}
		f := newExtractionFixture(reader)
		draft := seedImageDraft(f.repo)

		rec := postExtraction(f, draft.ID)

		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		body := rec.Body.String()
		assert.Contains(t, body, `"sub_transactions":[]`)
		assert.NotContains(t, body, "null")
	})

	t.Run("multiple drafts can be extracted independently without cross-file state", func(t *testing.T) {
		reader := &stubExtractionReader{result: &transaction_attachment_extraction.TransactionPrefillResult{
			Name: &transaction_attachment_extraction.PrefillField[string]{Value: "Receipt"},
		}}
		f := newExtractionFixture(reader)
		first := seedImageDraft(f.repo)
		second := seedImageDraft(f.repo)

		firstRec := postExtraction(f, first.ID)
		secondRec := postExtraction(f, second.ID)

		require.Equal(t, http.StatusOK, firstRec.Code, firstRec.Body.String())
		require.Equal(t, http.StatusOK, secondRec.Code, secondRec.Body.String())
		var firstResp schema.TransactionPrefillResponse
		var secondResp schema.TransactionPrefillResponse
		decodeData(t, firstRec, &firstResp)
		decodeData(t, secondRec, &secondResp)
		assert.Equal(t, first.ID, firstResp.AttachmentID)
		assert.Equal(t, second.ID, secondResp.AttachmentID)
		assert.Equal(t, 2, reader.calls)
	})

	t.Run("multiple drafts are sent to one combined reader call", func(t *testing.T) {
		reader := &stubExtractionReader{result: &transaction_attachment_extraction.TransactionPrefillResult{
			Name: &transaction_attachment_extraction.PrefillField[string]{Value: "Combined receipt"},
		}}
		f := newExtractionFixture(reader)
		first := seedImageDraft(f.repo)
		second := seedImageDraft(f.repo)

		rec := postBatchExtraction(f, []uuid.UUID{first.ID, second.ID})

		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		var response schema.TransactionPrefillResponse
		decodeData(t, rec, &response)
		assert.Equal(t, []uuid.UUID{first.ID, second.ID}, response.AttachmentIDs)
		assert.Equal(t, 1, reader.calls, "all files must share one vision-model request")
		require.Len(t, reader.lastReq.ImageURLs, 2)
		assert.Equal(t, []string{"Groceries"}, reader.lastReq.ExpenseCategories)
		assert.Equal(t, []string{"Salary"}, reader.lastReq.IncomeCategories)
	})

	t.Run("absent scalar fields are omitted rather than serialized null", func(t *testing.T) {
		reader := &stubExtractionReader{result: &transaction_attachment_extraction.TransactionPrefillResult{}}
		f := newExtractionFixture(reader)
		draft := seedImageDraft(f.repo)

		rec := postExtraction(f, draft.ID)

		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		body := rec.Body.String()
		for _, field := range []string{"type", "amount", "currency_code", "occurred_on", "name", "description", "category_hint"} {
			assert.NotContainsf(t, body, `"`+field+`"`, "%s must be omitted when the model proposed nothing", field)
		}
	})

	t.Run("a non-image draft is 400 and the reader is never called", func(t *testing.T) {
		reader := &stubExtractionReader{result: &transaction_attachment_extraction.TransactionPrefillResult{}}
		f := newExtractionFixture(reader)
		pdf := seedDraft(f.repo) // seedDraft is a PDF, non-image

		rec := postExtraction(f, pdf.ID)

		assert.Equal(t, http.StatusBadRequest, rec.Code, rec.Body.String())
		assert.Zero(t, reader.calls, "the provider must never be called for a non-image")
	})

	t.Run("a reader failure is a 503", func(t *testing.T) {
		reader := &stubExtractionReader{err: assert.AnError}
		f := newExtractionFixture(reader)
		draft := seedImageDraft(f.repo)

		rec := postExtraction(f, draft.ID)

		assert.Equal(t, http.StatusServiceUnavailable, rec.Code, rec.Body.String())
		assert.Equal(t, 1, reader.calls)
	})

	t.Run("a missing draft is a 404 and the reader is never called", func(t *testing.T) {
		reader := &stubExtractionReader{result: &transaction_attachment_extraction.TransactionPrefillResult{}}
		f := newExtractionFixture(reader)

		rec := postExtraction(f, uuid.New())

		assert.Equal(t, http.StatusNotFound, rec.Code, rec.Body.String())
		assert.Zero(t, reader.calls)
	})

	t.Run("a malformed attachment id is 400", func(t *testing.T) {
		reader := &stubExtractionReader{result: &transaction_attachment_extraction.TransactionPrefillResult{}}
		f := newExtractionFixture(reader)

		req := httptest.NewRequest(http.MethodPost, "/v1/transaction-attachments/not-a-uuid/extraction", nil)
		rec := httptest.NewRecorder()
		f.router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Zero(t, reader.calls)
	})
}
