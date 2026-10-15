package transaction_category

import (
	"net/http"

	"jww/internal/adapter/schema"
	"jww/internal/infrastructure/http_server/kernel"
)

func createCategoryHandler(deps *kernel.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		appCtx := kernel.RequestContext(r)

		req, err := kernel.DecodeJSON[schema.TransactionCategoryCreateRequest](r)
		if err != nil {
			kernel.WriteBadRequest(w, err)
			return
		}
		if err := req.Validate(); err != nil {
			kernel.WriteBadRequest(w, err)
			return
		}

		c, err := deps.TransactionCategoryService.Create(appCtx, req.Type, req.Name, req.SortOrder)
		if err != nil {
			kernel.WriteError(w, err)
			return
		}
		kernel.WriteJSON(w, http.StatusCreated, schema.TransactionCategoryToResponse(*c))
	}
}
