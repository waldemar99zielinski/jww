package transaction

import (
	"fmt"
	"time"

	"github.com/google/uuid"

	"jww/internal/domain"
)

// Filter names the exact transaction or related-resource fields used by the
// repository query. Related domain structs are deliberately not accepted:
// each supported nested field is represented explicitly.
type Filter struct {
	UserID       *uuid.UUID
	CategoryID   *uuid.UUID
	Type         *domain.TransactionType
	CurrencyCode string
	Source       *domain.TransactionSource
	CreatedFrom  *time.Time
	CreatedTo    *time.Time
}

func (f Filter) Validate() error {
	if f.CreatedFrom != nil && f.CreatedTo != nil && f.CreatedFrom.After(*f.CreatedTo) {
		return fmt.Errorf("created_from must not be after created_to")
	}
	return nil
}
