package currency_exchange_rate

import (
	"net/http"

	"jww/internal/adapter/schema"
	"jww/internal/infrastructure/http_server/kernel"
)

func createExchangeRateHandler(deps *kernel.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		appCtx := kernel.RequestContext(r)

		req, err := kernel.DecodeJSON[schema.CurrencyExchangeRateCreateRequest](r)
		if err != nil {
			kernel.WriteBadRequest(w, err)
			return
		}
		if err := req.Validate(); err != nil {
			kernel.WriteBadRequest(w, err)
			return
		}

		rate, err := deps.CurrencyExchangeRateWorkflow.Create(appCtx, req.FromCurrencyCode, req.ToCurrencyCode, req.Rate)
		if err != nil {
			kernel.WriteError(w, err)
			return
		}
		kernel.WriteJSON(w, http.StatusCreated, schema.ExchangeRateToResponse(*rate))
	}
}
