package transaction

import (
	"net/http"

	"jww/internal/adapter/schema"
	"jww/internal/infrastructure/http_server/kernel"
	httputils "jww/internal/infrastructure/http_server/utils"
	transactionservice "jww/internal/service/transaction"
)

func updateTransactionHandler(deps *kernel.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		appCtx := kernel.RequestContext(r)

		transactionID, err := httputils.ParseInt64PathParam(r, "transactionID")
		if err != nil {
			kernel.WriteBadRequest(w, err)
			return
		}
		req, err := kernel.DecodeJSON[schema.TransactionUpdateRequest](r)
		if err != nil {
			kernel.WriteBadRequest(w, err)
			return
		}
		if err := req.Validate(); err != nil {
			kernel.WriteBadRequest(w, err)
			return
		}

		// The schema's optional fields mirror transaction.Patch one-to-one, so
		// only present fields (non-nil) are carried into the patch.
		result, err := deps.TransactionWorkflow.Update(appCtx, transactionID, transactionservice.Patch{
			Type:         req.Type,
			Amount:       req.Amount,
			CurrencyCode: req.CurrencyCode,
			CategoryID:   req.CategoryID,
			UserID:       req.UserID,
			Source:       req.Source,
			Name:         req.Name,
			Description:  req.Description,
			CreatedAt:    req.CreatedAt,
		})
		if err != nil {
			kernel.WriteError(w, r, err)
			return
		}
		w.Header().Set("Cache-Control", "private, no-store")
		kernel.WriteJSON(w, http.StatusOK, schema.TransactionToDetailsResponse(result.Transaction, result.PreviewGrants))
	}
}
