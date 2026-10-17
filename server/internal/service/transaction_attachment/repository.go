package transaction_attachment

import (
	"github.com/google/uuid"

	"jww/internal/domain"
	"jww/internal/service/unit_of_work"
)

// Repository is the metadata port. GetByID is an identity-only lookup; Find
// applies explicit access constraints such as draft-only or parent-scoped.
type Repository interface {
	Create(ctx domain.AppContext, attachment domain.TransactionAttachment, uow ...unit_of_work.UnitOfWork) error
	GetByID(ctx domain.AppContext, attachmentID uuid.UUID, uow ...unit_of_work.UnitOfWork) (*domain.TransactionAttachment, error)
	Find(ctx domain.AppContext, filter Filter, uow ...unit_of_work.UnitOfWork) (*domain.TransactionAttachment, error)
	LockByIDs(ctx domain.AppContext, attachmentIDs []uuid.UUID, uow ...unit_of_work.UnitOfWork) ([]domain.TransactionAttachment, error)
	LinkToTransaction(ctx domain.AppContext, transactionID int64, attachmentIDs []uuid.UUID, uow ...unit_of_work.UnitOfWork) error
	ListByTransactionIDs(ctx domain.AppContext, transactionIDs []int64, uow ...unit_of_work.UnitOfWork) ([]domain.TransactionAttachment, error)
	Delete(ctx domain.AppContext, attachment domain.TransactionAttachment, uow ...unit_of_work.UnitOfWork) error
}
