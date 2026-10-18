package currency

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"jww/internal/adapter/schema"
	"jww/internal/infrastructure/http_server/kernel"
)

func getCurrencyHandler(deps *kernel.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		code := chi.URLParam(r, "code")
		appCtx := kernel.RequestContext(r)

		c, err := deps.CurrencyService.GetByCode(appCtx, code)
		if err != nil {
			kernel.WriteError(w, r, err)
			return
		}
		kernel.WriteJSON(w, http.StatusOK, schema.CurrencyToResponse(*c))
	}
}
