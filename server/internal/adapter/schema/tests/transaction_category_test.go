package schema_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"jww/internal/adapter/schema"
	"jww/internal/domain"
)

func TestTransactionCategoryCreateRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     schema.TransactionCategoryCreateRequest
		wantErr bool
	}{
		{"valid", schema.TransactionCategoryCreateRequest{Type: domain.TransactionCategoryTypeExpense, Name: "Groceries", SortOrder: 0}, false},
		{"invalid type", schema.TransactionCategoryCreateRequest{Type: "bogus", Name: "Groceries", SortOrder: 0}, true},
		{"blank name", schema.TransactionCategoryCreateRequest{Type: domain.TransactionCategoryTypeExpense, Name: "", SortOrder: 0}, true},
		{"negative sort order", schema.TransactionCategoryCreateRequest{Type: domain.TransactionCategoryTypeExpense, Name: "Groceries", SortOrder: -1}, true},
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

func TestTransactionCategoryUpdateRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     schema.TransactionCategoryUpdateRequest
		wantErr bool
	}{
		{"valid", schema.TransactionCategoryUpdateRequest{Name: "Groceries", Active: true, SortOrder: 2}, false},
		{"blank name", schema.TransactionCategoryUpdateRequest{Name: "", Active: true, SortOrder: 0}, true},
		{"negative sort order", schema.TransactionCategoryUpdateRequest{Name: "Groceries", SortOrder: -5}, true},
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

func TestTransactionCategoryToResponse(t *testing.T) {
	id := uuid.New()
	c := domain.TransactionCategory{ID: id, Type: domain.TransactionCategoryTypeIncome, Name: "Salary", Active: true, SortOrder: 3}

	resp := schema.TransactionCategoryToResponse(c)

	assert.Equal(t, schema.TransactionCategoryResponse{
		ID: id, Type: domain.TransactionCategoryTypeIncome, Name: "Salary", Active: true, SortOrder: 3,
	}, resp)
}
