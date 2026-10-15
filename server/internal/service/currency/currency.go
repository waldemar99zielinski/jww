// Package currency implements CRUD use cases for the Currency entity, keyed
// by its natural 3-letter code rather than a surrogate id.
package currency

import (
	"jww/internal/domain"
	"jww/internal/service/unit_of_work"
)

type Repository interface {
	Create(ctx domain.AppContext, currency domain.Currency, uow ...unit_of_work.UnitOfWork) error
	GetByCode(ctx domain.AppContext, code string, uow ...unit_of_work.UnitOfWork) (*domain.Currency, error)
	List(ctx domain.AppContext, uow ...unit_of_work.UnitOfWork) ([]domain.Currency, error)
	Update(ctx domain.AppContext, currency domain.Currency, uow ...unit_of_work.UnitOfWork) error
	Delete(ctx domain.AppContext, currency domain.Currency, uow ...unit_of_work.UnitOfWork) error
}

type Service struct {
	repository Repository
	uowManager unit_of_work.Manager
}

func NewService(repository Repository, uowManager unit_of_work.Manager) *Service {
	return &Service{repository: repository, uowManager: uowManager}
}

func (s *Service) Create(ctx domain.AppContext, code, name string, isBase bool, uow ...unit_of_work.UnitOfWork) (*domain.Currency, error) {
	currency := domain.Currency{Code: code, Name: name, IsBase: isBase}
	if err := currency.Validate(); err != nil {
		return nil, err
	}
	if err := s.repository.Create(ctx, currency, uow...); err != nil {
		return nil, err
	}
	return &currency, nil
}

func (s *Service) GetByCode(ctx domain.AppContext, code string, uow ...unit_of_work.UnitOfWork) (*domain.Currency, error) {
	currency, err := s.repository.GetByCode(ctx, code, uow...)
	if err != nil {
		return nil, err
	}
	if currency == nil {
		return nil, domain.NewCurrencyNotFoundError(code)
	}
	return currency, nil
}

func (s *Service) List(ctx domain.AppContext, uow ...unit_of_work.UnitOfWork) ([]domain.Currency, error) {
	return s.repository.List(ctx, uow...)
}

func (s *Service) Update(ctx domain.AppContext, currency domain.Currency) (*domain.Currency, error) {
	if err := currency.Validate(); err != nil {
		return nil, err
	}
	active, err := s.uowManager.Create(ctx)
	if err != nil {
		return nil, err
	}
	existing, err := s.repository.GetByCode(ctx, currency.Code, active)
	if err != nil {
		_ = active.Restore(ctx)
		return nil, err
	}
	if existing == nil {
		_ = active.Restore(ctx)
		return nil, domain.NewCurrencyNotFoundError(currency.Code)
	}
	if err := s.repository.Update(ctx, currency, active); err != nil {
		_ = active.Restore(ctx)
		return nil, err
	}
	if err := active.Finish(ctx); err != nil {
		return nil, err
	}
	return &currency, nil
}

func (s *Service) Delete(ctx domain.AppContext, currency domain.Currency) error {
	active, err := s.uowManager.Create(ctx)
	if err != nil {
		return err
	}
	existing, err := s.repository.GetByCode(ctx, currency.Code, active)
	if err != nil {
		_ = active.Restore(ctx)
		return err
	}
	if existing == nil {
		_ = active.Restore(ctx)
		return domain.NewCurrencyNotFoundError(currency.Code)
	}
	if err := s.repository.Delete(ctx, currency, active); err != nil {
		_ = active.Restore(ctx)
		return err
	}
	return active.Finish(ctx)
}
