//go:build e2e

package e2e_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"jww/internal/adapter/schema"
	"jww/internal/domain"
	"jww/tests/utils"
)

// TestCRUDJourney exercises currency -> category -> transaction creation
// (each referencing the previous), fetches the transaction back, lists with
// a filter that should match it, deletes it, and confirms the delete stuck.
// Uses currency code ZZC, not used by any other journey or by the migration
// seed data (seeded categories only, no currencies).
func TestCRUDJourney(t *testing.T) {
	token := apiClient.Login(t, seededUser.Username, seededPassword)

	var currencyResp utils.Envelope[schema.CurrencyResponse]
	status := apiClient.Do(t, http.MethodPost, "/api/v1/currencies", token, schema.CurrencyCreateRequest{
		Code:   "ZZC",
		Name:   "E2E CRUD Currency",
		IsBase: false,
	}, &currencyResp)
	require.Equal(t, http.StatusCreated, status)
	require.Equal(t, "ZZC", currencyResp.Data.Code)

	var categoryResp utils.Envelope[schema.TransactionCategoryResponse]
	status = apiClient.Do(t, http.MethodPost, "/api/v1/transaction-categories", token, schema.TransactionCategoryCreateRequest{
		Type:      domain.TransactionCategoryTypeExpense,
		Name:      "E2E CRUD Category",
		SortOrder: 0,
	}, &categoryResp)
	require.Equal(t, http.StatusCreated, status)
	categoryID := categoryResp.Data.ID

	var txResp utils.Envelope[schema.TransactionResponse]
	status = apiClient.Do(t, http.MethodPost, "/api/v1/transactions", token, schema.TransactionCreateRequest{
		Type:         domain.TransactionTypeExpense,
		Amount:       domain.Money(1234), // $12.34
		CurrencyCode: "ZZC",
		CategoryID:   categoryID,
		UserID:       seededUser.ID,
		Source:       domain.TransactionSourceManual,
		Name:         "E2E CRUD Transaction",
		Description:  "created by the CRUD journey test",
	}, &txResp)
	require.Equal(t, http.StatusCreated, status)
	txID := txResp.Data.ID
	require.NotZero(t, txID)
	assert.Equal(t, "ZZC", txResp.Data.CurrencyCode)
	assert.Equal(t, categoryID, txResp.Data.CategoryID)
	assert.Equal(t, domain.Money(1234), txResp.Data.Amount)

	t.Run("fetch it back", func(t *testing.T) {
		var got utils.Envelope[schema.TransactionResponse]
		status := apiClient.Do(t, http.MethodGet, fmt.Sprintf("/api/v1/transactions/%d", txID), token, nil, &got)
		require.Equal(t, http.StatusOK, status)
		assert.Equal(t, txID, got.Data.ID)
	})

	t.Run("list filtered by currency matches it", func(t *testing.T) {
		var list utils.Envelope[[]schema.TransactionResponse]
		status := apiClient.Do(t, http.MethodGet, "/api/v1/transactions?currency_code=ZZC", token, nil, &list)
		require.Equal(t, http.StatusOK, status)
		found := false
		for _, tx := range list.Data {
			if tx.ID == txID {
				found = true
				break
			}
		}
		assert.True(t, found, "expected transaction %d in currency_code=ZZC list", txID)
	})

	t.Run("delete then re-fetch is 404", func(t *testing.T) {
		status := apiClient.Do(t, http.MethodDelete, fmt.Sprintf("/api/v1/transactions/%d", txID), token, nil, nil)
		require.Equal(t, http.StatusNoContent, status)

		var errResp utils.Envelope[struct{}]
		status = apiClient.Do(t, http.MethodGet, fmt.Sprintf("/api/v1/transactions/%d", txID), token, nil, &errResp)
		assert.Equal(t, http.StatusNotFound, status)
		require.NotNil(t, errResp.Error)
		assert.Equal(t, "not_found", errResp.Error.Code)
	})
}
