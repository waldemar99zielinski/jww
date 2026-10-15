package schema_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"jww/internal/adapter/schema"
	"jww/internal/domain"
)

func TestCurrencyExchangeRateCreateRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     schema.CurrencyExchangeRateCreateRequest
		wantErr bool
	}{
		{
			"valid",
			schema.CurrencyExchangeRateCreateRequest{FromCurrencyCode: "USD", ToCurrencyCode: "EUR", Rate: 10_000},
			false,
		},
		{
			"from code wrong length",
			schema.CurrencyExchangeRateCreateRequest{FromCurrencyCode: "US", ToCurrencyCode: "EUR", Rate: 10_000},
			true,
		},
		{
			"to code wrong length",
			schema.CurrencyExchangeRateCreateRequest{FromCurrencyCode: "USD", ToCurrencyCode: "EURO", Rate: 10_000},
			true,
		},
		{
			"zero rate",
			schema.CurrencyExchangeRateCreateRequest{FromCurrencyCode: "USD", ToCurrencyCode: "EUR", Rate: 0},
			true,
		},
		{
			"negative rate",
			schema.CurrencyExchangeRateCreateRequest{FromCurrencyCode: "USD", ToCurrencyCode: "EUR", Rate: -1},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestExchangeRateToResponse(t *testing.T) {
	recordedAt := time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)
	rate := domain.CurrencyExchangeRate{
		FromCurrency: domain.Currency{Code: "USD", Name: "US Dollar"},
		ToCurrency:   domain.Currency{Code: "EUR", Name: "Euro"},
		Rate:         domain.ExchangeRate(9_123),
		RecordedAt:   recordedAt,
	}

	resp := schema.ExchangeRateToResponse(rate)

	assert.Equal(t, "USD", resp.FromCurrencyCode)
	assert.Equal(t, "EUR", resp.ToCurrencyCode)
	assert.Equal(t, domain.ExchangeRate(9_123), resp.Rate)
	assert.True(t, recordedAt.Equal(resp.RecordedAt))
}
