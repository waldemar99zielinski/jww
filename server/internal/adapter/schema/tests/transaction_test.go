package schema_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

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
}
