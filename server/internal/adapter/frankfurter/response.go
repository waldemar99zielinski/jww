package frankfurter

import (
	"encoding/json"
	"fmt"
	"time"

	"jww/internal/domain"
	currency_exchange_rate "jww/internal/service/currency_exchange_rate"
)

// rateResponse is one row of the /v2/rates array. Rate is decoded as a
// json.Number (never float64) so its exact text reaches
// domain.ParseExchangeRate.
type rateResponse struct {
	Date      string      `json:"date"`
	Base      string      `json:"base"`
	QuoteCode string      `json:"quote"`
	Rate      json.Number `json:"rate"`
}

// mapRates converts the decoded rows for one base into the readable directed
// pairs — each anchored to the echoed business date at midnight UTC — plus a
// RateSkip for every requested quote the provider omitted. The first
// unparsable row aborts.
func mapRates(base string, quotes []string, body []rateResponse) (currency_exchange_rate.RateFetch, error) {
	fetch := currency_exchange_rate.RateFetch{Rates: make([]domain.CurrencyExchangeRateFlatten, 0, len(body))}
	seen := make(map[string]struct{}, len(body))
	for _, row := range body {
		rate, err := domain.ParseExchangeRate(row.Rate.String())
		if err != nil {
			return currency_exchange_rate.RateFetch{}, fmt.Errorf("frankfurter: parsing %s→%s rate %q: %w", row.Base, row.QuoteCode, row.Rate.String(), err)
		}
		date, err := time.Parse("2006-01-02", row.Date)
		if err != nil {
			return currency_exchange_rate.RateFetch{}, fmt.Errorf("frankfurter: parsing date %q: %w", row.Date, err)
		}
		fetch.Rates = append(fetch.Rates, domain.CurrencyExchangeRateFlatten{
			FromCurrencyCode: row.Base,
			ToCurrencyCode:   row.QuoteCode,
			Rate:             rate,
			RecordedAt:       date.UTC(),
		})
		seen[row.QuoteCode] = struct{}{}
	}
	for _, quote := range quotes {
		if _, ok := seen[quote]; !ok {
			fetch.Skipped = append(fetch.Skipped, currency_exchange_rate.RateSkip{
				From: base, To: quote, Reason: "provider returned no rate",
			})
		}
	}
	return fetch, nil
}
