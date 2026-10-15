package transaction

import (
	"net/http"

	"jww/internal/adapter/schema"
	"jww/internal/infrastructure/http_server/kernel"
)

func createTransactionHandler(deps *kernel.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		appCtx := kernel.RequestContext(r)

		req, err := kernel.DecodeJSON[schema.TransactionCreateRequest](r)
		if err != nil {
			kernel.WriteBadRequest(w, err)
			return
		}
		if err := req.Validate(); err != nil {
			kernel.WriteBadRequest(w, err)
			return
		}

		t, err := deps.TransactionWorkflow.Create(appCtx, req.Type, req.Amount, req.CurrencyCode, req.CategoryID, req.UserID, req.Source, req.Name, req.Description)
		if err != nil {
			kernel.WriteError(w, err)
			return
		}
		kernel.WriteJSON(w, http.StatusCreated, schema.TransactionToResponse(*t))
	}
}
