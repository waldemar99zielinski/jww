//go:build e2e

package e2e_test

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"jww/internal/adapter/schema"
	"jww/internal/domain"
	"jww/tests/utils"
)

// TestSummaryJourney is the riskiest read path in the system: it touches
// currency, exchange rate, transaction, and user services together. Two
// currencies (ZZA is the summary's target currency, ZZB only exists to be
// converted from) with a ZZB->ZZA exchange rate of 2.0, one expense
// transaction in each currency, and asserts the summary's totals reflect
// the conversion, not just a sum of raw amounts.
func TestSummaryJourney(t *testing.T) {
	token := apiClient.Login(t, seededUser.Username, seededPassword)

	for _, c := range []schema.CurrencyCreateRequest{
		{Code: "ZZA", Name: "E2E Summary Target Currency", IsBase: false},
		{Code: "ZZB", Name: "E2E Summary Source Currency", IsBase: false},
	} {
		var resp utils.Envelope[schema.CurrencyResponse]
		status := apiClient.Do(t, http.MethodPost, "/api/v1/currencies", token, c, &resp)
		require.Equal(t, http.StatusCreated, status, "creating currency %s", c.Code)
	}

	var rateResp utils.Envelope[schema.CurrencyExchangeRateResponse]
	status := apiClient.Do(t, http.MethodPost, "/api/v1/currency-exchange-rates", token, schema.CurrencyExchangeRateCreateRequest{
		FromCurrencyCode: "ZZB",
		ToCurrencyCode:   "ZZA",
		Rate:             domain.ExchangeRate(20000), // 2.0000: 1 ZZB == 2 ZZA
	}, &rateResp)
	require.Equal(t, http.StatusCreated, status)

	var categoryResp utils.Envelope[schema.TransactionCategoryResponse]
	status = apiClient.Do(t, http.MethodPost, "/api/v1/transaction-categories", token, schema.TransactionCategoryCreateRequest{
		Type: domain.TransactionCategoryTypeExpense,
		Name: "E2E Summary Category",
	}, &categoryResp)
	require.Equal(t, http.StatusCreated, status)
	categoryID := categoryResp.Data.ID

	createExpense := func(currencyCode string, amount domain.Money) {
		var txResp utils.Envelope[schema.TransactionResponse]
		status := apiClient.Do(t, http.MethodPost, "/api/v1/transactions", token, schema.TransactionCreateRequest{
			Type:         domain.TransactionTypeExpense,
			Amount:       amount,
			CurrencyCode: currencyCode,
			CategoryID:   categoryID,
			UserID:       seededUser.ID,
			Source:       domain.TransactionSourceManual,
			Name:         fmt.Sprintf("E2E Summary Transaction (%s)", currencyCode),
		}, &txResp)
		require.Equal(t, http.StatusCreated, status, "creating transaction in %s", currencyCode)
	}

	createExpense("ZZA", domain.Money(500))  // $5.00 native
	createExpense("ZZB", domain.Money(1000)) // $10.00, converts to $20.00 at rate 2.0

	periodEnd := time.Now().UTC().Format("2006-01-02")
	path := fmt.Sprintf("/api/v1/summary?currency_code=ZZA&period_end=%s&user_id=%s", periodEnd, seededUser.ID)

	var summaryResp utils.Envelope[schema.SummaryResponse]
	status = apiClient.Do(t, http.MethodGet, path, token, nil, &summaryResp)
	require.Equal(t, http.StatusOK, status)

	// 500 (native ZZA) + 1000*2.0 (ZZB converted to ZZA) == 2500 == $25.00.
	assert.Equal(t, domain.Money(2500), summaryResp.Data.TotalExpense,
		"expected native ZZA expense plus the converted ZZB expense")
	assert.Equal(t, domain.Money(0), summaryResp.Data.TotalIncome)
	// Net = income - expense (domain.Summary.Validate() enforces this
	// invariant) — an all-expense scenario nets negative.
	assert.Equal(t, domain.Money(-2500), summaryResp.Data.Net)
	assert.Equal(t, "ZZA", summaryResp.Data.CurrencyCode)

	byCurrency := summaryResp.Data.ByCurrency
	require.Contains(t, byCurrency, "ZZA")
	require.Contains(t, byCurrency, "ZZB")
	assert.Equal(t, domain.Money(500), byCurrency["ZZA"].TotalExpense, "ZZA's own bucket stays unconverted")
	assert.Equal(t, domain.Money(1000), byCurrency["ZZB"].TotalExpense, "ZZB's own bucket stays unconverted")
}
