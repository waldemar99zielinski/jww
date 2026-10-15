package currency

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"jww/internal/infrastructure/http_server/kernel"
)

func deleteCurrencyHandler(deps *kernel.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		code := chi.URLParam(r, "code")
		appCtx := kernel.RequestContext(r)

		c, err := deps.CurrencyService.GetByCode(appCtx, code)
		if err != nil {
			kernel.WriteError(w, err)
			return
		}
		if err := deps.CurrencyService.Delete(appCtx, *c); err != nil {
			kernel.WriteError(w, err)
			return
		}
		kernel.WriteNoContent(w)
	}
}
