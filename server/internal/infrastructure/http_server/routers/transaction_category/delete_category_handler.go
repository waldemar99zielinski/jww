package transaction_category

import (
	"net/http"

	"jww/internal/infrastructure/http_server/kernel"
	httputils "jww/internal/infrastructure/http_server/utils"
)

func deleteCategoryHandler(deps *kernel.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		appCtx := kernel.RequestContext(r)

		categoryID, err := httputils.ParseUUIDPathParam(r, "categoryID")
		if err != nil {
			kernel.WriteBadRequest(w, err)
			return
		}

		c, err := deps.TransactionCategoryService.GetByID(appCtx, categoryID)
		if err != nil {
			kernel.WriteError(w, r, err)
			return
		}
		if err := deps.TransactionCategoryService.Delete(appCtx, *c); err != nil {
			kernel.WriteError(w, r, err)
			return
		}
		kernel.WriteNoContent(w)
	}
}
