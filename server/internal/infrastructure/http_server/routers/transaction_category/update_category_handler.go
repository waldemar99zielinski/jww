package transaction_category

import (
	"net/http"

	"jww/internal/adapter/schema"
	"jww/internal/infrastructure/http_server/kernel"
	httputils "jww/internal/infrastructure/http_server/utils"
)

func updateCategoryHandler(deps *kernel.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		appCtx := kernel.RequestContext(r)

		categoryID, err := httputils.ParseUUIDPathParam(r, "categoryID")
		if err != nil {
			kernel.WriteBadRequest(w, err)
			return
		}
		req, err := kernel.DecodeJSON[schema.TransactionCategoryUpdateRequest](r)
		if err != nil {
			kernel.WriteBadRequest(w, err)
			return
		}
		if err := req.Validate(); err != nil {
			kernel.WriteBadRequest(w, err)
			return
		}

		existing, err := deps.TransactionCategoryService.GetByID(appCtx, categoryID)
		if err != nil {
			kernel.WriteError(w, r, err)
			return
		}
		updated := *existing
		updated.Name = req.Name
		updated.Active = req.Active
		updated.SortOrder = req.SortOrder
		if _, err := deps.TransactionCategoryService.Update(appCtx, updated); err != nil {
			kernel.WriteError(w, r, err)
			return
		}
		kernel.WriteJSON(w, http.StatusOK, schema.TransactionCategoryToResponse(updated))
	}
}
