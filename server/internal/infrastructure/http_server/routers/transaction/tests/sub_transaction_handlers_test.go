package transaction_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"jww/internal/adapter/schema"
	"jww/internal/domain"
)

// postSubTransaction POSTs one line item onto the path transaction and returns
// the raw response recorder.
func postSubTransaction(t *testing.T, r http.Handler, transactionID string, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/v1/transactions/"+transactionID+"/sub-transactions", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func patchSubTransaction(t *testing.T, r http.Handler, transactionID, subID string, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPatch, "/v1/transactions/"+transactionID+"/sub-transactions/"+subID, bytes.NewBufferString(body))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func deleteSubTransaction(t *testing.T, r http.Handler, transactionID, subID string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodDelete, "/v1/transactions/"+transactionID+"/sub-transactions/"+subID, nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

// seedSubTransaction creates one line item through the HTTP API and returns the
// id of the item named name (the create response carries every item on the
// transaction, so the newly created one is found by name).
func seedSubTransaction(t *testing.T, r http.Handler, transactionID int64, name, body string) uuid.UUID {
	t.Helper()
	rec := postSubTransaction(t, r, itoa(transactionID), body)
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
	var resp schema.TransactionResponse
	decodeData(t, rec, &resp)
	for _, item := range resp.SubTransactions {
		if item.Name == name {
			return item.ID
		}
	}
	t.Fatalf("no sub-transaction named %q in response %s", name, rec.Body.String())
	return uuid.Nil
}

func TestCreateSubTransactionHandler(t *testing.T) {
	t.Run("happy path returns 201 with the recomputed amount and a non-null item list", func(t *testing.T) {
		r, fx := newTestRouter(t)
		txID := seedTransaction(t, fx, "shop")

		rec := postSubTransaction(t, r, itoa(txID), `{"name":"Milk","label":"dairy","amount":349}`)

		require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
		assert.Contains(t, rec.Body.String(), `"sub_transactions":[`)
		var resp schema.TransactionResponse
		decodeData(t, rec, &resp)
		assert.Equal(t, domain.Money(349), resp.Amount, "the amount is derived from the item")
		require.Len(t, resp.SubTransactions, 1)
		assert.Equal(t, "Milk", resp.SubTransactions[0].Name)
		assert.Equal(t, "dairy", resp.SubTransactions[0].Label)
		assert.Equal(t, domain.Money(349), resp.SubTransactions[0].Amount)
	})

	t.Run("a second item accumulates into the parent amount", func(t *testing.T) {
		r, fx := newTestRouter(t)
		txID := seedTransaction(t, fx, "shop")
		seedSubTransaction(t, r, txID, "Milk", `{"name":"Milk","label":"dairy","amount":349}`)

		rec := postSubTransaction(t, r, itoa(txID), `{"name":"Bread","label":"bakery","amount":219}`)

		require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
		var resp schema.TransactionResponse
		decodeData(t, rec, &resp)
		assert.Equal(t, domain.Money(568), resp.Amount)
		require.Len(t, resp.SubTransactions, 2)
	})

	t.Run("malformed JSON is 400", func(t *testing.T) {
		r, fx := newTestRouter(t)
		txID := seedTransaction(t, fx, "shop")
		rec := postSubTransaction(t, r, itoa(txID), `{"name":`)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("a blank name is 400", func(t *testing.T) {
		r, fx := newTestRouter(t)
		txID := seedTransaction(t, fx, "shop")
		rec := postSubTransaction(t, r, itoa(txID), `{"name":"   ","amount":349}`)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("a zero item amount is 400", func(t *testing.T) {
		r, fx := newTestRouter(t)
		txID := seedTransaction(t, fx, "shop")
		assert.Equal(t, http.StatusBadRequest, postSubTransaction(t, r, itoa(txID), `{"name":"Milk","amount":0}`).Code)
	})

	t.Run("a negative discount item is accepted when the total stays positive", func(t *testing.T) {
		r, fx := newTestRouter(t)
		txID := seedTransaction(t, fx, "shop")
		seedSubTransaction(t, r, txID, "Milk", `{"name":"Milk","amount":500}`)
		rec := postSubTransaction(t, r, itoa(txID), `{"name":"Coupon","label":"discount","amount":-100}`)
		require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
		var resp schema.TransactionResponse
		decodeData(t, rec, &resp)
		require.Len(t, resp.SubTransactions, 2)
		amounts := []domain.Money{resp.SubTransactions[0].Amount, resp.SubTransactions[1].Amount}
		assert.ElementsMatch(t, []domain.Money{500, -100}, amounts)
		assert.Equal(t, domain.Money(400), resp.Amount)
	})

	t.Run("a missing transaction is 404", func(t *testing.T) {
		r, _ := newTestRouter(t)
		rec := postSubTransaction(t, r, "999999", `{"name":"Milk","amount":349}`)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("a malformed transaction id is 400", func(t *testing.T) {
		r, _ := newTestRouter(t)
		rec := postSubTransaction(t, r, "not-a-number", `{"name":"Milk","amount":349}`)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func TestUpdateSubTransactionHandler(t *testing.T) {
	t.Run("a partial amount patch recomputes the parent amount", func(t *testing.T) {
		r, fx := newTestRouter(t)
		txID := seedTransaction(t, fx, "shop")
		subID := seedSubTransaction(t, r, txID, "Milk", `{"name":"Milk","amount":349}`)

		rec := patchSubTransaction(t, r, itoa(txID), subID.String(), `{"amount":500}`)

		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		var resp schema.TransactionResponse
		decodeData(t, rec, &resp)
		assert.Equal(t, domain.Money(500), resp.Amount)
		require.Len(t, resp.SubTransactions, 1)
		assert.Equal(t, domain.Money(500), resp.SubTransactions[0].Amount)
	})

	t.Run("name and label patch without touching the amount", func(t *testing.T) {
		r, fx := newTestRouter(t)
		txID := seedTransaction(t, fx, "shop")
		subID := seedSubTransaction(t, r, txID, "Milk", `{"name":"Milk","label":"","amount":349}`)

		rec := patchSubTransaction(t, r, itoa(txID), subID.String(), `{"name":"Oat milk","label":"dairy"}`)

		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		var resp schema.TransactionResponse
		decodeData(t, rec, &resp)
		require.Len(t, resp.SubTransactions, 1)
		assert.Equal(t, "Oat milk", resp.SubTransactions[0].Name)
		assert.Equal(t, "dairy", resp.SubTransactions[0].Label)
		assert.Equal(t, domain.Money(349), resp.SubTransactions[0].Amount)
		assert.Equal(t, domain.Money(349), resp.Amount)
	})

	t.Run("malformed JSON is 400", func(t *testing.T) {
		r, fx := newTestRouter(t)
		txID := seedTransaction(t, fx, "shop")
		subID := seedSubTransaction(t, r, txID, "Milk", `{"name":"Milk","amount":349}`)
		rec := patchSubTransaction(t, r, itoa(txID), subID.String(), `{"name":`)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("a blank name is 400", func(t *testing.T) {
		r, fx := newTestRouter(t)
		txID := seedTransaction(t, fx, "shop")
		subID := seedSubTransaction(t, r, txID, "Milk", `{"name":"Milk","amount":349}`)
		rec := patchSubTransaction(t, r, itoa(txID), subID.String(), `{"name":"  "}`)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("a non-positive amount is 400", func(t *testing.T) {
		r, fx := newTestRouter(t)
		txID := seedTransaction(t, fx, "shop")
		subID := seedSubTransaction(t, r, txID, "Milk", `{"name":"Milk","amount":349}`)
		rec := patchSubTransaction(t, r, itoa(txID), subID.String(), `{"amount":0}`)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("a missing transaction is 404", func(t *testing.T) {
		r, _ := newTestRouter(t)
		rec := patchSubTransaction(t, r, "999999", uuid.NewString(), `{"amount":500}`)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("a missing sub-transaction is 404", func(t *testing.T) {
		r, fx := newTestRouter(t)
		txID := seedTransaction(t, fx, "shop")
		rec := patchSubTransaction(t, r, itoa(txID), uuid.NewString(), `{"amount":500}`)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("a sub-transaction belonging to another transaction is 404", func(t *testing.T) {
		r, fx := newTestRouter(t)
		otherTx := seedTransaction(t, fx, "other")
		subID := seedSubTransaction(t, r, otherTx, "Milk", `{"name":"Milk","amount":349}`)
		targetTx := seedTransaction(t, fx, "target")

		rec := patchSubTransaction(t, r, itoa(targetTx), subID.String(), `{"amount":500}`)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("a malformed sub-transaction id is 400", func(t *testing.T) {
		r, fx := newTestRouter(t)
		txID := seedTransaction(t, fx, "shop")
		rec := patchSubTransaction(t, r, itoa(txID), "not-a-uuid", `{"amount":500}`)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func TestDeleteSubTransactionHandler(t *testing.T) {
	t.Run("happy path returns 204 and the parent amount recomputes", func(t *testing.T) {
		r, fx := newTestRouter(t)
		txID := seedTransaction(t, fx, "shop")
		seedSubTransaction(t, r, txID, "Milk", `{"name":"Milk","amount":349}`)
		second := seedSubTransaction(t, r, txID, "Bread", `{"name":"Bread","amount":219}`)

		rec := deleteSubTransaction(t, r, itoa(txID), second.String())

		require.Equal(t, http.StatusNoContent, rec.Code, rec.Body.String())

		req := httptest.NewRequest(http.MethodGet, "/v1/transactions/"+itoa(txID), nil)
		getRec := httptest.NewRecorder()
		r.ServeHTTP(getRec, req)
		require.Equal(t, http.StatusOK, getRec.Code)
		var got schema.TransactionDetailsResponse
		decodeData(t, getRec, &got)
		assert.Equal(t, domain.Money(349), got.Amount)
		require.Len(t, got.SubTransactions, 1)
		assert.Equal(t, "Milk", got.SubTransactions[0].Name)
	})

	t.Run("deleting the last item keeps the amount and releases it for editing", func(t *testing.T) {
		r, fx := newTestRouter(t)
		txID := seedTransaction(t, fx, "shop")
		only := seedSubTransaction(t, r, txID, "Milk", `{"name":"Milk","amount":349}`)

		rec := deleteSubTransaction(t, r, itoa(txID), only.String())
		require.Equal(t, http.StatusNoContent, rec.Code, rec.Body.String())

		req := httptest.NewRequest(http.MethodGet, "/v1/transactions/"+itoa(txID), nil)
		getRec := httptest.NewRecorder()
		r.ServeHTTP(getRec, req)
		require.Equal(t, http.StatusOK, getRec.Code)
		var got schema.TransactionDetailsResponse
		decodeData(t, getRec, &got)
		assert.Equal(t, domain.Money(349), got.Amount, "the last computed amount is kept, not zeroed")
		assert.NotNil(t, got.SubTransactions)
		assert.Empty(t, got.SubTransactions)

		// With no items left, a transaction-level amount patch is allowed again.
		patchRec := patchTransaction(t, r, txID, `{"amount":999}`)
		assert.Equal(t, http.StatusOK, patchRec.Code, patchRec.Body.String())
		var patched schema.TransactionDetailsResponse
		decodeData(t, patchRec, &patched)
		assert.Equal(t, domain.Money(999), patched.Amount)
	})

	t.Run("a missing sub-transaction is 404", func(t *testing.T) {
		r, fx := newTestRouter(t)
		txID := seedTransaction(t, fx, "shop")
		rec := deleteSubTransaction(t, r, itoa(txID), uuid.NewString())
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("a sub-transaction belonging to another transaction is 404", func(t *testing.T) {
		r, fx := newTestRouter(t)
		otherTx := seedTransaction(t, fx, "other")
		subID := seedSubTransaction(t, r, otherTx, "Milk", `{"name":"Milk","amount":349}`)
		targetTx := seedTransaction(t, fx, "target")

		rec := deleteSubTransaction(t, r, itoa(targetTx), subID.String())

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("a missing transaction is 404", func(t *testing.T) {
		r, _ := newTestRouter(t)
		rec := deleteSubTransaction(t, r, "999999", uuid.NewString())
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("a malformed sub-transaction id is 400", func(t *testing.T) {
		r, fx := newTestRouter(t)
		txID := seedTransaction(t, fx, "shop")
		rec := deleteSubTransaction(t, r, itoa(txID), "not-a-uuid")
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

// TestCreateTransactionWithSubTransactions covers the create-transaction route's
// new sub_transactions[] field: the derived amount wins over a conflicting
// top-level amount, and the response carries the items.
func TestCreateTransactionWithSubTransactions(t *testing.T) {
	t.Run("items derive the amount, ignoring the conflicting top-level amount", func(t *testing.T) {
		r, fx := newTestRouter(t)
		body := `{"type":"expense","amount":9999,"currency_code":"` + fx.currency.Code +
			`","category_id":"` + fx.category.ID.String() + `","user_id":"` + fx.user.ID.String() +
			`","source":"manual","name":"Groceries","sub_transactions":[` +
			`{"name":"Milk","label":"dairy","amount":349},{"name":"Bread","label":"bakery","amount":219}]}`
		req := httptest.NewRequest(http.MethodPost, "/v1/transactions", strings.NewReader(body))
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
		var resp schema.TransactionResponse
		decodeData(t, rec, &resp)
		assert.Equal(t, domain.Money(568), resp.Amount, "the sum of the items wins")
		require.Len(t, resp.SubTransactions, 2)
	})

	t.Run("a blank item name is 400", func(t *testing.T) {
		r, fx := newTestRouter(t)
		body := `{"type":"expense","amount":"0","currency_code":"` + fx.currency.Code +
			`","category_id":"` + fx.category.ID.String() + `","user_id":"` + fx.user.ID.String() +
			`","source":"manual","name":"Groceries","sub_transactions":[{"name":"","amount":349}]}`
		req := httptest.NewRequest(http.MethodPost, "/v1/transactions", strings.NewReader(body))
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}
