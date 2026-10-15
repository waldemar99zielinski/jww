package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func validCurrencyExchangeRate() CurrencyExchangeRate {
	return CurrencyExchangeRate{
		FromCurrency: Currency{Code: "USD", Name: "US Dollar"},
		ToCurrency:   Currency{Code: "EUR", Name: "Euro"},
		Rate:         ExchangeRate(10_000), // 1.0000
		RecordedAt:   time.Now().UTC(),
	}
}

func TestCurrencyExchangeRate_Validate(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(r CurrencyExchangeRate) CurrencyExchangeRate
		wantErr string
	}{
		{"valid", func(r CurrencyExchangeRate) CurrencyExchangeRate { return r }, ""},
		{
			"same currency on both sides",
			func(r CurrencyExchangeRate) CurrencyExchangeRate { r.ToCurrency = r.FromCurrency; return r },
			"differ",
		},
		{"zero rate", func(r CurrencyExchangeRate) CurrencyExchangeRate { r.Rate = 0; return r }, "rate"},
		{"negative rate", func(r CurrencyExchangeRate) CurrencyExchangeRate { r.Rate = -1; return r }, "rate"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.mutate(validCurrencyExchangeRate()).Validate()
			if tt.wantErr == "" {
				assert.NoError(t, err)
				return
			}
			assert.ErrorContains(t, err, tt.wantErr)
		})
	}
}
