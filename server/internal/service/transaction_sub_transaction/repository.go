package transaction_sub_transaction

import (
	"github.com/google/uuid"

	"jww/internal/domain"
	"jww/internal/service/unit_of_work"
)

// Repository is the line-item persistence port. Every method takes an optional
// unit of work last and variadic, so a caller — normally the transaction
// workflow — can compose a line-item write with its parent transaction mutation
// atomically. GetByID is an identity-only lookup.
type Repository interface {
	Create(ctx domain.AppContext, sub domain.TransactionSubTransaction, uow ...unit_of_work.UnitOfWork) error
	Update(ctx domain.AppContext, sub domain.TransactionSubTransaction, uow ...unit_of_work.UnitOfWork) error
	Delete(ctx domain.AppContext, sub domain.TransactionSubTransaction, uow ...unit_of_work.UnitOfWork) error
	GetByID(ctx domain.AppContext, id uuid.UUID, uow ...unit_of_work.UnitOfWork) (*domain.TransactionSubTransaction, error)
	ListByTransactionIDs(ctx domain.AppContext, transactionIDs []int64, uow ...unit_of_work.UnitOfWork) ([]domain.TransactionSubTransaction, error)
}
