package transaction

import "fmt"

// SortKey names a transaction field the list can be ordered by.
type SortKey string

const (
	// SortKeyDate orders by the transaction's date (created_at).
	SortKeyDate SortKey = "date"
)

// SortOrder is the direction a list is ordered in.
type SortOrder string

const (
	SortOrderAsc  SortOrder = "asc"
	SortOrderDesc SortOrder = "desc"
)

// Sort names the list ordering. Empty fields take the defaults
// (SortKeyDate, SortOrderDesc) when Normalize is applied.
type Sort struct {
	Key   SortKey
	Order SortOrder
}

// Normalize fills in the default date-descending ordering for empty fields.
func (s Sort) Normalize() Sort {
	if s.Key == "" {
		s.Key = SortKeyDate
	}
	if s.Order == "" {
		s.Order = SortOrderDesc
	}
	return s
}

func (s Sort) Validate() error {
	switch s.Key {
	case "", SortKeyDate:
	default:
		return fmt.Errorf("sort_key must be 'date'")
	}
	switch s.Order {
	case "", SortOrderAsc, SortOrderDesc:
	default:
		return fmt.Errorf("sort_order must be 'asc' or 'desc'")
	}
	return nil
}
