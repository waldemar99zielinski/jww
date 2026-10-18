package schema_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"jww/internal/adapter/schema"
	"jww/internal/domain"
)

func validTransactionCreateRequest() schema.TransactionCreateRequest {
	return schema.TransactionCreateRequest{
		Type:         domain.TransactionTypeExpense,
		Amount:       domain.Money(1000),
		CurrencyCode: "USD",
		CategoryID:   uuid.New(),
		UserID:       uuid.New(),
		Source:       domain.TransactionSourceManual,
		Name:         "Groceries",
		Description:  "",
	}
}

func TestTransactionCreateRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(r *schema.TransactionCreateRequest)
		wantErr bool
	}{
		{"valid", func(r *schema.TransactionCreateRequest) {}, false},
		{"invalid type", func(r *schema.TransactionCreateRequest) { r.Type = "bogus" }, true},
		{"zero amount", func(r *schema.TransactionCreateRequest) { r.Amount = 0 }, true},
		{"negative amount", func(r *schema.TransactionCreateRequest) { r.Amount = -1 }, true},
		{"invalid source", func(r *schema.TransactionCreateRequest) { r.Source = "bogus" }, true},
		{"currency code wrong length", func(r *schema.TransactionCreateRequest) { r.CurrencyCode = "US" }, true},
		{"blank name", func(r *schema.TransactionCreateRequest) { r.Name = "" }, true},
		{"duplicate attachment ids", func(r *schema.TransactionCreateRequest) {
			id := uuid.New()
			r.AttachmentIDs = []uuid.UUID{id, id}
		}, true},
		{"empty attachment id", func(r *schema.TransactionCreateRequest) {
			r.AttachmentIDs = []uuid.UUID{uuid.Nil}
		}, true},
		{"optional attachment ids accepted", func(r *schema.TransactionCreateRequest) {
			r.AttachmentIDs = []uuid.UUID{uuid.New()}
		}, false},
		{"absent created_at accepted", func(r *schema.TransactionCreateRequest) { r.CreatedAt = nil }, false},
		{"explicit created_at accepted", func(r *schema.TransactionCreateRequest) {
			when := time.Date(2024, 3, 15, 12, 30, 0, 0, time.UTC)
			r.CreatedAt = &when
		}, false},
		{"zero created_at rejected", func(r *schema.TransactionCreateRequest) {
			zero := time.Time{}
			r.CreatedAt = &zero
		}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := validTransactionCreateRequest()
			tt.mutate(&req)
			err := req.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ptr returns a pointer to v, for building optional request fields.
func ptr[T any](v T) *T { return &v }

func TestTransactionUpdateRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     schema.TransactionUpdateRequest
		wantErr bool
	}{
		{"empty patch is valid", schema.TransactionUpdateRequest{}, false},
		{"valid type", schema.TransactionUpdateRequest{Type: ptr(domain.TransactionTypeIncome)}, false},
		{"invalid type", schema.TransactionUpdateRequest{Type: ptr(domain.TransactionType("bogus"))}, true},
		{"valid amount", schema.TransactionUpdateRequest{Amount: ptr(domain.Money(1))}, false},
		{"zero amount", schema.TransactionUpdateRequest{Amount: ptr(domain.Money(0))}, true},
		{"negative amount", schema.TransactionUpdateRequest{Amount: ptr(domain.Money(-5))}, true},
		{"invalid source", schema.TransactionUpdateRequest{Source: ptr(domain.TransactionSource("bogus"))}, true},
		{"currency too short", schema.TransactionUpdateRequest{CurrencyCode: ptr("US")}, true},
		{"currency ok", schema.TransactionUpdateRequest{CurrencyCode: ptr("USD")}, false},
		{"blank name", schema.TransactionUpdateRequest{Name: ptr("")}, true},
		{"name ok", schema.TransactionUpdateRequest{Name: ptr("Rent")}, false},
		{"empty description allowed", schema.TransactionUpdateRequest{Description: ptr("")}, false},
		{"valid created_at", schema.TransactionUpdateRequest{CreatedAt: ptr(time.Date(2024, 3, 15, 12, 30, 0, 0, time.UTC))}, false},
		{"zero created_at", schema.TransactionUpdateRequest{CreatedAt: ptr(time.Time{})}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestTransactionUpdateRequest_JSONPresence proves absent, null, and present
// fields are distinguished: only present, non-null fields become non-nil.
func TestTransactionUpdateRequest_JSONPresence(t *testing.T) {
	var req schema.TransactionUpdateRequest
	require.NoError(t, json.Unmarshal([]byte(`{"amount":"12.34","description":"","category_id":null,"created_at":"2024-03-15T12:30:00Z"}`), &req))

	require.NotNil(t, req.Amount)
	assert.Equal(t, domain.Money(1234), *req.Amount, "a quoted decimal must parse as money")
	require.NotNil(t, req.Description)
	assert.Equal(t, "", *req.Description, "an explicit empty string is a real value")
	require.NotNil(t, req.CreatedAt)
	assert.True(t, time.Date(2024, 3, 15, 12, 30, 0, 0, time.UTC).Equal(*req.CreatedAt))
	assert.Nil(t, req.CategoryID, "explicit null means unchanged")
	assert.Nil(t, req.Name, "an absent field means unchanged")
	assert.Nil(t, req.Type)
}

func TestTransactionToResponse(t *testing.T) {
	categoryID := uuid.New()
	userID := uuid.New()
	createdAt := time.Date(2026, 2, 10, 9, 30, 0, 0, time.UTC)

	tx := domain.Transaction{
		ID:          42,
		Type:        domain.TransactionTypeExpense,
		Amount:      domain.Money(1234),
		Currency:    domain.Currency{Code: "USD", Name: "US Dollar"},
		Category:    domain.TransactionCategory{ID: categoryID, Name: "Groceries"},
		User:        domain.User{ID: userID},
		Source:      domain.TransactionSourceManual,
		Name:        "Weekly shop",
		Description: "at the market",
		CreatedAt:   createdAt,
	}

	resp := schema.TransactionToResponse(tx)

	assert.Equal(t, int64(42), resp.ID)
	assert.Equal(t, domain.TransactionTypeExpense, resp.Type)
	assert.Equal(t, domain.Money(1234), resp.Amount)
	assert.Equal(t, "USD", resp.CurrencyCode)
	assert.Equal(t, categoryID, resp.CategoryID)
	assert.Equal(t, "Groceries", resp.CategoryName)
	assert.Equal(t, userID, resp.UserID)
	assert.Equal(t, domain.TransactionSourceManual, resp.Source)
	assert.Equal(t, "Weekly shop", resp.Name)
	assert.Equal(t, "at the market", resp.Description)
	assert.True(t, createdAt.Equal(resp.CreatedAt))
	assert.NotNil(t, resp.Attachments, "attachments must never be null in JSON")
	assert.Empty(t, resp.Attachments)
}

func TestTransactionToResponse_IncludesAttachments(t *testing.T) {
	txID := int64(9)
	tx := domain.Transaction{
		ID: 1,
		Attachments: []domain.TransactionAttachment{
			{ID: uuid.New(), TransactionID: &txID, OriginalFilename: "a.pdf", MediaType: "application/pdf", ByteSize: 3},
		},
	}

	resp := schema.TransactionToResponse(tx)

	require.Len(t, resp.Attachments, 1)
	assert.Equal(t, "a.pdf", resp.Attachments[0].OriginalFilename)
	require.NotNil(t, resp.Attachments[0].TransactionID)
	assert.Equal(t, txID, *resp.Attachments[0].TransactionID)
}

// TestTransactionAttachmentsToResponse_DoesNotLeakStorageFields proves the
// public DTO never exposes storage_key/storage_version/etag.
func TestTransactionAttachmentsToResponse_DoesNotLeakStorageFields(t *testing.T) {
	raw, err := json.Marshal(schema.TransactionAttachmentToResponse(domain.TransactionAttachment{
		ID:             uuid.New(),
		StorageKey:     "jww/transaction-attachments/secret-key",
		StorageVersion: "v42",
		ETag:           "etag-secret",
		MediaType:      "application/pdf",
		ByteSize:       1,
	}))
	require.NoError(t, err)
	assert.NotContains(t, string(raw), "secret")
	assert.NotContains(t, string(raw), "storage_key")
	assert.NotContains(t, string(raw), "storage_version")
	assert.NotContains(t, string(raw), "etag")
}
