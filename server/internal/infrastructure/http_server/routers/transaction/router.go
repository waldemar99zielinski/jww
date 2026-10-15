package transaction

import (
	"github.com/go-chi/chi/v5"

	"jww/internal/infrastructure/http_server/kernel"
)

func Mount(_, protected chi.Router, deps *kernel.Deps) {
	protected.Post("/v1/transactions", createTransactionHandler(deps))
	protected.Get("/v1/transactions", listTransactionsHandler(deps))
	protected.Get("/v1/transactions/{transactionID}", getTransactionHandler(deps))
	protected.Delete("/v1/transactions/{transactionID}", deleteTransactionHandler(deps))
}
