package domain

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// NewSubTransaction is one line item supplied alongside a transaction create or
// through a sub-transaction CRUD call. Amount is a non-zero minor-unit value;
// negative amounts represent discounts/coupons. The derived parent total must
// still be positive, and that rule is enforced by Transaction.Validate.
type NewSubTransaction struct {
	Name   string
	Label  string
	Amount Money
}

// SubTransactionPatch is a partial edit of one line item; a nil field means
// "leave unchanged". The HTTP layer translates its request DTO into this.
type SubTransactionPatch struct {
	Name   *string
	Label  *string
	Amount *Money
}

// TransactionSubTransaction is one line item under a transaction: a product on
// a receipt, or one split part of an expense. TransactionID is a plain int64 (a
// Postgres bigint identity column), unlike the UUID-keyed entities.
//
// Label is free text assigned by the user or an extraction run (e.g. "dairy").
// It is deliberately not a TransactionCategory FK — per-item categorisation is
// out of scope.
type TransactionSubTransaction struct {
	ID            uuid.UUID
	TransactionID int64
	Name          string
	Label         string
	Amount        Money
	CreatedAt     time.Time
}

// Validate enforces the row's own invariants, mirroring the table's CHECK
// constraints: a non-blank name and a non-zero amount. Negative amounts are
// legitimate discount, coupon, or adjustment lines; the parent transaction's
// total must still be positive.
func (s TransactionSubTransaction) Validate() error {
	if strings.TrimSpace(s.Name) == "" {
		return fmt.Errorf("name must not be blank")
	}
	if s.Amount == 0 {
		return fmt.Errorf("amount must not be zero")
	}
	return nil
}
