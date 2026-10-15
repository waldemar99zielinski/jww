package domain

import (
	"fmt"
	"time"
)

// CurrencyExchangeRate is a point-in-time conversion rate between two
// currencies.
type CurrencyExchangeRate struct {
	FromCurrency Currency
	ToCurrency   Currency
	Rate         ExchangeRate
	RecordedAt   time.Time
}

func (r CurrencyExchangeRate) Validate() error {
	if r.FromCurrency.Code == r.ToCurrency.Code {
		return fmt.Errorf("from_currency and to_currency must differ")
	}
	if r.Rate <= 0 {
		return fmt.Errorf("rate must be positive")
	}
	return nil
}
