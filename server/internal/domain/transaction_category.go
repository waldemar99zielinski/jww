package domain

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// TransactionCategory groups transactions by type (income/expense) for
// display and reporting.
type TransactionCategory struct {
	ID        uuid.UUID
	Type      TransactionCategoryType
	Name      string
	Active    bool
	SortOrder int
}

func (c TransactionCategory) Validate() error {
	if strings.TrimSpace(c.Name) == "" {
		return fmt.Errorf("transaction category name must not be blank")
	}
	if c.SortOrder < 0 {
		return fmt.Errorf("sort_order must not be negative")
	}
	return nil
}
