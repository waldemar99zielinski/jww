// Package frankfurter implements currency_exchange_rate.RateSource against the
// Frankfurter API (https://frankfurter.dev). It is the only place that speaks
// HTTP to an external rate provider. Adapters don't log — it returns the
// readable pairs and a RateSkip for every pair the provider omitted, and lets
// the boundary that called it log the outcome.
package frankfurter

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"jww/internal/domain"
	currency_exchange_rate "jww/internal/service/currency_exchange_rate"
)

// maxResponseBytes bounds how much of a response body is read, so a misbehaving
// server can't exhaust memory.
const maxResponseBytes = 1 << 20

// Client fetches rates from a Frankfurter-compatible API.
type Client struct {
	baseURL string
	http    *http.Client
}

func NewClient(cfg Config) *Client {
	return &Client{
		baseURL: strings.TrimRight(cfg.BaseURL, "/"),
		http:    &http.Client{Timeout: time.Duration(cfg.TimeoutSeconds) * time.Second},
	}
}

// Rates fetches every directed pair among currencies for asOf (zero = the
// provider's latest day). It makes one request per currency as base, with the
// rest as quotes, and returns the readable pairs plus a RateSkip for every
// requested quote the provider left out — nothing is dropped silently. A
// transport/decoding/non-200 failure aborts the whole fetch.
func (c *Client) Rates(ctx domain.AppContext, currencies []string, asOf time.Time) (currency_exchange_rate.RateFetch, error) {
	fetch := currency_exchange_rate.RateFetch{
		Rates: make([]domain.CurrencyExchangeRateFlatten, 0, len(currencies)*(len(currencies)-1)),
	}
	if len(currencies) < 2 {
		return fetch, nil
	}
	for _, base := range currencies {
		quotes := make([]string, 0, len(currencies)-1)
		for _, code := range currencies {
			if code != base {
				quotes = append(quotes, code)
			}
		}
		batch, err := c.rates(ctx, base, quotes, asOf)
		if err != nil {
			return currency_exchange_rate.RateFetch{}, err
		}
		fetch.Rates = append(fetch.Rates, batch.Rates...)
		fetch.Skipped = append(fetch.Skipped, batch.Skipped...)
	}
	return fetch, nil
}

// rates performs one GET /v2/rates?base=…&quotes=…[&date=…] and maps the
// response via mapRates.
func (c *Client) rates(ctx domain.AppContext, base string, quotes []string, asOf time.Time) (currency_exchange_rate.RateFetch, error) {
	query := url.Values{}
	query.Set("base", base)
	query.Set("quotes", strings.Join(quotes, ","))
	if !asOf.IsZero() {
		query.Set("date", asOf.Format("2006-01-02"))
	}
	endpoint := c.baseURL + "/v2/rates?" + query.Encode()

	req, err := http.NewRequestWithContext(ctx.Context(), http.MethodGet, endpoint, nil)
	if err != nil {
		return currency_exchange_rate.RateFetch{}, fmt.Errorf("frankfurter: building request for base %s: %w", base, err)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return currency_exchange_rate.RateFetch{}, fmt.Errorf("frankfurter: requesting rates for base %s: %w", base, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return currency_exchange_rate.RateFetch{}, fmt.Errorf("frankfurter: rates for base %s: unexpected status %d", base, resp.StatusCode)
	}

	var body []rateResponse
	if err := json.NewDecoder(io.LimitReader(resp.Body, maxResponseBytes)).Decode(&body); err != nil {
		return currency_exchange_rate.RateFetch{}, fmt.Errorf("frankfurter: decoding rates for base %s: %w", base, err)
	}
	return mapRates(base, quotes, body)
}
