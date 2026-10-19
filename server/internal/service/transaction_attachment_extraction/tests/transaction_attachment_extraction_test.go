package transaction_attachment_extraction_test

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"jww/internal/domain"
	"jww/internal/service/transaction_attachment_extraction"
)

func newAppContext() domain.AppContext {
	var appCtx domain.AppContext
	appCtx.SetContext(context.Background())
	return appCtx
}

// fakeSource is a hand-written transaction_attachment_extraction.AttachmentSource.
type fakeSource struct {
	attachment   *domain.TransactionAttachment
	draftErr     error
	grant        string
	grantErr     error
	draftCalls   int
	grantCalls   int
	lastGrantID  uuid.UUID
	lastDraftOrd uuid.UUID
}

func (s *fakeSource) DraftAttachment(_ domain.AppContext, id uuid.UUID) (*domain.TransactionAttachment, error) {
	s.draftCalls++
	s.lastDraftOrd = id
	if s.draftErr != nil {
		return nil, s.draftErr
	}
	return s.attachment, nil
}

func (s *fakeSource) ContentGrant(_ domain.AppContext, id uuid.UUID) (string, error) {
	s.grantCalls++
	s.lastGrantID = id
	if s.grantErr != nil {
		return "", s.grantErr
	}
	return s.grant, nil
}

// fakeExtractor is a hand-written transaction_attachment_extraction.Extractor.
type fakeExtractor struct {
	result      *transaction_attachment_extraction.TransactionPrefillResult
	err         error
	calls       int
	lastRequest transaction_attachment_extraction.TransactionPrefillInput
}

type fakeCategorySource struct {
	expense []string
	income  []string
	err     error
	calls   int
}

func (s *fakeCategorySource) TransactionCategories(_ domain.AppContext) (expense, income []string, err error) {
	s.calls++
	return s.expense, s.income, s.err
}

func (x *fakeExtractor) ExtractTransactionPrefill(_ domain.AppContext, req transaction_attachment_extraction.TransactionPrefillInput) (*transaction_attachment_extraction.TransactionPrefillResult, error) {
	x.calls++
	x.lastRequest = req
	if x.err != nil {
		return nil, x.err
	}
	return x.result, nil
}

const testBaseURL = "https://api.example.com"

func newService(extractor *fakeExtractor, source *fakeSource) *transaction_attachment_extraction.Service {
	return transaction_attachment_extraction.NewService(
		extractor,
		source,
		&fakeCategorySource{},
		transaction_attachment_extraction.Config{PublicBaseURL: testBaseURL},
	)
}

func imageAttachment(id uuid.UUID) *domain.TransactionAttachment {
	return &domain.TransactionAttachment{
		ID:               id,
		OriginalFilename: "receipt.jpg",
		MediaType:        "image/jpeg",
		ByteSize:         10,
		StorageKey:       id.String(),
		StorageVersion:   "v1",
		CreatedAt:        time.Now().UTC(),
	}
}

// TestService_Extract_NonImageIsUnsupportedBeforeAnyExtractorCall pins the
// resolution that non-image media is a validation failure checked BEFORE a
// grant is minted and before the vision extractor is called — the PDF case
// included, even though PDF is previewable elsewhere.
func TestService_Extract_NonImageIsUnsupportedBeforeAnyExtractorCall(t *testing.T) {
	for _, mediaType := range []string{"application/pdf", "text/plain", "application/octet-stream"} {
		t.Run(mediaType, func(t *testing.T) {
			id := uuid.New()
			att := imageAttachment(id)
			att.MediaType = mediaType
			source := &fakeSource{attachment: att, grant: "grant-token"}
			extractor := &fakeExtractor{result: &transaction_attachment_extraction.TransactionPrefillResult{}}
			svc := newService(extractor, source)

			_, err := svc.Extract(newAppContext(), id)

			require.Error(t, err)
			var domainErr domain.DomainError
			require.ErrorAs(t, err, &domainErr)
			assert.Equal(t, domain.ErrorKindValidation, domainErr.Kind())
			var unsupported *domain.UnsupportedAttachmentMediaTypeError
			assert.ErrorAs(t, err, &unsupported)
			assert.Zero(t, extractor.calls, "the extractor must never be called for a non-image")
			assert.Zero(t, source.grantCalls, "no grant may be minted for a non-image")
		})
	}
}

func TestService_Extract_DraftNotFoundPropagates(t *testing.T) {
	id := uuid.New()
	source := &fakeSource{draftErr: domain.NewTransactionAttachmentNotFoundError(id)}
	extractor := &fakeExtractor{}
	svc := newService(extractor, source)

	_, err := svc.Extract(newAppContext(), id)

	require.Error(t, err)
	var notFound *domain.TransactionAttachmentNotFoundError
	require.ErrorAs(t, err, &notFound)
	assert.Equal(t, id, notFound.AttachmentID)
	assert.Equal(t, 1, source.draftCalls)
	assert.Zero(t, source.grantCalls)
	assert.Zero(t, extractor.calls)
}

// TestService_Extract_BuildsAbsoluteGrantURL proves the extractor receives the
// absolute, URL-escaped content URL the vision adapter sends as image_url.
func TestService_Extract_BuildsAbsoluteGrantURL(t *testing.T) {
	id := uuid.New()
	// A token containing several characters that must be percent-escaped.
	grant := "abc+def/ghi=jkl?mno&pqr"
	source := &fakeSource{attachment: imageAttachment(id), grant: grant}
	extractor := &fakeExtractor{result: &transaction_attachment_extraction.TransactionPrefillResult{}}
	svc := newService(extractor, source)

	prefill, err := svc.Extract(newAppContext(), id)

	require.NoError(t, err)
	require.NotNil(t, prefill)
	assert.Equal(t, id, prefill.AttachmentID)
	require.Equal(t, 1, extractor.calls)
	want := testBaseURL + "/api/v1/attachment-content?grant=" + url.QueryEscape(grant)
	require.Len(t, extractor.lastRequest.ImageURLs, 1)
	assert.Equal(t, want, extractor.lastRequest.ImageURLs[0])
	assert.Equal(t, id, source.lastGrantID, "the grant is minted for the requested attachment")
}

func TestService_Extract_ExtractorFailureBecomesUnavailable(t *testing.T) {
	id := uuid.New()
	source := &fakeSource{attachment: imageAttachment(id), grant: "grant-token"}
	extractor := &fakeExtractor{err: errors.New("openai: unexpected status 500")}
	svc := newService(extractor, source)

	_, err := svc.Extract(newAppContext(), id)

	require.Error(t, err)
	var domainErr domain.DomainError
	require.ErrorAs(t, err, &domainErr)
	assert.Equal(t, domain.ErrorKindUnavailable, domainErr.Kind(), "a provider failure is a 503, never a 500")
	var failed *domain.AttachmentExtractionFailedError
	assert.ErrorAs(t, err, &failed)
}

func TestService_Extract_GrantFailurePropagates(t *testing.T) {
	id := uuid.New()
	source := &fakeSource{attachment: imageAttachment(id), grantErr: errors.New("token signing failed")}
	extractor := &fakeExtractor{}
	svc := newService(extractor, source)

	_, err := svc.Extract(newAppContext(), id)

	require.Error(t, err)
	assert.Zero(t, extractor.calls, "the extractor is never reached without a content URL")
}

func TestService_ExtractMany_ValidatesTheWholeGroupBeforeMintingGrants(t *testing.T) {
	firstID := uuid.New()
	secondID := uuid.New()
	first := imageAttachment(firstID)
	second := imageAttachment(secondID)
	second.MediaType = "application/pdf"
	source := &orderedFakeSource{attachments: map[uuid.UUID]*domain.TransactionAttachment{
		firstID:  first,
		secondID: second,
	}}
	extractor := &fakeExtractor{}
	svc := transaction_attachment_extraction.NewService(extractor, source, &fakeCategorySource{}, transaction_attachment_extraction.Config{PublicBaseURL: testBaseURL})

	_, err := svc.ExtractMany(newAppContext(), []uuid.UUID{firstID, secondID})

	require.Error(t, err)
	assert.Zero(t, source.grantCalls, "no grants are minted unless every image passes validation")
	assert.Zero(t, extractor.calls)
}

func TestService_ExtractMany_SendsEveryImageOnceInRequestOrder(t *testing.T) {
	firstID := uuid.New()
	secondID := uuid.New()
	source := &orderedFakeSource{attachments: map[uuid.UUID]*domain.TransactionAttachment{
		firstID:  imageAttachment(firstID),
		secondID: imageAttachment(secondID),
	}}
	extractor := &fakeExtractor{result: &transaction_attachment_extraction.TransactionPrefillResult{}}
	categories := &fakeCategorySource{expense: []string{"Groceries"}, income: []string{"Salary"}}
	svc := transaction_attachment_extraction.NewService(extractor, source, categories, transaction_attachment_extraction.Config{PublicBaseURL: testBaseURL})

	prefill, err := svc.ExtractMany(newAppContext(), []uuid.UUID{firstID, secondID})

	require.NoError(t, err)
	assert.Equal(t, []uuid.UUID{firstID, secondID}, prefill.AttachmentIDs)
	assert.Equal(t, firstID, prefill.AttachmentID, "the compatibility field carries the first source id")
	assert.Equal(t, 2, source.grantCalls)
	require.Equal(t, 1, extractor.calls)
	assert.Equal(t, []string{"Groceries"}, extractor.lastRequest.ExpenseCategories)
	assert.Equal(t, []string{"Salary"}, extractor.lastRequest.IncomeCategories)
	assert.Equal(t, []string{
		testBaseURL + "/api/v1/attachment-content?grant=" + url.QueryEscape("grant-"+firstID.String()),
		testBaseURL + "/api/v1/attachment-content?grant=" + url.QueryEscape("grant-"+secondID.String()),
	}, extractor.lastRequest.ImageURLs)
}

func TestService_ExtractMany_CategoryFailureStopsBeforeGrantsAndExtractor(t *testing.T) {
	id := uuid.New()
	source := &fakeSource{attachment: imageAttachment(id), grant: "grant-token"}
	categoryErr := errors.New("categories unavailable")
	categories := &fakeCategorySource{err: categoryErr}
	extractor := &fakeExtractor{}
	svc := transaction_attachment_extraction.NewService(extractor, source, categories, transaction_attachment_extraction.Config{PublicBaseURL: testBaseURL})

	_, err := svc.ExtractMany(newAppContext(), []uuid.UUID{id})

	require.ErrorIs(t, err, categoryErr)
	assert.Equal(t, 1, categories.calls)
	assert.Zero(t, source.grantCalls)
	assert.Zero(t, extractor.calls)
}

func TestService_ExtractMany_RejectsInvalidIDListsBeforeLookup(t *testing.T) {
	id := uuid.New()
	tests := map[string][]uuid.UUID{
		"empty":     {},
		"nil id":    {uuid.Nil},
		"duplicate": {id, id},
		"too many":  make([]uuid.UUID, 11),
	}
	for name, ids := range tests {
		t.Run(name, func(t *testing.T) {
			source := &fakeSource{attachment: imageAttachment(id), grant: "grant-token"}
			extractor := &fakeExtractor{}

			_, err := newService(extractor, source).ExtractMany(newAppContext(), ids)

			require.Error(t, err)
			assert.Zero(t, source.draftCalls)
			assert.Zero(t, source.grantCalls)
			assert.Zero(t, extractor.calls)
		})
	}
}

type orderedFakeSource struct {
	attachments map[uuid.UUID]*domain.TransactionAttachment
	grantCalls  int
}

func (s *orderedFakeSource) DraftAttachment(_ domain.AppContext, id uuid.UUID) (*domain.TransactionAttachment, error) {
	attachment, ok := s.attachments[id]
	if !ok {
		return nil, domain.NewTransactionAttachmentNotFoundError(id)
	}
	return attachment, nil
}

func (s *orderedFakeSource) ContentGrant(_ domain.AppContext, id uuid.UUID) (string, error) {
	s.grantCalls++
	return "grant-" + id.String(), nil
}

// TestService_Extract_MapsExtractorResultIntoPrefill covers the mapping from the
// extractor's structured result to the service-level Prefill, including the new
// sub-transaction list (free-text label, no confidence threshold).
func TestService_Extract_MapsExtractorResultIntoPrefill(t *testing.T) {
	id := uuid.New()
	source := &fakeSource{attachment: imageAttachment(id), grant: "grant-token"}
	confidence := 87.0
	extractor := &fakeExtractor{result: &transaction_attachment_extraction.TransactionPrefillResult{
		Type: &transaction_attachment_extraction.PrefillField[domain.TransactionType]{
			Value: domain.TransactionTypeExpense, Confidence: &confidence, Source: "receipt header",
		},
		Amount:       &transaction_attachment_extraction.PrefillField[domain.Money]{Value: domain.Money(2490)},
		CurrencyCode: &transaction_attachment_extraction.PrefillField[string]{Value: "USD"},
		OccurredAt:   &transaction_attachment_extraction.PrefillField[string]{Value: "14:35"},
		Name:         &transaction_attachment_extraction.PrefillField[string]{Value: "Corner Market"},
		SubTransactions: []transaction_attachment_extraction.PrefillSubTransaction{
			{Name: "Milk 2%", Label: "dairy", Amount: domain.Money(349)},
			{Name: "Bread", Label: "bakery", Amount: domain.Money(219)},
		},
	}}
	svc := newService(extractor, source)

	prefill, err := svc.Extract(newAppContext(), id)

	require.NoError(t, err)
	require.NotNil(t, prefill)
	assert.Equal(t, id, prefill.AttachmentID)
	require.NotNil(t, prefill.Amount)
	assert.Equal(t, domain.Money(2490), prefill.Amount.Value)
	require.NotNil(t, prefill.OccurredAt)
	assert.Equal(t, "14:35", prefill.OccurredAt.Value)
	require.NotNil(t, prefill.Name)
	assert.Equal(t, "Corner Market", prefill.Name.Value)
	require.Len(t, prefill.SubTransactions, 2)
	assert.Equal(t, transaction_attachment_extraction.PrefillSubTransaction{Name: "Milk 2%", Label: "dairy", Amount: domain.Money(349)}, prefill.SubTransactions[0])
	assert.Equal(t, "Bread", prefill.SubTransactions[1].Name)
}

// TestTransactionPrefillResult_String_RendersReadableLine pins the pretty-log shape: one
// readable line, only the proposed fields, each scalar with its optional
// confidence/source, and each line item with its optional label.
func TestTransactionPrefillResult_String_RendersReadableLine(t *testing.T) {
	confidence := 90.0
	result := transaction_attachment_extraction.TransactionPrefillResult{
		Type: &transaction_attachment_extraction.PrefillField[domain.TransactionType]{
			Value: domain.TransactionTypeExpense, Confidence: &confidence, Source: "receipt header",
		},
		Amount: &transaction_attachment_extraction.PrefillField[domain.Money]{Value: domain.Money(2490)},
		SubTransactions: []transaction_attachment_extraction.PrefillSubTransaction{
			{Name: "Milk 2%", Label: "dairy", Amount: domain.Money(349)},
			{Name: "Bread", Amount: domain.Money(219)},
		},
		Warnings: []string{"low confidence on amount"},
	}

	assert.Equal(t,
		"TransactionPrefillResult{type=expense (confidence 90, source: receipt header), amount=24.90, sub_transactions=[Milk 2% (dairy): 3.49, Bread: 2.19], warnings=[low confidence on amount]}",
		result.String(),
	)
}

// TestTransactionPrefillResult_String_OmitsAbsentFields covers the empty and nil cases: an
// empty result renders as an empty brace pair, and a nil result is left to fmt
// (which prints `<nil>`) rather than panicking.
func TestTransactionPrefillResult_String_OmitsAbsentFields(t *testing.T) {
	assert.Equal(t, "TransactionPrefillResult{}", transaction_attachment_extraction.TransactionPrefillResult{}.String())

	assert.Equal(t, "<nil>", fmt.Sprintf("%v", (*transaction_attachment_extraction.TransactionPrefillResult)(nil)))
}
