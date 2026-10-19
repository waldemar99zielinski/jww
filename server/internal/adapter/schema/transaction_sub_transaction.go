package schema

import (
	"fmt"
	"strings"

	"github.com/google/uuid"

	"jww/internal/domain"
)

// TransactionSubTransactionResponse is the public view of one line item under
// a transaction. Amount relies on domain.Money's own marshaling, so it goes out
// as a bare integer number of cents (1234), the same convention as
// TransactionResponse.amount.
type TransactionSubTransactionResponse struct {
	ID     uuid.UUID    `json:"id"`
	Name   string       `json:"name"`
	Label  string       `json:"label"`
	Amount domain.Money `json:"amount"`
}

// TransactionSubTransactionCreateRequest is the body of POST
// /transactions/{id}/sub-transactions and one element of the create
// transaction's sub_transactions array. Label is free text and may be empty.
type TransactionSubTransactionCreateRequest struct {
	Name   string       `json:"name"`
	Label  string       `json:"label"`
	Amount domain.Money `json:"amount"`
}

func (r TransactionSubTransactionCreateRequest) Validate() error {
	if strings.TrimSpace(r.Name) == "" {
		return fmt.Errorf("name must not be blank")
	}
	if r.Amount == 0 {
		return fmt.Errorf("amount must not be zero")
	}
	return nil
}

// TransactionSubTransactionUpdateRequest is a partial update for PATCH
// /transactions/{id}/sub-transactions/{id}. An absent (or null) field is left
// unchanged; a present label may be empty (clearing it is a real value).
type TransactionSubTransactionUpdateRequest struct {
	Name   *string       `json:"name"`
	Label  *string       `json:"label"`
	Amount *domain.Money `json:"amount"`
}

func (r TransactionSubTransactionUpdateRequest) Validate() error {
	if r.Name != nil && strings.TrimSpace(*r.Name) == "" {
		return fmt.Errorf("name must not be blank")
	}
	if r.Amount != nil && *r.Amount == 0 {
		return fmt.Errorf("amount must not be zero")
	}
	return nil
}

// TransactionSubTransactionsToResponse converts a slice, always returning a
// non-nil (possibly empty) slice so JSON encodes [] rather than null.
func TransactionSubTransactionsToResponse(subTransactions []domain.TransactionSubTransaction) []TransactionSubTransactionResponse {
	out := make([]TransactionSubTransactionResponse, len(subTransactions))
	for i, s := range subTransactions {
		out[i] = TransactionSubTransactionResponse{
			ID:     s.ID,
			Name:   s.Name,
			Label:  s.Label,
			Amount: s.Amount,
		}
	}
	return out
}
