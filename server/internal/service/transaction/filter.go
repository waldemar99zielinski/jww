package transaction

import (
	"fmt"
	"time"

	"jww/internal/domain"
)

type Filter struct {
	User        *domain.User
	Category    *domain.TransactionCategory
	Type        *domain.TransactionType
	Currency    *domain.Currency
	Source      *domain.TransactionSource
	CreatedFrom *time.Time
	CreatedTo   *time.Time
}

func (f Filter) Validate() error {
	if f.CreatedFrom != nil && f.CreatedTo != nil && f.CreatedFrom.After(*f.CreatedTo) {
		return fmt.Errorf("created_from must not be after created_to")
	}
	if f.Category != nil && f.Type != nil && string(f.Category.Type) != string(*f.Type) {
		return fmt.Errorf("category type must match the filter's transaction type")
	}
	return nil
}
