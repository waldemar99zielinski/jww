//go:build e2e

package e2e_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"jww/internal/adapter/schema"
	"jww/internal/domain"
	"jww/tests/utils"
)

// TestSubTransactionJourney covers the analysis §14 "Server e2e" row end to
// end: upload an IMAGE draft → run extraction (the e2e stub reader returns a
// deterministic canned prefill) → create a transaction from the accepted items
// plus the attachment → read it back with the derived Σ(items) amount → then
// add/patch/delete line items on the existing transaction and assert the parent
// amount recomputes each time.
//
// The amount-accumulation rule under test (analysis §3/§7): whenever a
// transaction has one or more sub-transactions its amount is derived as their
// sum — the create request's own amount is ignored, a direct amount patch is
// rejected, and deleting the last remaining item leaves the amount at its last
// computed value (never zeroed) while unlocking it for a direct patch again.
//
// Data isolation: unique currency code ZZD and "E2E Sub-Transaction …" names;
// no exact list-count assertions (see this directory's README).
func TestSubTransactionJourney(t *testing.T) {
	token := apiClient.Login(t, seededUser.Username, seededPassword)

	var currencyResp utils.Envelope[schema.CurrencyResponse]
	status := apiClient.Do(t, http.MethodPost, "/api/v1/currencies", token, schema.CurrencyCreateRequest{
		Code: "ZZD", Name: "E2E Sub-Transaction Currency", IsBase: false,
	}, &currencyResp)
	require.Equal(t, http.StatusCreated, status)

	var categoryResp utils.Envelope[schema.TransactionCategoryResponse]
	status = apiClient.Do(t, http.MethodPost, "/api/v1/transaction-categories", token, schema.TransactionCategoryCreateRequest{
		Type: domain.TransactionCategoryTypeExpense, Name: "E2E Sub-Transaction Category", SortOrder: 2,
	}, &categoryResp)
	require.Equal(t, http.StatusCreated, status)

	// 1. Upload an IMAGE draft: extraction is images-only in v1, so any other
	// media type (PDF included) would be a 400 before the reader is reached.
	resp, raw := apiClient.UploadFile(t, "/api/v1/transaction-attachments", token, "receipt.jpg", "image/jpeg",
		[]byte("\xff\xd8\xff\xe0\x00\x10JFIF e2e canned extraction receipt"))
	require.Equal(t, http.StatusCreated, resp.StatusCode, "upload draft body: %s", raw)
	var draft utils.Envelope[schema.TransactionAttachmentResponse]
	require.NoError(t, json.Unmarshal(raw, &draft))
	draftID := draft.Data.ID
	require.NotEqual(t, uuid.Nil, draftID)
	require.Nil(t, draft.Data.TransactionID, "a fresh draft is unattached")

	// 2. Extract: the stub reader makes this deterministic and offline. The
	// response must carry the canned scalar suggestions and both line items.
	var prefill utils.Envelope[schema.TransactionPrefillResponse]
	status = apiClient.Do(t, http.MethodPost,
		"/api/v1/transaction-attachments/"+draftID.String()+"/extraction", token, nil, &prefill)
	require.Equal(t, http.StatusOK, status)
	assert.Equal(t, draftID, prefill.Data.AttachmentID)

	require.NotNil(t, prefill.Data.Type, "type suggestion is present")
	assert.Equal(t, domain.TransactionTypeExpense, prefill.Data.Type.Value)
	require.NotNil(t, prefill.Data.Amount, "amount suggestion is present")
	assert.Equal(t, domain.Money(1799), prefill.Data.Amount.Value)
	require.NotNil(t, prefill.Data.Name, "name suggestion is present")
	assert.Equal(t, "Zzyzx Grocery", prefill.Data.Name.Value)

	require.Len(t, prefill.Data.SubTransactions, 2, "the canned prefill proposes two line items")
	assert.Equal(t, "Zzyzx Coffee Beans", prefill.Data.SubTransactions[0].Name)
	assert.Equal(t, "pantry", prefill.Data.SubTransactions[0].Label)
	assert.Equal(t, domain.Money(1234), prefill.Data.SubTransactions[0].Amount)
	assert.Equal(t, "Zzyzx Oat Milk", prefill.Data.SubTransactions[1].Name)
	assert.Equal(t, "dairy-free", prefill.Data.SubTransactions[1].Label)
	assert.Equal(t, domain.Money(565), prefill.Data.SubTransactions[1].Amount)

	// Accept both items: the prefill already carries minor-unit amounts, so
	// they map straight into the create request and the derived sum is theirs.
	accepted := make([]schema.TransactionSubTransactionCreateRequest, 0, len(prefill.Data.SubTransactions))
	var derivedSum domain.Money
	for _, item := range prefill.Data.SubTransactions {
		accepted = append(accepted, schema.TransactionSubTransactionCreateRequest{
			Name: item.Name, Label: item.Label, Amount: item.Amount,
		})
		derivedSum += item.Amount
	}
	require.Equal(t, domain.Money(1799), derivedSum, "12.34 + 5.65")

	// 3. Create the transaction with the accepted items and the extracted draft.
	// The top-level amount is deliberately conflicting: with items present the
	// server ignores it and derives Σ(items) instead.
	var txResp utils.Envelope[schema.TransactionResponse]
	status = apiClient.Do(t, http.MethodPost, "/api/v1/transactions", token, schema.TransactionCreateRequest{
		Type:            domain.TransactionTypeExpense,
		Amount:          domain.Money(9999), // ignored: items are present
		CurrencyCode:    "ZZD",
		CategoryID:      categoryResp.Data.ID,
		UserID:          seededUser.ID,
		Source:          domain.TransactionSourceManual,
		Name:            "E2E Sub-Transaction Receipt",
		AttachmentIDs:   []uuid.UUID{draftID},
		SubTransactions: accepted,
	}, &txResp)
	require.Equal(t, http.StatusCreated, status)
	txID := txResp.Data.ID
	require.NotZero(t, txID)
	assert.Equal(t, derivedSum, txResp.Data.Amount, "amount is derived as Σ(items); the sent amount is ignored")
	require.Len(t, txResp.Data.Attachments, 1, "the extracted draft is linked by the create")
	assert.Equal(t, draftID, txResp.Data.Attachments[0].ID)
	require.Len(t, txResp.Data.SubTransactions, 2)

	// 4. GET returns the same items (always present, never null) and the
	// derived amount.
	readBack := getTransactionDetails(t, token, txID)
	assert.Equal(t, derivedSum, readBack.Amount)
	require.NotNil(t, readBack.SubTransactions, "sub_transactions must be present, never null")
	require.Len(t, readBack.SubTransactions, 2)
	for _, item := range accepted {
		row := subTransactionByName(t, readBack.SubTransactions, item.Name)
		assert.Equal(t, item.Label, row.Label)
		assert.Equal(t, item.Amount, row.Amount)
	}

	// 5. Add a line item to the existing transaction: the amount recomputes.
	var addResp utils.Envelope[schema.TransactionResponse]
	status = apiClient.Do(t, http.MethodPost,
		fmt.Sprintf("/api/v1/transactions/%d/sub-transactions", txID), token,
		schema.TransactionSubTransactionCreateRequest{Name: "Zzyzx Almonds", Label: "snacks", Amount: domain.Money(200)},
		&addResp)
	require.Equal(t, http.StatusCreated, status)
	assert.Equal(t, domain.Money(1999), addResp.Data.Amount, "1234 + 565 + 200")
	require.Len(t, addResp.Data.SubTransactions, 3)
	almond := subTransactionByName(t, addResp.Data.SubTransactions, "Zzyzx Almonds")

	// 6. Patch that item's amount: the amount recomputes again.
	newAmount := domain.Money(300)
	var patchResp utils.Envelope[schema.TransactionResponse]
	status = apiClient.Do(t, http.MethodPatch,
		fmt.Sprintf("/api/v1/transactions/%d/sub-transactions/%s", txID, almond.ID), token,
		schema.TransactionSubTransactionUpdateRequest{Amount: &newAmount}, &patchResp)
	require.Equal(t, http.StatusOK, status)
	assert.Equal(t, domain.Money(2099), patchResp.Data.Amount, "1234 + 565 + 300")
	almond = subTransactionByName(t, patchResp.Data.SubTransactions, "Zzyzx Almonds")

	// While items exist, a direct transaction-level amount patch is rejected
	// rather than silently overwriting the derived total.
	conflictAmount := domain.Money(1)
	var conflictResp utils.Envelope[struct{}]
	status = apiClient.Do(t, http.MethodPatch, fmt.Sprintf("/api/v1/transactions/%d", txID), token,
		schema.TransactionUpdateRequest{Amount: &conflictAmount}, &conflictResp)
	assert.Equal(t, http.StatusBadRequest, status)
	assert.Equal(t, domain.Money(2099), getTransactionDetails(t, token, txID).Amount)

	// 7. Delete rows one at a time; the amount follows the remaining items.
	oat := subTransactionByName(t, patchResp.Data.SubTransactions, "Zzyzx Oat Milk")
	status = apiClient.Do(t, http.MethodDelete,
		fmt.Sprintf("/api/v1/transactions/%d/sub-transactions/%s", txID, oat.ID), token, nil, nil)
	require.Equal(t, http.StatusNoContent, status)
	assert.Equal(t, domain.Money(1534), getTransactionDetails(t, token, txID).Amount, "1234 + 300")

	status = apiClient.Do(t, http.MethodDelete,
		fmt.Sprintf("/api/v1/transactions/%d/sub-transactions/%s", txID, almond.ID), token, nil, nil)
	require.Equal(t, http.StatusNoContent, status)
	assert.Equal(t, domain.Money(1234), getTransactionDetails(t, token, txID).Amount, "coffee only")

	coffee := subTransactionByName(t, patchResp.Data.SubTransactions, "Zzyzx Coffee Beans")
	status = apiClient.Do(t, http.MethodDelete,
		fmt.Sprintf("/api/v1/transactions/%d/sub-transactions/%s", txID, coffee.ID), token, nil, nil)
	require.Equal(t, http.StatusNoContent, status)

	// Deleting the last item leaves the amount at its last computed value — it
	// must NOT be zeroed — and the item list is empty, not null.
	emptied := getTransactionDetails(t, token, txID)
	assert.Equal(t, domain.Money(1234), emptied.Amount, "the last delete must not zero the amount")
	require.NotNil(t, emptied.SubTransactions, "sub_transactions must be present, never null")
	assert.Empty(t, emptied.SubTransactions)

	// 8. With no items left the amount is unlocked: a direct patch is accepted.
	directAmount := domain.Money(4200)
	var amountResp utils.Envelope[schema.TransactionDetailsResponse]
	status = apiClient.Do(t, http.MethodPatch, fmt.Sprintf("/api/v1/transactions/%d", txID), token,
		schema.TransactionUpdateRequest{Amount: &directAmount}, &amountResp)
	require.Equal(t, http.StatusOK, status)
	assert.Equal(t, directAmount, amountResp.Data.Amount)

	// Cleanup, matching the other transaction-creating journeys (CRUD, attachment):
	// leaving a ZZD transaction behind would break the summary journey, which
	// converts every transaction into its ZZA target and 404s on a currency with
	// no ZZD->ZZA rate. Deleting cascades to the line items and the linked draft.
	status = apiClient.Do(t, http.MethodDelete, fmt.Sprintf("/api/v1/transactions/%d", txID), token, nil, nil)
	require.Equal(t, http.StatusNoContent, status)
}

// getTransactionDetails GETs one transaction and returns its detail body,
// failing the test unless the read succeeds, so callers can inspect the
// parent amount and hydrated line items after a mutation.
func getTransactionDetails(t testing.TB, token string, transactionID int64) schema.TransactionDetailsResponse {
	t.Helper()
	var details utils.Envelope[schema.TransactionDetailsResponse]
	status := apiClient.Do(t, http.MethodGet, fmt.Sprintf("/api/v1/transactions/%d", transactionID), token, nil, &details)
	require.Equal(t, http.StatusOK, status)
	return details.Data
}

// subTransactionByName finds a returned line item by its name, so assertions
// don't depend on the server's ordering of the hydrated rows.
func subTransactionByName(t testing.TB, rows []schema.TransactionSubTransactionResponse, name string) schema.TransactionSubTransactionResponse {
	t.Helper()
	for _, row := range rows {
		if row.Name == name {
			return row
		}
	}
	t.Fatalf("no sub-transaction named %q in %+v", name, rows)
	return schema.TransactionSubTransactionResponse{}
}
