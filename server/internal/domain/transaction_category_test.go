package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func validTransactionCategory() TransactionCategory {
	return TransactionCategory{
		Type:      TransactionCategoryTypeExpense,
		Name:      "Groceries",
		Active:    true,
		SortOrder: 0,
	}
}

func TestTransactionCategory_Validate(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(c TransactionCategory) TransactionCategory
		wantErr string
	}{
		{"valid", func(c TransactionCategory) TransactionCategory { return c }, ""},
		{"blank name", func(c TransactionCategory) TransactionCategory { c.Name = "  "; return c }, "name"},
		{
			"negative sort order",
			func(c TransactionCategory) TransactionCategory { c.SortOrder = -1; return c },
			"sort_order",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.mutate(validTransactionCategory()).Validate()
			if tt.wantErr == "" {
				assert.NoError(t, err)
				return
			}
			assert.ErrorContains(t, err, tt.wantErr)
		})
	}
}
