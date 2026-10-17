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
	// AttachmentIDs optionally links already-uploaded unattached drafts to
	// the new transaction, atomically with its creation. Absent/empty keeps
	// the request valid and creates a transaction with no attachments.
	AttachmentIDs []uuid.UUID `json:"attachment_ids"`
}

func (r TransactionCreateRequest) Validate() error {
	if !r.Type.Valid() {
		return fmt.Errorf("type must be 'income' or 'expense'")
	}
	if r.Amount <= 0 {
		return fmt.Errorf("amount must be greater than 0")
	}
	if !r.Source.Valid() {
		return fmt.Errorf("source must be 'manual' or 'recurring'")
	}
	if err := firstError(
		exactLength("currency_code", r.CurrencyCode, 3),
		minLength("name", r.Name, 1),
	); err != nil {
		return err
	}
	return firstError(
		noNilUUIDs("attachment_ids", r.AttachmentIDs),
		noDuplicateUUIDs("attachment_ids", r.AttachmentIDs),
	)
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
}

func TransactionToResponse(t domain.Transaction) TransactionResponse {
	return TransactionResponse{
		ID:           t.ID,
		Type:         t.Type,
		Amount:       t.Amount,
		CurrencyCode: t.Currency.Code,
		CategoryID:   t.Category.ID,
		CategoryName: t.Category.Name,
		UserID:       t.User.ID,
		Source:       t.Source,
		Name:         t.Name,
		CreatedAt:    t.CreatedAt,
		Description:  t.Description,
		Attachments:  TransactionAttachmentsToResponse(t.Attachments),
	}
}
