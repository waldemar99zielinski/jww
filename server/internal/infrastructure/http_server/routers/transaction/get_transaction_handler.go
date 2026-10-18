package transaction

import (
	"net/http"

	"jww/internal/adapter/schema"
	"jww/internal/infrastructure/http_server/kernel"
	httputils "jww/internal/infrastructure/http_server/utils"
)

func getTransactionHandler(deps *kernel.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		appCtx := kernel.RequestContext(r)

		transactionID, err := httputils.ParseInt64PathParam(r, "transactionID")
		if err != nil {
			kernel.WriteBadRequest(w, err)
			return
		}

		result, err := deps.TransactionWorkflow.GetWithPreviewGrants(appCtx, transactionID)
		if err != nil {
			kernel.WriteError(w, r, err)
			return
		}
		w.Header().Set("Cache-Control", "private, no-store")
		kernel.WriteJSON(w, http.StatusOK, schema.TransactionToDetailsResponse(result.Transaction, result.PreviewGrants))
	}
}
