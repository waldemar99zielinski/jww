package transaction

import (
	"net/http"

	"jww/internal/adapter/schema"
	"jww/internal/domain"
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

		subTransactions := make([]domain.NewSubTransaction, len(req.SubTransactions))
		for i, item := range req.SubTransactions {
			subTransactions[i] = domain.NewSubTransaction{Name: item.Name, Label: item.Label, Amount: item.Amount}
		}

		t, err := deps.TransactionWorkflow.Create(appCtx, req.Type, req.Amount, req.CurrencyCode, req.CategoryID, req.UserID, req.Source, req.Name, req.Description, req.CreatedAt, req.AttachmentIDs, subTransactions)
		if err != nil {
			kernel.WriteError(w, r, err)
			return
		}
		kernel.WriteJSON(w, http.StatusCreated, schema.TransactionToResponse(*t))
	}
}
