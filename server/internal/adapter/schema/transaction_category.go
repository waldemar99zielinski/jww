package schema

import (
	"fmt"

	"github.com/google/uuid"

	"jww/internal/domain"
)

type TransactionCategoryCreateRequest struct {
	Type      domain.TransactionCategoryType `json:"type"`
	Name      string                         `json:"name"`
	SortOrder int                            `json:"sort_order"`
}

func (r TransactionCategoryCreateRequest) Validate() error {
	if !r.Type.Valid() {
		return fmt.Errorf("type must be 'income' or 'expense'")
	}
	return firstError(
		minLength("name", r.Name, 1),
		nonNegative("sort_order", r.SortOrder),
	)
}

type TransactionCategoryUpdateRequest struct {
	Name      string `json:"name"`
	Active    bool   `json:"active"`
	SortOrder int    `json:"sort_order"`
}

func (r TransactionCategoryUpdateRequest) Validate() error {
	return firstError(
		minLength("name", r.Name, 1),
		nonNegative("sort_order", r.SortOrder),
	)
}

type TransactionCategoryResponse struct {
	ID        uuid.UUID                      `json:"id"`
	Type      domain.TransactionCategoryType `json:"type"`
	Name      string                         `json:"name"`
	Active    bool                           `json:"active"`
	SortOrder int                            `json:"sort_order"`
}

func TransactionCategoryToResponse(c domain.TransactionCategory) TransactionCategoryResponse {
	return TransactionCategoryResponse{
		ID:        c.ID,
		Type:      c.Type,
		Name:      c.Name,
		Active:    c.Active,
		SortOrder: c.SortOrder,
	}
}
