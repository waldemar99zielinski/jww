package transaction_attachment_extraction

import (
	"fmt"
	"strings"

	"github.com/google/uuid"

	"jww/internal/domain"
)

// PrefillField is one proposed scalar value plus the extractor's optional
// confidence and a short human-readable provenance string. It mirrors the
// client's TransactionPrefillDto/PrefillField<T> so a suggestion can be shown
// and accepted individually.
type PrefillField[T any] struct {
	Value      T
	Confidence *float64
	Source     string
}

// String renders one proposed scalar for humans: the value, plus its optional
// confidence and provenance. It is what the pretty log handler shows (via
// fmt's Stringer support); the JSON handler ignores it and marshals the fields.
func (f PrefillField[T]) String() string {
	switch {
	case f.Confidence != nil && f.Source != "":
		return fmt.Sprintf("%v (confidence %.0f, source: %s)", f.Value, *f.Confidence, f.Source)
	case f.Confidence != nil:
		return fmt.Sprintf("%v (confidence %.0f)", f.Value, *f.Confidence)
	case f.Source != "":
		return fmt.Sprintf("%v (source: %s)", f.Value, f.Source)
	default:
		return fmt.Sprintf("%v", f.Value)
	}
}

// PrefillSubTransaction is one proposed line item: a product/line name, a
// free-text label (never a category id), and a decimal-string amount matching
// the transaction amount convention. Iterated amounts may be negative for
// discounts; Confidence is the extractor's 0-100 score.
type PrefillSubTransaction struct {
	Name       string
	Label      string
	Amount     domain.Money
	Confidence *float64
}

// String renders one line item for humans: `Milk 2% (dairy): 3.49`, dropping
// the parenthesised label when the provider proposed none.
func (s PrefillSubTransaction) String() string {
	if s.Label == "" {
		return fmt.Sprintf("%s: %s", s.Name, s.Amount)
	}
	return fmt.Sprintf("%s (%s): %s", s.Name, s.Label, s.Amount)
}

// TransactionPrefillInput is the transaction-receipt extraction flow's input.
// ImageURLs contains ordered, short-lived content URLs for pages/views of one
// logical transaction. Categories are user-configurable prompt hints, passed
// separately so the adapter can state the allowed values for each type.
type TransactionPrefillInput struct {
	ImageURLs         []string
	ExpenseCategories []string
	IncomeCategories  []string
}

// TransactionPrefillResult is the provider-agnostic transaction extraction
// result. Every scalar is optional; a nil field means the extractor found
// nothing to propose.
type TransactionPrefillResult struct {
	Type            *PrefillField[domain.TransactionType]
	Amount          *PrefillField[domain.Money]
	CurrencyCode    *PrefillField[string]
	OccurredOn      *PrefillField[string]
	OccurredAt      *PrefillField[string]
	Name            *PrefillField[string]
	Description     *PrefillField[string]
	CategoryHint    *PrefillField[string]
	SubTransactions []PrefillSubTransaction
	Warnings        []string
}

// String renders the whole result as one readable line, e.g.
//
//	TransactionPrefillResult{type=expense (confidence 0.90), amount=24.90,
//	currency=USD, sub_transactions=[Milk 2% (dairy): 3.49]}
//
// Only proposed fields are shown. The pretty log handler uses this Stringer;
// the JSON handler marshals the struct fields instead.
func (r TransactionPrefillResult) String() string {
	parts := make([]string, 0, 10)
	if r.Type != nil {
		parts = append(parts, "type="+r.Type.String())
	}
	if r.Amount != nil {
		parts = append(parts, "amount="+r.Amount.String())
	}
	if r.CurrencyCode != nil {
		parts = append(parts, "currency="+r.CurrencyCode.String())
	}
	if r.OccurredOn != nil {
		parts = append(parts, "occurred_on="+r.OccurredOn.String())
	}
	if r.OccurredAt != nil {
		parts = append(parts, "occurred_at="+r.OccurredAt.String())
	}
	if r.Name != nil {
		parts = append(parts, "name="+r.Name.String())
	}
	if r.Description != nil {
		parts = append(parts, "description="+r.Description.String())
	}
	if r.CategoryHint != nil {
		parts = append(parts, "category_hint="+r.CategoryHint.String())
	}
	if len(r.SubTransactions) > 0 {
		items := make([]string, 0, len(r.SubTransactions))
		for _, sub := range r.SubTransactions {
			items = append(items, sub.String())
		}
		parts = append(parts, "sub_transactions=["+strings.Join(items, ", ")+"]")
	}
	if len(r.Warnings) > 0 {
		parts = append(parts, "warnings=["+strings.Join(r.Warnings, ", ")+"]")
	}
	return "TransactionPrefillResult{" + strings.Join(parts, ", ") + "}"
}

// Extractor is the provider-agnostic port implemented by an adapter (e.g.
// internal/adapter/openai). Each flow has its own method and matching input
// contract, so adding a flow cannot accidentally reuse another flow's prompt.
type Extractor interface {
	ExtractTransactionPrefill(ctx domain.AppContext, input TransactionPrefillInput) (*TransactionPrefillResult, error)
}

// AttachmentSource supplies the draft's metadata and a short-lived content
// grant. It is satisfied structurally by the transaction_attachment service
// (DraftAttachment wraps its draft-only lookup; ContentGrant wraps
// IssueDraftGrant) or by a small adapter wired in cmd/server.
type AttachmentSource interface {
	DraftAttachment(ctx domain.AppContext, id uuid.UUID) (*domain.TransactionAttachment, error)
	ContentGrant(ctx domain.AppContext, id uuid.UUID) (string, error)
}

// CategorySource supplies the active category names the transaction flow may
// return as category_hint, split by transaction type.
type CategorySource interface {
	TransactionCategories(ctx domain.AppContext) (expense, income []string, err error)
}
