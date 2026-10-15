package transaction_category

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"jww/internal/infrastructure/http_server/kernel"
)

func deleteCategoryHandler(deps *kernel.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		appCtx := kernel.RequestContext(r)

		categoryID, err := uuid.Parse(chi.URLParam(r, "categoryID"))
		if err != nil {
			kernel.WriteBadRequest(w, err)
			return
		}

		c, err := deps.TransactionCategoryService.GetByID(appCtx, categoryID)
		if err != nil {
			kernel.WriteError(w, err)
			return
		}
		if err := deps.TransactionCategoryService.Delete(appCtx, *c); err != nil {
			kernel.WriteError(w, err)
			return
		}
		kernel.WriteNoContent(w)
	}
}
