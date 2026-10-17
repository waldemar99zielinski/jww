package transaction

import (
	"time"

	"jww/internal/domain"
	"jww/internal/service/unit_of_work"
)

type Repository interface {
	Create(
		ctx domain.AppContext,
		transactionType domain.TransactionType,
		amount domain.Money,
		currency domain.Currency,
		category domain.TransactionCategory,
		user domain.User,
		source domain.TransactionSource,
		createdAt time.Time,
		name string,
		description string,
		uow ...unit_of_work.UnitOfWork,
	) (*domain.Transaction, error)
	GetByID(ctx domain.AppContext, transactionID int64, uow ...unit_of_work.UnitOfWork) (*domain.Transaction, error)
	List(ctx domain.AppContext, filter *Filter, uow ...unit_of_work.UnitOfWork) ([]domain.Transaction, error)
	Update(ctx domain.AppContext, transaction domain.Transaction, uow ...unit_of_work.UnitOfWork) error
	Delete(ctx domain.AppContext, transaction domain.Transaction, uow ...unit_of_work.UnitOfWork) error
}
