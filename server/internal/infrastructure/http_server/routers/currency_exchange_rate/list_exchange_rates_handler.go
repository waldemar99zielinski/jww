package currency_exchange_rate

import (
	"net/http"

	"jww/internal/adapter/schema"
	"jww/internal/infrastructure/http_server/kernel"
)

func listExchangeRatesHandler(deps *kernel.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		appCtx := kernel.RequestContext(r)

		rates, err := deps.CurrencyExchangeRateWorkflow.List(appCtx)
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
