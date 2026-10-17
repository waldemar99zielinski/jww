package transaction

import (
	"net/http"
	"time"

	"github.com/google/uuid"

	"jww/internal/adapter/schema"
	"jww/internal/domain"
	"jww/internal/infrastructure/http_server/kernel"
	transactionservice "jww/internal/service/transaction"
)

func listTransactionsHandler(deps *kernel.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		appCtx := kernel.RequestContext(r)

		filter, err := transactionFilterFromRequest(r)
		if err != nil {
			kernel.WriteBadRequest(w, err)
			return
		}
		transactions, err := deps.TransactionWorkflow.List(appCtx, filter)
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

func transactionFilterFromRequest(r *http.Request) (transactionservice.Filter, error) {
	q := r.URL.Query()
	filter := transactionservice.Filter{CurrencyCode: q.Get("currency_code")}

	if v := q.Get("user_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			return transactionservice.Filter{}, err
		}
		filter.UserID = &id
	}
	if v := q.Get("category_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			return transactionservice.Filter{}, err
		}
		filter.CategoryID = &id
	}
	if v := q.Get("type"); v != "" {
		transactionType := domain.TransactionType(v)
		if !transactionType.Valid() {
			return transactionservice.Filter{}, kernel.ErrInvalidType
		}
		filter.Type = &transactionType
	}
	if v := q.Get("source"); v != "" {
		source := domain.TransactionSource(v)
		if !source.Valid() {
			return transactionservice.Filter{}, kernel.ErrInvalidSource
		}
		filter.Source = &source
	}
	if v := q.Get("created_from"); v != "" {
		createdFrom, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return transactionservice.Filter{}, err
		}
		filter.CreatedFrom = &createdFrom
	}
	if v := q.Get("created_to"); v != "" {
		createdTo, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return transactionservice.Filter{}, err
		}
		filter.CreatedTo = &createdTo
	}

	return filter, nil
}
