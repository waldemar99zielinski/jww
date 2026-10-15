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
	return firstError(
		exactLength("currency_code", r.CurrencyCode, 3),
		minLength("name", r.Name, 1),
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
	}
}
