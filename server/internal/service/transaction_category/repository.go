package transaction_category

import (
	"github.com/google/uuid"

	"jww/internal/domain"
	"jww/internal/service/unit_of_work"
)

type Repository interface {
	Create(ctx domain.AppContext, category domain.TransactionCategory, uow ...unit_of_work.UnitOfWork) error
	GetByID(ctx domain.AppContext, categoryID uuid.UUID, uow ...unit_of_work.UnitOfWork) (*domain.TransactionCategory, error)
	ListByType(ctx domain.AppContext, categoryType domain.TransactionCategoryType, activeOnly bool, uow ...unit_of_work.UnitOfWork) ([]domain.TransactionCategory, error)
	Update(ctx domain.AppContext, category domain.TransactionCategory, uow ...unit_of_work.UnitOfWork) error
	Delete(ctx domain.AppContext, category domain.TransactionCategory, uow ...unit_of_work.UnitOfWork) error
}
