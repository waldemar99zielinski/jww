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

// CurrencyExchangeRateFlatten is the code-only projection of a rate used by the
// scheduled sync path: the shape an external provider supplies and the shape
// the currency_exchange_rates table stores, with no joined Currency values.
// CurrencyExchangeRate (which carries nested Currency values) remains the
// read/HTTP shape. Both enforce the same two business-rule CHECK constraints
// the migration declares (distinct currencies, positive rate).
type CurrencyExchangeRateFlatten struct {
	FromCurrencyCode string
	ToCurrencyCode   string
	Rate             ExchangeRate
	RecordedAt       time.Time
}

func (r CurrencyExchangeRateFlatten) Validate() error {
	if r.FromCurrencyCode == r.ToCurrencyCode {
		return fmt.Errorf("from_currency_code and to_currency_code must differ")
	}
	if r.Rate <= 0 {
		return fmt.Errorf("rate must be positive")
	}
	return nil
}
