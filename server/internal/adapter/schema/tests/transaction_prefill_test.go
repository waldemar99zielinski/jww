package schema_test

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"jww/internal/adapter/schema"
	"jww/internal/domain"
)

// TestPrefillFieldResponse_JSONShape pins the per-field wire shape the client's
// PrefillField<T> expects: {value, confidence?, source?}, with confidence and
// source omitted when absent (never nulled).
func TestPrefillFieldResponse_JSONShape(t *testing.T) {
	tests := []struct {
		name  string
		field schema.PrefillFieldResponse[string]
		want  string
	}{
		{
			name:  "value only omits confidence and source",
			field: schema.PrefillFieldResponse[string]{Value: "24.90"},
			want:  `{"value":"24.90"}`,
		},
		{
			name: "value with confidence and source",
			field: schema.PrefillFieldResponse[string]{
				Value: "24.90", Confidence: ptr(91.0), Source: "receipt total line",
			},
			want: `{"value":"24.90","confidence":91,"source":"receipt total line"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			raw, err := json.Marshal(tt.field)
			require.NoError(t, err)
			assert.JSONEq(t, tt.want, string(raw))
		})
	}
}

func TestTransactionPrefillResponse_SerializesFieldShapeAndOmitsAbsent(t *testing.T) {
	amount := &schema.PrefillFieldResponse[domain.Money]{Value: domain.Money(2490)}
	txType := &schema.PrefillFieldResponse[domain.TransactionType]{Value: domain.TransactionTypeExpense}
	resp := schema.TransactionPrefillResponse{
		AttachmentID: uuid.New(),
		Type:         txType,
		Amount:       amount,
		Name:         &schema.PrefillFieldResponse[string]{Value: "Corner Market"},
		// Description / CurrencyCode / OccurredOn / CategoryHint deliberately absent.
		SubTransactions: []schema.PrefillSubTransactionResponse{
			{Name: "Milk 2%", Label: "dairy", Amount: domain.Money(349)},
		},
	}

	raw, err := json.Marshal(resp)
	require.NoError(t, err)
	body := string(raw)

	assert.Contains(t, body, `"attachment_id"`)
	assert.Contains(t, body, `"value":2490`)
	assert.Contains(t, body, `"sub_transactions"`)
	assert.Contains(t, body, `"Milk 2%"`)
	assert.NotContains(t, body, `"description"`, "an absent field is omitted, not nulled")
	assert.NotContains(t, body, `"confidence"`, "confidence is omitted when the model gave none")
}

func TestTransactionPrefillResponse_NoSubTransactionsIsAnEmptyArray(t *testing.T) {
	// The producer supplies a non-nil slice; an explicit empty slice must encode
	// as [] rather than null.
	resp := schema.TransactionPrefillResponse{
		AttachmentID:    uuid.New(),
		SubTransactions: []schema.PrefillSubTransactionResponse{},
	}

	raw, err := json.Marshal(resp)
	require.NoError(t, err)
	assert.Contains(t, string(raw), `"sub_transactions":[]`)
}

// TestTransactionPrefillResponse_MapsSubTransactionsAndType proves the DTO
// values map through verbatim on the wire: type carries a domain.TransactionType
// value, and each sub-transaction carries name/label/amount in integer cents.
func TestTransactionPrefillResponse_MapsSubTransactionsAndType(t *testing.T) {
	type wireField[T any] struct {
		Value      T        `json:"value"`
		Confidence *float64 `json:"confidence"`
		Source     string   `json:"source"`
	}
	type wireSubTransaction struct {
		Name   string       `json:"name"`
		Label  string       `json:"label"`
		Amount domain.Money `json:"amount"`
	}
	var wire struct {
		Type            *wireField[domain.TransactionType] `json:"type"`
		SubTransactions []wireSubTransaction               `json:"sub_transactions"`
	}

	resp := schema.TransactionPrefillResponse{
		AttachmentID: uuid.New(),
		Type:         &schema.PrefillFieldResponse[domain.TransactionType]{Value: domain.TransactionTypeIncome},
		SubTransactions: []schema.PrefillSubTransactionResponse{
			{Name: "Milk 2%", Label: "dairy", Amount: domain.Money(349)},
			{Name: "Bread", Label: "", Amount: domain.Money(219)},
		},
	}

	raw, err := json.Marshal(resp)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(raw, &wire))

	require.NotNil(t, wire.Type)
	assert.Equal(t, domain.TransactionTypeIncome, wire.Type.Value)
	require.Len(t, wire.SubTransactions, 2)
	assert.Equal(t, wireSubTransaction{Name: "Milk 2%", Label: "dairy", Amount: domain.Money(349)}, wire.SubTransactions[0])
	assert.Equal(t, wireSubTransaction{Name: "Bread", Label: "", Amount: domain.Money(219)}, wire.SubTransactions[1])
}
