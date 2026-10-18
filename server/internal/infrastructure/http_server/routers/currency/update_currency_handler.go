package currency

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"jww/internal/adapter/schema"
	"jww/internal/infrastructure/http_server/kernel"
)

func updateCurrencyHandler(deps *kernel.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		code := chi.URLParam(r, "code")
		appCtx := kernel.RequestContext(r)

		req, err := kernel.DecodeJSON[schema.CurrencyUpdateRequest](r)
		if err != nil {
			kernel.WriteBadRequest(w, err)
			return
		}
		if err := req.Validate(); err != nil {
			kernel.WriteBadRequest(w, err)
			return
		}

		existing, err := deps.CurrencyService.GetByCode(appCtx, code)
		if err != nil {
			kernel.WriteError(w, r, err)
			return
		}
		updated := *existing
		updated.Name = req.Name
		updated.IsBase = req.IsBase
		if _, err := deps.CurrencyService.Update(appCtx, updated); err != nil {
			kernel.WriteError(w, r, err)
			return
		}
		kernel.WriteJSON(w, http.StatusOK, schema.CurrencyToResponse(updated))
	}
}
