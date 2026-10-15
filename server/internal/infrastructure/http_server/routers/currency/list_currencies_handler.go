package currency

import (
	"net/http"

	"jww/internal/adapter/schema"
	"jww/internal/infrastructure/http_server/kernel"
)

func listCurrenciesHandler(deps *kernel.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		appCtx := kernel.RequestContext(r)

		currencies, err := deps.CurrencyService.List(appCtx)
		if err != nil {
			kernel.WriteError(w, err)
			return
		}
		out := make([]schema.CurrencyResponse, len(currencies))
		for i, c := range currencies {
			out[i] = schema.CurrencyToResponse(c)
		}
		kernel.WriteJSON(w, http.StatusOK, out)
	}
}
