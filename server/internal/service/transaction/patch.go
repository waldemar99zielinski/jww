package transaction

import (
	"time"

	"github.com/google/uuid"

	"jww/internal/domain"
)

// Patch is a partial update for a transaction. A nil field means "leave
// unchanged"; a non-nil field means "set to this value". Related resources are
// referenced by their keys (currency code, category/user id) rather than by
// domain structs, so the workflow can resolve and validate them before writing.
//
// ID is deliberately absent: identity is immutable, and attachment membership is
// managed by the attachment routes.
type Patch struct {
	Type         *domain.TransactionType
	Amount       *domain.Money
	CurrencyCode *string
	CategoryID   *uuid.UUID
	UserID       *uuid.UUID
	Source       *domain.TransactionSource
	Name         *string
	Description  *string
	CreatedAt    *time.Time
}

// Empty reports whether the patch would change nothing, so callers can skip the
// write and return the current state.
func (p Patch) Empty() bool {
	return p.Type == nil &&
		p.Amount == nil &&
		p.CurrencyCode == nil &&
		p.CategoryID == nil &&
		p.UserID == nil &&
		p.Source == nil &&
		p.Name == nil &&
		p.Description == nil &&
		p.CreatedAt == nil
}
