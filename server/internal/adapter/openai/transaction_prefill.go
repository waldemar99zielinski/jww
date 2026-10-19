package openai

import (
	"encoding/json"
	"fmt"
	"strings"

	"jww/internal/domain"
	transaction_attachment_extraction "jww/internal/service/transaction_attachment_extraction"
)

// transactionPrefillSchemaName names the strict response format for the receipt
// flow. It must match [a-zA-Z0-9_-]{1,64}.
const transactionPrefillSchemaName = "transaction_prefill"

// ExtractTransactionPrefill sends one vision request containing every image URL
// and maps its structured output into the service's transaction-prefill result.
// Every provider/decode failure returns a plain error; the service maps it to
// the public 503 kind.
func (c *Client) ExtractTransactionPrefill(
	ctx domain.AppContext,
	input transaction_attachment_extraction.TransactionPrefillInput,
) (*transaction_attachment_extraction.TransactionPrefillResult, error) {
	content, err := c.structuredContent(ctx, c.visionParams(
		transactionPrefillInstruction(input),
		transactionPrefillSchemaName,
		transactionPrefillSchema(),
		input.ImageURLs,
	))
	if err != nil {
		return nil, err
	}
	return decodeTransactionPrefill(content)
}

// transactionPrefillSchema returns the strict JSON Schema for this flow
// (response_format.json_schema.schema). Strict mode requires
// additionalProperties:false and every property in required, so absent values
// are explicit nulls. Every present scalar/line item carries an integer
// confidence from 0 to 100. sub_transactions is always an array.
func transactionPrefillSchema() map[string]any {
	scoredString := map[string]any{
		"type":                 []string{"object", "null"},
		"additionalProperties": false,
		"properties": map[string]any{
			"value":      map[string]any{"type": []string{"string", "null"}},
			"confidence": confidenceSchema(),
		},
		"required": []string{"value", "confidence"},
	}

	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"properties": map[string]any{
			"type":          scoredString,
			"amount":        scoredString,
			"currency_code": scoredString,
			"occurred_on":   scoredString,
			"occurred_at":   scoredString,
			"name":          scoredString,
			"description":   scoredString,
			"category_hint": scoredString,
			"sub_transactions": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type":                 "object",
					"additionalProperties": false,
					"properties": map[string]any{
						"name":       map[string]any{"type": "string"},
						"label":      map[string]any{"type": "string"},
						"amount":     map[string]any{"type": "string"},
						"confidence": confidenceSchema(),
					},
					"required": []string{"name", "label", "amount", "confidence"},
				},
			},
		},
		"required": []string{
			"type", "amount", "currency_code", "occurred_on", "occurred_at",
			"name", "description", "category_hint", "sub_transactions",
		},
	}
}

// confidenceSchema intentionally relies on type + prompt wording rather than
// minimum/maximum: those numeric constraints are not part of the strict
// structured-output subset every OpenAI-compatible backend accepts.
// validateConfidence enforces the 0-100 range after decoding.
func confidenceSchema() map[string]any {
	return map[string]any{"type": "integer"}
}

// validateConfidence guards against providers that ignore the JSON Schema
// bounds: every present confidence must be an integer in 0-100.
func (r transactionPrefillResponse) validateConfidence() error {
	fields := map[string]*transactionPrefillField{
		"type": r.Type, "amount": r.Amount, "currency_code": r.CurrencyCode,
		"occurred_on": r.OccurredOn, "occurred_at": r.OccurredAt, "name": r.Name,
		"description": r.Description, "category_hint": r.CategoryHint,
	}
	for name, field := range fields {
		if field == nil {
			continue
		}
		if field.Confidence == nil {
			return fmt.Errorf("openai: %s confidence is missing", name)
		}
		if *field.Confidence < 0 || *field.Confidence > 100 {
			return fmt.Errorf("openai: %s confidence %d is outside 0-100", name, *field.Confidence)
		}
	}
	for i, item := range r.SubTransactions {
		if item.Confidence == nil {
			return fmt.Errorf("openai: sub_transactions[%d] confidence is missing", i)
		}
		if *item.Confidence < 0 || *item.Confidence > 100 {
			return fmt.Errorf("openai: sub_transactions[%d] confidence %d is outside 0-100", i, *item.Confidence)
		}
	}
	return nil
}

// transactionPrefillField is one scored scalar. Value is a pointer so an
// explicit JSON null (the forced representation of "not visible") stays
// distinguishable from a present value; Confidence is 0-100.
type transactionPrefillField struct {
	Value      *string `json:"value"`
	Confidence *int    `json:"confidence"`
}

// transactionPrefillResponse is the flat structured-output payload for this
// flow. Every scalar is nullable; sub_transactions is always present.
type transactionPrefillResponse struct {
	Type            *transactionPrefillField `json:"type"`
	Amount          *transactionPrefillField `json:"amount"`
	CurrencyCode    *transactionPrefillField `json:"currency_code"`
	OccurredOn      *transactionPrefillField `json:"occurred_on"`
	OccurredAt      *transactionPrefillField `json:"occurred_at"`
	Name            *transactionPrefillField `json:"name"`
	Description     *transactionPrefillField `json:"description"`
	CategoryHint    *transactionPrefillField `json:"category_hint"`
	SubTransactions []prefillSubTransaction  `json:"sub_transactions"`
}

// prefillSubTransaction is one decoded line item. label may be non-empty for a
// discount/adjustment; amount may be negative for such lines.
type prefillSubTransaction struct {
	Name       string `json:"name"`
	Label      string `json:"label"`
	Amount     string `json:"amount"`
	Confidence *int   `json:"confidence"`
}

// decodeTransactionPrefill turns structured content into the provider-agnostic
// result. Invalid JSON is an error; every present, non-null, non-empty scalar
// becomes a PrefillField with its 0-100 confidence. Absent values stay nil.
func decodeTransactionPrefill(content string) (*transaction_attachment_extraction.TransactionPrefillResult, error) {
	var decoded transactionPrefillResponse
	if err := json.Unmarshal([]byte(content), &decoded); err != nil {
		return nil, fmt.Errorf("openai: decoding structured output: %w", err)
	}
	if err := decoded.validateConfidence(); err != nil {
		return nil, err
	}

	amount, err := prefillMoney(decoded.Amount)
	if err != nil {
		return nil, err
	}
	result := &transaction_attachment_extraction.TransactionPrefillResult{
		Amount:       amount,
		CurrencyCode: prefillString(decoded.CurrencyCode),
		OccurredOn:   prefillString(decoded.OccurredOn),
		OccurredAt:   prefillString(decoded.OccurredAt),
		Name:         prefillString(decoded.Name),
		Description:  prefillString(decoded.Description),
		CategoryHint: prefillString(decoded.CategoryHint),
	}

	if value := fieldValue(decoded.Type); value != nil {
		result.Type = &transaction_attachment_extraction.PrefillField[domain.TransactionType]{
			Value:      domain.TransactionType(*value),
			Confidence: confidenceValue(decoded.Type),
		}
	}

	for i, sub := range decoded.SubTransactions {
		amount, err := domain.ParseMoney(sub.Amount)
		if err != nil {
			return nil, fmt.Errorf("openai: decoding sub_transactions[%d] amount: %w", i, err)
		}
		result.SubTransactions = append(result.SubTransactions, transaction_attachment_extraction.PrefillSubTransaction{
			Name:       sub.Name,
			Label:      sub.Label,
			Amount:     amount,
			Confidence: confidenceValue(&transactionPrefillField{Confidence: sub.Confidence}),
		})
	}

	return result, nil
}

// prefillMoney maps the model's decimal amount string into integer cents.
func prefillMoney(field *transactionPrefillField) (*transaction_attachment_extraction.PrefillField[domain.Money], error) {
	value := fieldValue(field)
	if value == nil {
		return nil, nil
	}
	amount, err := domain.ParseMoney(*value)
	if err != nil {
		return nil, fmt.Errorf("openai: decoding amount: %w", err)
	}
	return &transaction_attachment_extraction.PrefillField[domain.Money]{
		Value:      amount,
		Confidence: confidenceValue(field),
	}, nil
}

// prefillString maps one scored nullable scalar into an optional PrefillField.
func prefillString(field *transactionPrefillField) *transaction_attachment_extraction.PrefillField[string] {
	value := fieldValue(field)
	if value == nil {
		return nil
	}
	return &transaction_attachment_extraction.PrefillField[string]{
		Value:      *value,
		Confidence: confidenceValue(field),
	}
}

func fieldValue(field *transactionPrefillField) *string {
	if field == nil {
		return nil
	}
	trimmed := trimmedValue(field.Value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func confidenceValue(field *transactionPrefillField) *float64 {
	if field == nil || field.Confidence == nil {
		return nil
	}
	value := float64(*field.Confidence)
	return &value
}

// trimmedValue dereferences a nullable string and trims surrounding whitespace;
// nil and whitespace-only values both come back as "" (treated as absent).
func trimmedValue(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}
