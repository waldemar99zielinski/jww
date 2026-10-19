package schema_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"jww/internal/adapter/schema"
	"jww/internal/domain"
)

// TestTransactionSubTransactionResponse_AmountIsBareJSONNumber pins the wire
// contract: amount goes out as an unquoted JSON number (like TransactionResponse
// .amount), never a decimal string.
func TestTransactionSubTransactionResponse_AmountIsBareJSONNumber(t *testing.T) {
	resp := schema.TransactionSubTransactionResponse{
		ID:     uuid.New(),
		Name:   "Milk 2%",
		Label:  "dairy",
		Amount: domain.Money(349),
	}

	raw, err := json.Marshal(resp)
	require.NoError(t, err)
	assert.Contains(t, string(raw), `"amount":349`, "amount must be a bare number, not a quoted string")
	assert.NotContains(t, string(raw), `"amount":"`)

	var roundTripped schema.TransactionSubTransactionResponse
	require.NoError(t, json.Unmarshal(raw, &roundTripped))
	assert.Equal(t, domain.Money(349), roundTripped.Amount)
	assert.Equal(t, "dairy", roundTripped.Label)
}

func TestTransactionSubTransactionsToResponse_AlwaysNonNil(t *testing.T) {
	t.Run("nil input becomes an empty (non-null) slice", func(t *testing.T) {
		out := schema.TransactionSubTransactionsToResponse(nil)
		require.NotNil(t, out)
		raw, err := json.Marshal(out)
		require.NoError(t, err)
		assert.Equal(t, "[]", string(raw))
	})

	t.Run("maps every field", func(t *testing.T) {
		id := uuid.New()
		out := schema.TransactionSubTransactionsToResponse([]domain.TransactionSubTransaction{
			{ID: id, TransactionID: 5, Name: "Bread", Label: "bakery", Amount: domain.Money(219)},
		})
		require.Len(t, out, 1)
		assert.Equal(t, id, out[0].ID)
		assert.Equal(t, "Bread", out[0].Name)
		assert.Equal(t, "bakery", out[0].Label)
		assert.Equal(t, domain.Money(219), out[0].Amount)
	})
}

func validSubTransactionCreateRequest() schema.TransactionSubTransactionCreateRequest {
	return schema.TransactionSubTransactionCreateRequest{Name: "Milk", Label: "dairy", Amount: domain.Money(349)}
}

func TestTransactionSubTransactionCreateRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*schema.TransactionSubTransactionCreateRequest)
		wantErr bool
	}{
		{"valid", func(r *schema.TransactionSubTransactionCreateRequest) {}, false},
		{"empty label allowed (free text)", func(r *schema.TransactionSubTransactionCreateRequest) { r.Label = "" }, false},
		{"blank name", func(r *schema.TransactionSubTransactionCreateRequest) { r.Name = "   " }, true},
		{"empty name", func(r *schema.TransactionSubTransactionCreateRequest) { r.Name = "" }, true},
		{"zero amount", func(r *schema.TransactionSubTransactionCreateRequest) { r.Amount = 0 }, true},
		{"negative discount amount", func(r *schema.TransactionSubTransactionCreateRequest) { r.Amount = -1 }, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := validSubTransactionCreateRequest()
			tt.mutate(&req)
			err := req.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTransactionSubTransactionUpdateRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     schema.TransactionSubTransactionUpdateRequest
		wantErr bool
	}{
		{"empty patch is valid", schema.TransactionSubTransactionUpdateRequest{}, false},
		{"valid name", schema.TransactionSubTransactionUpdateRequest{Name: ptr("Espresso")}, false},
		{"blank name", schema.TransactionSubTransactionUpdateRequest{Name: ptr("  ")}, true},
		{"valid amount", schema.TransactionSubTransactionUpdateRequest{Amount: ptr(domain.Money(1))}, false},
		{"zero amount", schema.TransactionSubTransactionUpdateRequest{Amount: ptr(domain.Money(0))}, true},
		{"negative discount amount", schema.TransactionSubTransactionUpdateRequest{Amount: ptr(domain.Money(-5))}, false},
		{"empty label is a real value", schema.TransactionSubTransactionUpdateRequest{Label: ptr("")}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestTransactionCreateRequest_AllowsZeroAmountOnlyWithSubTransactions encodes
// the resolved rule: when sub_transactions are present the top-level amount is
// derived server-side and may be zero/omitted; without them the amount is a
// required positive value.
func TestTransactionCreateRequest_AllowsZeroAmountOnlyWithSubTransactions(t *testing.T) {
	items := []schema.TransactionSubTransactionCreateRequest{
		{Name: "Milk", Label: "dairy", Amount: domain.Money(349)},
		{Name: "Bread", Amount: domain.Money(219)},
	}

	t.Run("zero amount with items is accepted", func(t *testing.T) {
		req := validTransactionCreateRequest()
		req.Amount = 0
		req.SubTransactions = items
		assert.NoError(t, req.Validate())
	})

	t.Run("omitted amount with items is accepted", func(t *testing.T) {
		req := validTransactionCreateRequest()
		req.Amount = 0
		req.SubTransactions = items
		assert.NoError(t, req.Validate())
	})

	t.Run("zero amount without items is rejected", func(t *testing.T) {
		req := validTransactionCreateRequest()
		req.Amount = 0
		assert.Error(t, req.Validate())
	})

	t.Run("an empty items slice does not unlock the zero amount", func(t *testing.T) {
		req := validTransactionCreateRequest()
		req.Amount = 0
		req.SubTransactions = []schema.TransactionSubTransactionCreateRequest{}
		assert.Error(t, req.Validate())
	})

	t.Run("a positive amount with items is accepted", func(t *testing.T) {
		req := validTransactionCreateRequest()
		req.SubTransactions = items
		assert.NoError(t, req.Validate())
	})

	t.Run("an invalid item is rejected", func(t *testing.T) {
		req := validTransactionCreateRequest()
		req.Amount = 0
		req.SubTransactions = []schema.TransactionSubTransactionCreateRequest{{Name: "   ", Amount: domain.Money(100)}}
		assert.Error(t, req.Validate())
	})

	t.Run("no items field at all is still valid with a positive amount", func(t *testing.T) {
		req := validTransactionCreateRequest()
		assert.NoError(t, req.Validate())
	})
}

// TestTransactionCreateRequest_ParsesSubTransactions proves the request DTO
// actually decodes the wire shape {name,label,amount} with a bare-number amount.
func TestTransactionCreateRequest_ParsesSubTransactions(t *testing.T) {
	body := `{"type":"expense","amount":2490,"currency_code":"USD",` +
		`"category_id":"` + uuid.NewString() + `","user_id":"` + uuid.NewString() + `",` +
		`"source":"manual","name":"Corner Market",` +
		`"sub_transactions":[{"name":"Milk 2%","label":"dairy","amount":349}]}`

	var req schema.TransactionCreateRequest
	require.NoError(t, json.Unmarshal([]byte(body), &req))
	require.Len(t, req.SubTransactions, 1)
	assert.Equal(t, "Milk 2%", req.SubTransactions[0].Name)
	assert.Equal(t, "dairy", req.SubTransactions[0].Label)
	assert.Equal(t, domain.Money(349), req.SubTransactions[0].Amount)
}

// TestTransactionResponses_CarrySubTransactions proves both response shapes
// expose the hydrated items, always non-null.
func TestTransactionResponses_CarrySubTransactions(t *testing.T) {
	tx := domain.Transaction{
		ID:       7,
		Type:     domain.TransactionTypeExpense,
		Amount:   domain.Money(568),
		Currency: domain.Currency{Code: "USD"},
		Category: domain.TransactionCategory{ID: uuid.New(), Name: "Groceries"},
		User:     domain.User{ID: uuid.New()},
		Source:   domain.TransactionSourceManual,
		Name:     "Corner Market",
		SubTransactions: []domain.TransactionSubTransaction{
			{ID: uuid.New(), TransactionID: 7, Name: "Milk", Label: "dairy", Amount: domain.Money(349), CreatedAt: time.Now().UTC()},
			{ID: uuid.New(), TransactionID: 7, Name: "Bread", Amount: domain.Money(219), CreatedAt: time.Now().UTC()},
		},
	}

	list := schema.TransactionToResponse(tx)
	require.Len(t, list.SubTransactions, 2)
	assert.Equal(t, "Milk", list.SubTransactions[0].Name)
	assert.Equal(t, domain.Money(349), list.SubTransactions[0].Amount)

	detail := schema.TransactionToDetailsResponse(tx, nil)
	require.Len(t, detail.SubTransactions, 2)
	assert.Equal(t, "Bread", detail.SubTransactions[1].Name)
}

func TestTransactionResponses_SubTransactionsNotEmptyAreNilInJSON(t *testing.T) {
	tx := domain.Transaction{ID: 1, Type: domain.TransactionTypeExpense, Amount: domain.Money(100)}

	listRaw, err := json.Marshal(schema.TransactionToResponse(tx))
	require.NoError(t, err)
	assert.Contains(t, string(listRaw), `"sub_transactions":[]`, "no items must still serialize as an empty array")

	detailRaw, err := json.Marshal(schema.TransactionToDetailsResponse(tx, nil))
	require.NoError(t, err)
	assert.Contains(t, string(detailRaw), `"sub_transactions":[]`)
}
