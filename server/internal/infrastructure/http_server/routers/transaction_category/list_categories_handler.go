package transaction_category

import (
	"net/http"
	"strconv"

	"jww/internal/adapter/schema"
	"jww/internal/domain"
	"jww/internal/infrastructure/http_server/kernel"
)

func listCategoriesHandler(deps *kernel.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		appCtx := kernel.RequestContext(r)

		categoryType, activeOnly, err := categoryListParamsFromRequest(r)
		if err != nil {
			kernel.WriteBadRequest(w, err)
			return
		}

		categories, err := deps.TransactionCategoryService.ListByType(appCtx, categoryType, activeOnly)
		if err != nil {
			kernel.WriteError(w, err)
			return
		}
		out := make([]schema.TransactionCategoryResponse, len(categories))
		for i, c := range categories {
			out[i] = schema.TransactionCategoryToResponse(c)
		}
		kernel.WriteJSON(w, http.StatusOK, out)
	}
}

func categoryListParamsFromRequest(r *http.Request) (domain.TransactionCategoryType, bool, error) {
	q := r.URL.Query()
	categoryType := domain.TransactionCategoryType(q.Get("type"))
	if !categoryType.Valid() {
		return "", false, kernel.ErrInvalidType
	}

	activeOnly := true
	if v := q.Get("active_only"); v != "" {
		parsed, err := strconv.ParseBool(v)
		if err != nil {
			return "", false, err
		}
		activeOnly = parsed
	}
	return categoryType, activeOnly, nil
}
