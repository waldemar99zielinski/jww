package schema

import (
	"github.com/google/uuid"

	"jww/internal/domain"
)

// This file holds the wire shape of POST
// /transaction-attachments/{id}/extraction. The service-level Prefill
// (jww/internal/service/transaction_attachment_extraction) is mapped into these
// DTOs at the HTTP boundary, in the extraction handler, because this package
// must not import service packages (see this package's README). Build
// SubTransactions as a non-nil slice (empty when nothing was itemized) so it
// always encodes as [].

// PrefillFieldResponse is the wire shape of one proposed value:
// {value, confidence?, source?}. Confidence, when present, is the extractor's
// 0-100 score.
type PrefillFieldResponse[T any] struct {
	Value      T        `json:"value"`
	Confidence *float64 `json:"confidence,omitempty"`
	Source     string   `json:"source,omitempty"`
}

// PrefillSubTransactionResponse is one proposed line item, with amount in
// integer cents (negative for discounts). Confidence is the extractor's
// 0-100 score.
type PrefillSubTransactionResponse struct {
	Name       string       `json:"name"`
	Label      string       `json:"label"`
	Amount     domain.Money `json:"amount"`
	Confidence *float64     `json:"confidence,omitempty"`
}

// TransactionPrefillResponse mirrors the client's TransactionPrefillDto. Every
// scalar field is a nullable PrefillFieldResponse and is omitted entirely when
// the model did not propose it; sub_transactions is always present (an empty
// array when nothing was itemized).
type TransactionPrefillResponse struct {
	AttachmentID    uuid.UUID                                     `json:"attachment_id"`
	AttachmentIDs   []uuid.UUID                                   `json:"attachment_ids"`
	Type            *PrefillFieldResponse[domain.TransactionType] `json:"type,omitempty"`
	Amount          *PrefillFieldResponse[domain.Money]           `json:"amount,omitempty"`
	CurrencyCode    *PrefillFieldResponse[string]                 `json:"currency_code,omitempty"`
	OccurredOn      *PrefillFieldResponse[string]                 `json:"occurred_on,omitempty"`
	OccurredAt      *PrefillFieldResponse[string]                 `json:"occurred_at,omitempty"`
	Name            *PrefillFieldResponse[string]                 `json:"name,omitempty"`
	Description     *PrefillFieldResponse[string]                 `json:"description,omitempty"`
	CategoryHint    *PrefillFieldResponse[string]                 `json:"category_hint,omitempty"`
	SubTransactions []PrefillSubTransactionResponse               `json:"sub_transactions"`
	Warnings        []string                                      `json:"warnings,omitempty"`
	DurationMs      int64                                         `json:"duration_ms,omitempty"`
}
