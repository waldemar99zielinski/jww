package transaction_category

import (
	"github.com/go-chi/chi/v5"

	"jww/internal/infrastructure/http_server/kernel"
)

func Mount(_, protected chi.Router, deps *kernel.Deps) {
	protected.Post("/v1/transaction-categories", createCategoryHandler(deps))
	protected.Get("/v1/transaction-categories", listCategoriesHandler(deps))
	protected.Get("/v1/transaction-categories/{categoryID}", getCategoryHandler(deps))
	protected.Put("/v1/transaction-categories/{categoryID}", updateCategoryHandler(deps))
	protected.Delete("/v1/transaction-categories/{categoryID}", deleteCategoryHandler(deps))
}
