// Package transaction_category implements CRUD use cases for transaction
// categories, keyed by uuid.UUID.
package transaction_category

import (
	"github.com/google/uuid"

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

func (s *Service) Create(ctx domain.AppContext, categoryType domain.TransactionCategoryType, name string, sortOrder int, uow ...unit_of_work.UnitOfWork) (*domain.TransactionCategory, error) {
	category := domain.TransactionCategory{
		ID:        uuid.New(),
		Type:      categoryType,
		Name:      name,
		Active:    true,
		SortOrder: sortOrder,
	}
	if err := category.Validate(); err != nil {
		return nil, err
	}
	if err := s.repository.Create(ctx, category, uow...); err != nil {
		return nil, err
	}
	return &category, nil
}

func (s *Service) GetByID(ctx domain.AppContext, categoryID uuid.UUID, uow ...unit_of_work.UnitOfWork) (*domain.TransactionCategory, error) {
	category, err := s.repository.GetByID(ctx, categoryID, uow...)
	if err != nil {
		return nil, err
	}
	if category == nil {
		return nil, domain.NewTransactionCategoryNotFoundError(categoryID)
	}
	return category, nil
}

func (s *Service) ListByType(ctx domain.AppContext, categoryType domain.TransactionCategoryType, activeOnly bool, uow ...unit_of_work.UnitOfWork) ([]domain.TransactionCategory, error) {
	return s.repository.ListByType(ctx, categoryType, activeOnly, uow...)
}

func (s *Service) Update(ctx domain.AppContext, category domain.TransactionCategory) (*domain.TransactionCategory, error) {
	if err := category.Validate(); err != nil {
		return nil, err
	}
	uow, err := s.uowManager.Create(ctx)
	if err != nil {
		return nil, err
	}
	existing, err := s.repository.GetByID(ctx, category.ID, uow)
	if err != nil {
		_ = uow.Restore(ctx)
		return nil, err
	}
	if existing == nil {
		_ = uow.Restore(ctx)
		return nil, domain.NewTransactionCategoryNotFoundError(category.ID)
	}
	if err := s.repository.Update(ctx, category, uow); err != nil {
		_ = uow.Restore(ctx)
		return nil, err
	}
	if err := uow.Finish(ctx); err != nil {
		return nil, err
	}
	return &category, nil
}

func (s *Service) Delete(ctx domain.AppContext, category domain.TransactionCategory) error {
	uow, err := s.uowManager.Create(ctx)
	if err != nil {
		return err
	}
	existing, err := s.repository.GetByID(ctx, category.ID, uow)
	if err != nil {
		_ = uow.Restore(ctx)
		return err
	}
	if existing == nil {
		_ = uow.Restore(ctx)
		return domain.NewTransactionCategoryNotFoundError(category.ID)
	}
	if err := s.repository.Delete(ctx, category, uow); err != nil {
		_ = uow.Restore(ctx)
		return err
	}
	return uow.Finish(ctx)
}
