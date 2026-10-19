package transaction_sub_transaction

import (
	"time"

	"github.com/google/uuid"

	"jww/internal/domain"
	"jww/internal/service/unit_of_work"
)

// Service owns single-row CRUD for transaction line items. It deliberately has
// no amount-accumulation logic: recomputing the parent Transaction.Amount is a
// cross-entity concern owned by the transaction workflow, which composes this
// service with the transaction repository inside one unit of work.
type Service struct {
	repository Repository
	uowManager unit_of_work.Manager
}

func NewService(repository Repository, uowManager unit_of_work.Manager) *Service {
	return &Service{repository: repository, uowManager: uowManager}
}

// Create builds and validates the row, assigning its identity and creation
// timestamp, then persists it. A validation failure returns before the
// repository is touched.
func (s *Service) Create(
	ctx domain.AppContext,
	transactionID int64,
	name string,
	label string,
	amount domain.Money,
	uow ...unit_of_work.UnitOfWork,
) (*domain.TransactionSubTransaction, error) {
	sub := domain.TransactionSubTransaction{
		ID:            uuid.New(),
		TransactionID: transactionID,
		Name:          name,
		Label:         label,
		Amount:        amount,
		CreatedAt:     time.Now().UTC(),
	}
	if err := sub.Validate(); err != nil {
		return nil, err
	}
	if err := s.repository.Create(ctx, sub, uow...); err != nil {
		return nil, err
	}
	return &sub, nil
}

// GetByID is an identity-only lookup and translates a repository (nil, nil)
// into the not-found domain error, mirroring transaction_attachment.GetDraft.
func (s *Service) GetByID(ctx domain.AppContext, id uuid.UUID, uow ...unit_of_work.UnitOfWork) (*domain.TransactionSubTransaction, error) {
	sub, err := s.repository.GetByID(ctx, id, uow...)
	if err != nil {
		return nil, err
	}
	if sub == nil {
		return nil, domain.NewTransactionSubTransactionNotFoundError(id)
	}
	return sub, nil
}

// Update validates first — before any I/O, so an invalid update never reaches
// the repository or opens a unit of work — then owns a unit of work unless the
// caller supplied one.
func (s *Service) Update(ctx domain.AppContext, sub domain.TransactionSubTransaction, callerUoW ...unit_of_work.UnitOfWork) (*domain.TransactionSubTransaction, error) {
	if err := sub.Validate(); err != nil {
		return nil, err
	}
	if len(callerUoW) > 0 {
		return s.update(ctx, sub, callerUoW...)
	}

	uow, err := s.uowManager.Create(ctx)
	if err != nil {
		return nil, err
	}
	updated, err := s.update(ctx, sub, uow)
	if err != nil {
		_ = uow.Restore(ctx)
		return nil, err
	}
	if err := uow.Finish(ctx); err != nil {
		return nil, err
	}
	return updated, nil
}

// update participates in a caller-owned unit of work. The owner is responsible
// for finishing or restoring it.
func (s *Service) update(ctx domain.AppContext, sub domain.TransactionSubTransaction, uow ...unit_of_work.UnitOfWork) (*domain.TransactionSubTransaction, error) {
	existing, err := s.repository.GetByID(ctx, sub.ID, uow...)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, domain.NewTransactionSubTransactionNotFoundError(sub.ID)
	}
	if err := s.repository.Update(ctx, sub, uow...); err != nil {
		return nil, err
	}
	return &sub, nil
}

// Delete owns a unit of work unless the caller supplied one. A missing row is a
// not-found, not a silent no-op.
func (s *Service) Delete(ctx domain.AppContext, sub domain.TransactionSubTransaction, callerUoW ...unit_of_work.UnitOfWork) error {
	if len(callerUoW) > 0 {
		return s.delete(ctx, sub, callerUoW...)
	}

	uow, err := s.uowManager.Create(ctx)
	if err != nil {
		return err
	}
	if err := s.delete(ctx, sub, uow); err != nil {
		_ = uow.Restore(ctx)
		return err
	}
	return uow.Finish(ctx)
}

// delete participates in a caller-owned unit of work. The owner is responsible
// for finishing or restoring it.
func (s *Service) delete(ctx domain.AppContext, sub domain.TransactionSubTransaction, uow ...unit_of_work.UnitOfWork) error {
	existing, err := s.repository.GetByID(ctx, sub.ID, uow...)
	if err != nil {
		return err
	}
	if existing == nil {
		return domain.NewTransactionSubTransactionNotFoundError(sub.ID)
	}
	return s.repository.Delete(ctx, sub, uow...)
}

// ListByTransactionIDs hydrates line items for a set of parents. An empty id set
// short-circuits without a repository call, mirroring
// transaction_attachment.Service.ListByTransactionIDs.
func (s *Service) ListByTransactionIDs(ctx domain.AppContext, transactionIDs []int64, uow ...unit_of_work.UnitOfWork) ([]domain.TransactionSubTransaction, error) {
	if len(transactionIDs) == 0 {
		return nil, nil
	}
	return s.repository.ListByTransactionIDs(ctx, transactionIDs, uow...)
}
