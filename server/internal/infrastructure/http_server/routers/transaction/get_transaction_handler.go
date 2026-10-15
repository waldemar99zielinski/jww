package transaction

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"jww/internal/adapter/schema"
	"jww/internal/infrastructure/http_server/kernel"
)

func getTransactionHandler(deps *kernel.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		appCtx := kernel.RequestContext(r)

		transactionID, err := strconv.ParseInt(chi.URLParam(r, "transactionID"), 10, 64)
		if err != nil {
			kernel.WriteBadRequest(w, err)
			return
		}

		t, err := deps.TransactionService.GetByID(appCtx, transactionID)
		if err != nil {
			kernel.WriteError(w, err)
			return
		}
		kernel.WriteJSON(w, http.StatusOK, schema.TransactionToResponse(*t))
	}
}
