package transaction

import (
	"net/http"

	"jww/internal/adapter/schema"
	"jww/internal/domain"
	"jww/internal/infrastructure/http_server/kernel"
	httputils "jww/internal/infrastructure/http_server/utils"
)

// createSubTransactionHandler adds one line item to the path transaction and
// returns the updated transaction (amount recomputed as the sum of its items).
func createSubTransactionHandler(deps *kernel.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		appCtx := kernel.RequestContext(r)

		transactionID, err := httputils.ParseInt64PathParam(r, "transactionID")
		if err != nil {
			kernel.WriteBadRequest(w, err)
			return
		}
		req, err := kernel.DecodeJSON[schema.TransactionSubTransactionCreateRequest](r)
		if err != nil {
			kernel.WriteBadRequest(w, err)
			return
		}
		if err := req.Validate(); err != nil {
			kernel.WriteBadRequest(w, err)
			return
		}

		updated, err := deps.TransactionWorkflow.CreateSubTransaction(appCtx, transactionID, domain.NewSubTransaction{
			Name:   req.Name,
			Label:  req.Label,
			Amount: req.Amount,
		})
		if err != nil {
			kernel.WriteError(w, r, err)
			return
		}
		kernel.WriteJSON(w, http.StatusCreated, schema.TransactionToResponse(*updated))
	}
}

// updateSubTransactionHandler patches one line item and returns the updated
// transaction (amount recomputed as the sum of its items).
func updateSubTransactionHandler(deps *kernel.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		appCtx := kernel.RequestContext(r)

		transactionID, err := httputils.ParseInt64PathParam(r, "transactionID")
		if err != nil {
			kernel.WriteBadRequest(w, err)
			return
		}
		subTransactionID, err := httputils.ParseUUIDPathParam(r, "subTransactionID")
		if err != nil {
			kernel.WriteBadRequest(w, err)
			return
		}
		req, err := kernel.DecodeJSON[schema.TransactionSubTransactionUpdateRequest](r)
		if err != nil {
			kernel.WriteBadRequest(w, err)
			return
		}
		if err := req.Validate(); err != nil {
			kernel.WriteBadRequest(w, err)
			return
		}

		updated, err := deps.TransactionWorkflow.UpdateSubTransaction(appCtx, transactionID, subTransactionID, domain.SubTransactionPatch{
			Name:   req.Name,
			Label:  req.Label,
			Amount: req.Amount,
		})
		if err != nil {
			kernel.WriteError(w, r, err)
			return
		}
		kernel.WriteJSON(w, http.StatusOK, schema.TransactionToResponse(*updated))
	}
}

// deleteSubTransactionHandler removes one line item and returns 204; the parent
// amount is recomputed (or left unlocked when the list becomes empty).
func deleteSubTransactionHandler(deps *kernel.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		appCtx := kernel.RequestContext(r)

		transactionID, err := httputils.ParseInt64PathParam(r, "transactionID")
		if err != nil {
			kernel.WriteBadRequest(w, err)
			return
		}
		subTransactionID, err := httputils.ParseUUIDPathParam(r, "subTransactionID")
		if err != nil {
			kernel.WriteBadRequest(w, err)
			return
		}
		if _, err := deps.TransactionWorkflow.DeleteSubTransaction(appCtx, transactionID, subTransactionID); err != nil {
			kernel.WriteError(w, r, err)
			return
		}
		kernel.WriteNoContent(w)
	}
}
