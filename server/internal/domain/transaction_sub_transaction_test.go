package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func validSubTransaction() TransactionSubTransaction {
	return TransactionSubTransaction{
		ID:            uuid.New(),
		TransactionID: 7,
		Name:          "Milk 2%",
		Label:         "dairy",
		Amount:        Money(349),
		CreatedAt:     time.Now().UTC(),
	}
}

// TestTransactionSubTransaction_Validate covers the entity's own invariants:
// a non-blank Name (mirroring the table's btrim(name) <> ” CHECK) and a
// non-zero Amount (mirroring amount <> 0). Negative values are discounts;
// Label is deliberately free-text and carries no validation.
func TestTransactionSubTransaction_Validate(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*TransactionSubTransaction)
		wantErr string // substring; "" = no error expected
	}{
		{"valid", func(s *TransactionSubTransaction) {}, ""},
		{"empty label is allowed", func(s *TransactionSubTransaction) { s.Label = "" }, ""},
		{"arbitrary free-text label is allowed", func(s *TransactionSubTransaction) { s.Label = "Meat & Proteins" }, ""},
		{"blank name", func(s *TransactionSubTransaction) { s.Name = "   " }, "name"},
		{"empty name", func(s *TransactionSubTransaction) { s.Name = "" }, "name"},
		{"zero amount", func(s *TransactionSubTransaction) { s.Amount = 0 }, "amount"},
		{"negative discount amount is allowed", func(s *TransactionSubTransaction) { s.Amount = -1 }, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := validSubTransaction()
			tt.mutate(&s)
			err := s.Validate()
			if tt.wantErr == "" {
				assert.NoError(t, err)
				return
			}
			assert.ErrorContains(t, err, tt.wantErr)
		})
	}
}

// TestTransaction_SubTransactionsField proves the parent entity carries the
// hydrated line items and that a Transaction with line items still satisfies
// its own invariant set (the parent amount rule is enforced by the workflow,
// not by Transaction.Validate).
func TestTransaction_SubTransactionsField(t *testing.T) {
	tx := validTransaction()
	tx.SubTransactions = []TransactionSubTransaction{
		validSubTransaction(),
		{ID: uuid.New(), TransactionID: 1, Name: "Bread", Amount: Money(199), CreatedAt: time.Now().UTC()},
	}

	require.NoError(t, tx.Validate())
	require.Len(t, tx.SubTransactions, 2)
	assert.Equal(t, "Milk 2%", tx.SubTransactions[0].Name)
	assert.Equal(t, Money(349), tx.SubTransactions[0].Amount)
	assert.Equal(t, "Bread", tx.SubTransactions[1].Name)
}
