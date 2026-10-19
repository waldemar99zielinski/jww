package schema

import (
	"fmt"
	"time"

	"github.com/google/uuid"

	"jww/internal/domain"
)

type TransactionCreateRequest struct {
	Type         domain.TransactionType   `json:"type"`
	Amount       domain.Money             `json:"amount"`
	CurrencyCode string                   `json:"currency_code"`
	CategoryID   uuid.UUID                `json:"category_id"`
	UserID       uuid.UUID                `json:"user_id"`
	Source       domain.TransactionSource `json:"source"`
	Name         string                   `json:"name"`
	Description  string                   `json:"description"`
	// CreatedAt is the transaction's date. When absent the server uses now().
	CreatedAt *time.Time `json:"created_at"`
	// AttachmentIDs optionally links already-uploaded unattached drafts to
	// the new transaction, atomically with its creation. Absent/empty keeps
	// the request valid and creates a transaction with no attachments.
	AttachmentIDs []uuid.UUID `json:"attachment_ids"`
	// SubTransactions optionally creates line items together with the
	// transaction in one unit of work. When present and non-empty the
	// top-level amount is derived server-side as their sum, so a zero/omitted
	// amount is accepted; without them a positive amount is required.
	SubTransactions []TransactionSubTransactionCreateRequest `json:"sub_transactions"`
}

func (r TransactionCreateRequest) Validate() error {
	if !r.Type.Valid() {
		return fmt.Errorf("type must be 'income' or 'expense'")
	}
	if len(r.SubTransactions) == 0 && r.Amount <= 0 {
		return fmt.Errorf("amount must be greater than 0")
	}
	if !r.Source.Valid() {
		return fmt.Errorf("source must be 'manual' or 'recurring'")
	}
	if r.CreatedAt != nil && r.CreatedAt.IsZero() {
		return fmt.Errorf("created_at must not be zero")
	}
	if err := firstError(
		exactLength("currency_code", r.CurrencyCode, 3),
		minLength("name", r.Name, 1),
	); err != nil {
		return err
	}
	if err := firstError(
		noNilUUIDs("attachment_ids", r.AttachmentIDs),
		noDuplicateUUIDs("attachment_ids", r.AttachmentIDs),
	); err != nil {
		return err
	}
	for i, item := range r.SubTransactions {
		if err := item.Validate(); err != nil {
			return fmt.Errorf("sub_transactions[%d]: %w", i, err)
		}
	}
	return nil
}

// TransactionUpdateRequest is a partial update for PATCH /transactions/{id}.
// Every field is optional: an absent (or null) field is left unchanged, and a
// present field is applied even when its value is empty (e.g. clearing the
// description). ID and attachment membership are not patchable here —
// attachments have their own routes. created_at is the transaction's date and
// is patchable.
type TransactionUpdateRequest struct {
	Type         *domain.TransactionType   `json:"type"`
	Amount       *domain.Money             `json:"amount"`
	CurrencyCode *string                   `json:"currency_code"`
	CategoryID   *uuid.UUID                `json:"category_id"`
	UserID       *uuid.UUID                `json:"user_id"`
	Source       *domain.TransactionSource `json:"source"`
	Name         *string                   `json:"name"`
	Description  *string                   `json:"description"`
	// CreatedAt is the transaction's date/time; the RFC3339 form GET returns.
	CreatedAt *time.Time `json:"created_at"`
}

// Validate checks only the fields that are present. Cross-field rules that
// depend on the merged result (e.g. type/category agreement) are enforced by
// the workflow once the current row is loaded.
func (r TransactionUpdateRequest) Validate() error {
	if r.Type != nil && !r.Type.Valid() {
		return fmt.Errorf("type must be 'income' or 'expense'")
	}
	if r.Source != nil && !r.Source.Valid() {
		return fmt.Errorf("source must be 'manual' or 'recurring'")
	}
	if r.Amount != nil && *r.Amount <= 0 {
		return fmt.Errorf("amount must be greater than 0")
	}
	if r.CreatedAt != nil && r.CreatedAt.IsZero() {
		return fmt.Errorf("created_at must not be zero")
	}
	if r.CurrencyCode != nil {
		if err := exactLength("currency_code", *r.CurrencyCode, 3); err != nil {
			return err
		}
	}
	if r.Name != nil {
		if err := minLength("name", *r.Name, 1); err != nil {
			return err
		}
	}
	return nil
}

type TransactionResponse struct {
	ID           int64                    `json:"id"`
	Type         domain.TransactionType   `json:"type"`
	Amount       domain.Money             `json:"amount"`
	CurrencyCode string                   `json:"currency_code"`
	CategoryID   uuid.UUID                `json:"category_id"`
	CategoryName string                   `json:"category_name"`
	UserID       uuid.UUID                `json:"user_id"`
	Source       domain.TransactionSource `json:"source"`
	Name         string                   `json:"name"`
	CreatedAt    time.Time                `json:"created_at"`
	Description  string                   `json:"description"`
	// Attachments is always present and non-null; an attachment-count field
	// is deliberately not added.
	Attachments []TransactionAttachmentResponse `json:"attachments"`
	// SubTransactions is always present and non-null.
	SubTransactions []TransactionSubTransactionResponse `json:"sub_transactions"`
}

// TransactionDetailsResponse is used only by GET /transactions/{id}. Unlike
// create/list responses, its attachments carry short-lived preview grants.
type TransactionDetailsResponse struct {
	ID           int64                                           `json:"id"`
	Type         domain.TransactionType                          `json:"type"`
	Amount       domain.Money                                    `json:"amount"`
	CurrencyCode string                                          `json:"currency_code"`
	CategoryID   uuid.UUID                                       `json:"category_id"`
	CategoryName string                                          `json:"category_name"`
	UserID       uuid.UUID                                       `json:"user_id"`
	Source       domain.TransactionSource                        `json:"source"`
	Name         string                                          `json:"name"`
	CreatedAt    time.Time                                       `json:"created_at"`
	Description  string                                          `json:"description"`
	Attachments  []TransactionAttachmentWithPreviewGrantResponse `json:"attachments"`
	// SubTransactions is always present and non-null.
	SubTransactions []TransactionSubTransactionResponse `json:"sub_transactions"`
}

func TransactionToDetailsResponse(t domain.Transaction, grants map[uuid.UUID]string) TransactionDetailsResponse {
	return TransactionDetailsResponse{
		ID:              t.ID,
		Type:            t.Type,
		Amount:          t.Amount,
		CurrencyCode:    t.Currency.Code,
		CategoryID:      t.Category.ID,
		CategoryName:    t.Category.Name,
		UserID:          t.User.ID,
		Source:          t.Source,
		Name:            t.Name,
		CreatedAt:       t.CreatedAt,
		Description:     t.Description,
		Attachments:     TransactionAttachmentsWithPreviewGrantsToResponse(t.Attachments, grants),
		SubTransactions: TransactionSubTransactionsToResponse(t.SubTransactions),
	}
}

func TransactionToResponse(t domain.Transaction) TransactionResponse {
	return TransactionResponse{
		ID:              t.ID,
		Type:            t.Type,
		Amount:          t.Amount,
		CurrencyCode:    t.Currency.Code,
		CategoryID:      t.Category.ID,
		CategoryName:    t.Category.Name,
		UserID:          t.User.ID,
		Source:          t.Source,
		Name:            t.Name,
		CreatedAt:       t.CreatedAt,
		Description:     t.Description,
		Attachments:     TransactionAttachmentsToResponse(t.Attachments),
		SubTransactions: TransactionSubTransactionsToResponse(t.SubTransactions),
	}
}
