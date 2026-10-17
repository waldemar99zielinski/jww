package currency_exchange_rate

import (
	"fmt"
	"time"
)

// Filter names the exact exchange-rate fields used by the repository query.
// Currency relationships are represented by their codes, never nested domain
// structs.
type Filter struct {
	FromCurrencyCode string
	ToCurrencyCode   string
	RecordedFrom     *time.Time
	RecordedTo       *time.Time
}

func (f Filter) Validate() error {
	if f.RecordedFrom != nil && f.RecordedTo != nil && f.RecordedFrom.After(*f.RecordedTo) {
		return fmt.Errorf("recorded_from must not be after recorded_to")
	}
	return nil
}
