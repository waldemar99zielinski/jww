package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func validTransaction() Transaction {
	return Transaction{
		Type:     TransactionTypeExpense,
		Amount:   Money(1500),
		Currency: Currency{Code: "USD", Name: "US Dollar"},
		Category: TransactionCategory{Type: TransactionCategoryTypeExpense, Name: "Groceries"},
		User:     User{Username: "wozi", PasswordHash: "hash", DisplayName: "Wozi"},
		Source:   TransactionSourceManual,
		Name:     "Weekly shop",
	}
}

func TestTransaction_Validate(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(tr Transaction) Transaction
		wantErr string
	}{
		{"valid", func(tr Transaction) Transaction { return tr }, ""},
		{"zero amount", func(tr Transaction) Transaction { tr.Amount = 0; return tr }, "amount"},
		{"negative amount", func(tr Transaction) Transaction { tr.Amount = -500; return tr }, "amount"},
		{
			"category type does not match transaction type",
			func(tr Transaction) Transaction { tr.Category.Type = TransactionCategoryTypeIncome; return tr },
			"category type",
		},
		{"blank name", func(tr Transaction) Transaction { tr.Name = "   "; return tr }, "name"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.mutate(validTransaction()).Validate()
			if tt.wantErr == "" {
				assert.NoError(t, err)
				return
			}
			assert.ErrorContains(t, err, tt.wantErr)
		})
	}
}
