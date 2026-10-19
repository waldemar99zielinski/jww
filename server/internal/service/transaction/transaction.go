// Package transaction implements CRUD/list use cases for transactions,
// keyed by an int64 identity, plus the reusable Filter query object.
package transaction

import (
	"time"

	"jww/internal/domain"
	"jww/internal/service/unit_of_work"
)

type Service struct {
	repository Repository
	uowManager unit_of_work.Manager
}

func NewService(repository Repository, uowManager unit_of_work.Manager) *Service {
	return &Service{repository: repository, uowManager: uowManager}
}

func (s *Service) Create(
	ctx domain.AppContext,
	transactionType domain.TransactionType,
	amount domain.Money,
	currency domain.Currency,
	category domain.TransactionCategory,
	user domain.User,
	source domain.TransactionSource,
	name string,
	description string,
	createdAt *time.Time,
	uow ...unit_of_work.UnitOfWork,
) (*domain.Transaction, error) {
	tr := domain.Transaction{
		Type:        transactionType,
		Amount:      amount,
		Currency:    currency,
		Category:    category,
		User:        user,
		Source:      source,
		Name:        name,
		Description: description,
	}
	if err := tr.Validate(); err != nil {
		return nil, err
	}
	// The transaction's date defaults to now; a caller may supply it explicitly.
	when := time.Now().UTC()
	if createdAt != nil {
		when = createdAt.UTC()
	}
	created, err := s.repository.Create(ctx, transactionType, amount, currency, category, user, source, when, name, description, uow...)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func (s *Service) GetByID(ctx domain.AppContext, transactionID int64, uow ...unit_of_work.UnitOfWork) (*domain.Transaction, error) {
	transaction, err := s.repository.GetByID(ctx, transactionID, uow...)
	if err != nil {
		return nil, err
	}
	if transaction == nil {
		return nil, domain.NewTransactionNotFoundError(transactionID)
	}
	return transaction, nil
}

func (s *Service) List(ctx domain.AppContext, filter *Filter, uow ...unit_of_work.UnitOfWork) ([]domain.Transaction, error) {
	if filter != nil {
		if err := filter.Validate(); err != nil {
			return nil, err
		}
	}
	return s.repository.List(ctx, filter, uow...)
}

// Update validates the entity, checks that the row still exists, and persists
// it. It owns a unit of work unless the caller supplied one; with a supplied
// UoW it participates in it and leaves finish/restore to the owner.
func (s *Service) Update(ctx domain.AppContext, transaction domain.Transaction, callerUoW ...unit_of_work.UnitOfWork) (*domain.Transaction, error) {
	if err := transaction.Validate(); err != nil {
		return nil, err
	}
	if len(callerUoW) > 0 {
		return s.update(ctx, transaction, callerUoW...)
	}

	uow, err := s.uowManager.Create(ctx)
	if err != nil {
		return nil, err
	}
	updated, err := s.update(ctx, transaction, uow)
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
func (s *Service) update(ctx domain.AppContext, transaction domain.Transaction, uow ...unit_of_work.UnitOfWork) (*domain.Transaction, error) {
	existing, err := s.repository.GetByID(ctx, transaction.ID, uow...)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, domain.NewTransactionNotFoundError(transaction.ID)
	}
	if err := s.repository.Update(ctx, transaction, uow...); err != nil {
		return nil, err
	}
	return &transaction, nil
}

func (s *Service) Delete(ctx domain.AppContext, transaction domain.Transaction, callerUoW ...unit_of_work.UnitOfWork) error {
	if len(callerUoW) > 0 {
		return s.delete(ctx, transaction, callerUoW...)
	}

	uow, err := s.uowManager.Create(ctx)
	if err != nil {
		return err
	}
	if err := s.delete(ctx, transaction, uow); err != nil {
		_ = uow.Restore(ctx)
		return err
	}
	return uow.Finish(ctx)
}

// delete participates in a caller-owned unit of work. The owner is responsible
// for finishing or restoring it.
func (s *Service) delete(ctx domain.AppContext, transaction domain.Transaction, uow ...unit_of_work.UnitOfWork) error {
	existing, err := s.repository.GetByID(ctx, transaction.ID, uow...)
	if err != nil {
		return err
	}
	if existing == nil {
		return domain.NewTransactionNotFoundError(transaction.ID)
	}
	return s.repository.Delete(ctx, transaction, uow...)
}
