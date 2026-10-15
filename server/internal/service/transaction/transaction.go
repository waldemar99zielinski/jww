// Package transaction implements CRUD/list use cases for transactions,
// keyed by an int64 identity, plus the reusable Filter query object.
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
	created, err := s.repository.Create(ctx, transactionType, amount, currency, category, user, source, time.Now().UTC(), name, description, uow...)
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
	return s.repository.List(ctx, filter, uow...)
}

func (s *Service) Update(ctx domain.AppContext, transaction domain.Transaction) (*domain.Transaction, error) {
	if err := transaction.Validate(); err != nil {
		return nil, err
	}
	active, err := s.uowManager.Create(ctx)
	if err != nil {
		return nil, err
	}
	existing, err := s.repository.GetByID(ctx, transaction.ID, active)
	if err != nil {
		_ = active.Restore(ctx)
		return nil, err
	}
	if existing == nil {
		_ = active.Restore(ctx)
		return nil, domain.NewTransactionNotFoundError(transaction.ID)
	}
	if err := s.repository.Update(ctx, transaction, active); err != nil {
		_ = active.Restore(ctx)
		return nil, err
	}
	if err := active.Finish(ctx); err != nil {
		return nil, err
	}
	return &transaction, nil
}

func (s *Service) Delete(ctx domain.AppContext, transaction domain.Transaction) error {
	active, err := s.uowManager.Create(ctx)
	if err != nil {
		return err
	}
	existing, err := s.repository.GetByID(ctx, transaction.ID, active)
	if err != nil {
		_ = active.Restore(ctx)
		return err
	}
	if existing == nil {
		_ = active.Restore(ctx)
		return domain.NewTransactionNotFoundError(transaction.ID)
	}
	if err := s.repository.Delete(ctx, transaction, active); err != nil {
		_ = active.Restore(ctx)
		return err
	}
	return active.Finish(ctx)
}
