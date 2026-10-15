package transaction

import (
	"net/http"
	"time"

	"github.com/google/uuid"

	"jww/internal/adapter/schema"
	"jww/internal/domain"
	"jww/internal/infrastructure/http_server/kernel"
	"jww/internal/service/workflow"
)

func listTransactionsHandler(deps *kernel.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		appCtx := kernel.RequestContext(r)

		q := r.URL.Query()

		var userID, categoryID *uuid.UUID
		if v := q.Get("user_id"); v != "" {
			id, err := uuid.Parse(v)
			if err != nil {
				kernel.WriteBadRequest(w, err)
				return
			}
			userID = &id
		}
		if v := q.Get("category_id"); v != "" {
			id, err := uuid.Parse(v)
			if err != nil {
				kernel.WriteBadRequest(w, err)
				return
			}
			categoryID = &id
		}

		var transactionType *domain.TransactionType
		if v := q.Get("type"); v != "" {
			t := domain.TransactionType(v)
			if !t.Valid() {
				kernel.WriteBadRequest(w, kernel.ErrInvalidType)
				return
			}
			transactionType = &t
		}

		var source *domain.TransactionSource
		if v := q.Get("source"); v != "" {
			s := domain.TransactionSource(v)
			if !s.Valid() {
				kernel.WriteBadRequest(w, kernel.ErrInvalidSource)
				return
			}
			source = &s
		}

		currencyCode := q.Get("currency_code")

		var createdFrom, createdTo *time.Time
		if v := q.Get("created_from"); v != "" {
			t, err := time.Parse(time.RFC3339, v)
			if err != nil {
				kernel.WriteBadRequest(w, err)
				return
			}
			createdFrom = &t
		}
		if v := q.Get("created_to"); v != "" {
			t, err := time.Parse(time.RFC3339, v)
			if err != nil {
				kernel.WriteBadRequest(w, err)
				return
			}
			createdTo = &t
		}

		transactions, err := deps.TransactionWorkflow.List(appCtx, workflow.TransactionListFilter{
			UserID:       userID,
			CategoryID:   categoryID,
			Type:         transactionType,
			CurrencyCode: currencyCode,
			Source:       source,
			CreatedFrom:  createdFrom,
			CreatedTo:    createdTo,
		})
		if err != nil {
			kernel.WriteError(w, err)
			return
		}
		out := make([]schema.TransactionResponse, len(transactions))
		for i, t := range transactions {
			out[i] = schema.TransactionToResponse(t)
		}
		kernel.WriteJSON(w, http.StatusOK, out)
	}
}
