package currency_exchange_rate

import (
	"net/http"
	"time"

	"jww/internal/adapter/schema"
	"jww/internal/infrastructure/http_server/kernel"
	currencyexchangerate "jww/internal/service/currency_exchange_rate"
)

func listExchangeRatesHandler(deps *kernel.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		appCtx := kernel.RequestContext(r)

		filter, err := exchangeRateFilterFromRequest(r)
		if err != nil {
			kernel.WriteBadRequest(w, err)
			return
		}

		rates, err := deps.CurrencyExchangeRateWorkflow.List(appCtx, filter)
		if err != nil {
			kernel.WriteError(w, err)
			return
		}
		out := make([]schema.CurrencyExchangeRateResponse, len(rates))
		for i, rt := range rates {
			out[i] = schema.ExchangeRateToResponse(rt)
		}
		kernel.WriteJSON(w, http.StatusOK, out)
	}
}

func exchangeRateFilterFromRequest(r *http.Request) (currencyexchangerate.Filter, error) {
	q := r.URL.Query()
	filter := currencyexchangerate.Filter{
		FromCurrencyCode: q.Get("from_currency_code"),
		ToCurrencyCode:   q.Get("to_currency_code"),
	}
	if v := q.Get("recorded_from"); v != "" {
		recordedFrom, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return currencyexchangerate.Filter{}, err
		}
		filter.RecordedFrom = &recordedFrom
	}
	if v := q.Get("recorded_to"); v != "" {
		recordedTo, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return currencyexchangerate.Filter{}, err
		}
		filter.RecordedTo = &recordedTo
	}
	return filter, nil
}
