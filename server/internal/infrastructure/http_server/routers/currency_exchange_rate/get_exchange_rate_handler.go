package currency_exchange_rate

import (
	"errors"
	"net/http"

	"jww/internal/adapter/schema"
	"jww/internal/infrastructure/http_server/kernel"
)

var errMissingCurrencyCodes = errors.New("from_currency_code and to_currency_code are required")

func getExchangeRateHandler(deps *kernel.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		appCtx := kernel.RequestContext(r)

		q := r.URL.Query()
		fromCode := q.Get("from_currency_code")
		toCode := q.Get("to_currency_code")
		if fromCode == "" || toCode == "" {
			kernel.WriteBadRequest(w, errMissingCurrencyCodes)
			return
		}
		asOf, err := schema.ParseDate(q.Get("as_of"))
		if err != nil {
			kernel.WriteBadRequest(w, err)
			return
		}

		rate, err := deps.CurrencyExchangeRateWorkflow.GetByCurrencies(appCtx, fromCode, toCode, asOf.Time())
		if err != nil {
			kernel.WriteError(w, r, err)
			return
		}
		kernel.WriteJSON(w, http.StatusOK, schema.ExchangeRateToResponse(*rate))
	}
}
