package transaction

import (
	"net/http"

	"jww/internal/infrastructure/http_server/kernel"
	httputils "jww/internal/infrastructure/http_server/utils"
)

func deleteTransactionHandler(deps *kernel.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		appCtx := kernel.RequestContext(r)

		transactionID, err := httputils.ParseInt64PathParam(r, "transactionID")
		if err != nil {
			kernel.WriteBadRequest(w, err)
			return
		}
		if err := deps.TransactionWorkflow.Delete(appCtx, transactionID); err != nil {
			kernel.WriteError(w, err)
			return
		}
		kernel.WriteNoContent(w)
	}
}
