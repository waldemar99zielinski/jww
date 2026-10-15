package schema

import (
	"fmt"
	"time"

	"jww/internal/domain"
)

type CurrencyExchangeRateCreateRequest struct {
	FromCurrencyCode string              `json:"from_currency_code"`
	ToCurrencyCode   string              `json:"to_currency_code"`
	Rate             domain.ExchangeRate `json:"rate"`
}

func (r CurrencyExchangeRateCreateRequest) Validate() error {
	if err := firstError(
		exactLength("from_currency_code", r.FromCurrencyCode, 3),
		exactLength("to_currency_code", r.ToCurrencyCode, 3),
	); err != nil {
		return err
	}
	if r.Rate <= 0 {
		return fmt.Errorf("rate must be greater than 0")
	}
	return nil
}

type CurrencyExchangeRateResponse struct {
	FromCurrencyCode string              `json:"from_currency_code"`
	ToCurrencyCode   string              `json:"to_currency_code"`
	Rate             domain.ExchangeRate `json:"rate"`
	RecordedAt       time.Time           `json:"recorded_at"`
}

func ExchangeRateToResponse(r domain.CurrencyExchangeRate) CurrencyExchangeRateResponse {
	return CurrencyExchangeRateResponse{
		FromCurrencyCode: r.FromCurrency.Code,
		ToCurrencyCode:   r.ToCurrency.Code,
		Rate:             r.Rate,
		RecordedAt:       r.RecordedAt,
	}
}
