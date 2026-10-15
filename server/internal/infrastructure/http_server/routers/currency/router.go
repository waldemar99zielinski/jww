package currency

import (
	"github.com/go-chi/chi/v5"

	"jww/internal/infrastructure/http_server/kernel"
)

func Mount(_, protected chi.Router, deps *kernel.Deps) {
	protected.Post("/v1/currencies", createCurrencyHandler(deps))
	protected.Get("/v1/currencies", listCurrenciesHandler(deps))
	protected.Get("/v1/currencies/{code}", getCurrencyHandler(deps))
	protected.Put("/v1/currencies/{code}", updateCurrencyHandler(deps))
	protected.Delete("/v1/currencies/{code}", deleteCurrencyHandler(deps))
}
