package workflow

import (
	"time"

	"github.com/google/uuid"

	"jww/internal/domain"
	"jww/internal/service/currency"
	"jww/internal/service/transaction"
	"jww/internal/service/transaction_category"
	"jww/internal/service/unit_of_work"
	"jww/internal/service/user"
)

// TransactionWorkflow resolves the currency/category/user a transaction
// references before handing off to transaction.Service — the orchestration
// that create_transaction_handler.go and list_transactions_handler.go used
// to do inline, composing four services directly in the HTTP layer.
type TransactionWorkflow struct {
	currencyService    *currency.Service
	categoryService    *transaction_category.Service
	userService        *user.Service
	transactionService *transaction.Service
	uowManager         unit_of_work.Manager
}

func NewTransactionWorkflow(
	currencyService *currency.Service,
	categoryService *transaction_category.Service,
	userService *user.Service,
	transactionService *transaction.Service,
	uowManager unit_of_work.Manager,
) *TransactionWorkflow {
	return &TransactionWorkflow{
		currencyService:    currencyService,
		categoryService:    categoryService,
		userService:        userService,
		transactionService: transactionService,
		uowManager:         uowManager,
	}
}

func (w *TransactionWorkflow) Create(
	ctx domain.AppContext,
	transactionType domain.TransactionType,
	amount domain.Money,
	currencyCode string,
	categoryID uuid.UUID,
	userID uuid.UUID,
	source domain.TransactionSource,
	name string,
	description string,
) (*domain.Transaction, error) {
	active, err := w.uowManager.Create(ctx)
	if err != nil {
		return nil, err
	}
	curr, err := w.currencyService.GetByCode(ctx, currencyCode, active)
	if err != nil {
		_ = active.Restore(ctx)
		return nil, err
	}
	category, err := w.categoryService.GetByID(ctx, categoryID, active)
	if err != nil {
		_ = active.Restore(ctx)
		return nil, err
	}
	u, err := w.userService.GetByID(ctx, userID, active)
	if err != nil {
		_ = active.Restore(ctx)
		return nil, err
	}
	created, err := w.transactionService.Create(ctx, transactionType, amount, *curr, *category, *u, source, name, description, active)
	if err != nil {
		_ = active.Restore(ctx)
		return nil, err
	}
	if err := active.Finish(ctx); err != nil {
		return nil, err
	}
	return created, nil
}

// TransactionListFilter mirrors transaction.Filter but references
// resources by the raw identifiers a request carries (currency code,
// category/user IDs) rather than already-resolved domain structs.
type TransactionListFilter struct {
	UserID       *uuid.UUID
	CategoryID   *uuid.UUID
	Type         *domain.TransactionType
	CurrencyCode string
	Source       *domain.TransactionSource
	CreatedFrom  *time.Time
	CreatedTo    *time.Time
}

func (w *TransactionWorkflow) List(ctx domain.AppContext, f TransactionListFilter) ([]domain.Transaction, error) {
	active, err := w.uowManager.Create(ctx)
	if err != nil {
		return nil, err
	}
	filter := transaction.Filter{
		Type:        f.Type,
		Source:      f.Source,
		CreatedFrom: f.CreatedFrom,
		CreatedTo:   f.CreatedTo,
	}
	if f.UserID != nil {
		u, err := w.userService.GetByID(ctx, *f.UserID, active)
		if err != nil {
			_ = active.Restore(ctx)
			return nil, err
		}
		filter.User = u
	}
	if f.CategoryID != nil {
		c, err := w.categoryService.GetByID(ctx, *f.CategoryID, active)
		if err != nil {
			_ = active.Restore(ctx)
			return nil, err
		}
		filter.Category = c
	}
	if f.CurrencyCode != "" {
		c, err := w.currencyService.GetByCode(ctx, f.CurrencyCode, active)
		if err != nil {
			_ = active.Restore(ctx)
			return nil, err
		}
		filter.Currency = c
	}
	if err := filter.Validate(); err != nil {
		_ = active.Restore(ctx)
		return nil, err
	}
	transactions, err := w.transactionService.List(ctx, &filter, active)
	if err != nil {
		_ = active.Restore(ctx)
		return nil, err
	}
	if err := active.Finish(ctx); err != nil {
		return nil, err
	}
	return transactions, nil
}
