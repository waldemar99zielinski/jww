package currency_exchange_rate

import (
	"github.com/go-chi/chi/v5"

	"jww/internal/infrastructure/http_server/kernel"
)

func Mount(_, protected chi.Router, deps *kernel.Deps) {
	protected.Post("/v1/currency-exchange-rates", createExchangeRateHandler(deps))
	protected.Get("/v1/currency-exchange-rates", getExchangeRateHandler(deps))
	protected.Get("/v1/currency-exchange-rates/all", listExchangeRatesHandler(deps))
}
