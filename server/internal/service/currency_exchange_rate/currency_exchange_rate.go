// Package currency_exchange_rate implements create/read use cases for
// exchange-rate records. It owns no unit of work manager: every method is a
// single repository call, so callers supply the transaction.
package currency_exchange_rate

import (
	"time"

	"jww/internal/domain"
	"jww/internal/service/unit_of_work"
)

type Repository interface {
	Create(ctx domain.AppContext, exchangeRate domain.CurrencyExchangeRate, uow ...unit_of_work.UnitOfWork) error
	GetByCurrencies(ctx domain.AppContext, from, to domain.Currency, asOf time.Time, uow ...unit_of_work.UnitOfWork) (*domain.CurrencyExchangeRate, error)
	List(ctx domain.AppContext, uow ...unit_of_work.UnitOfWork) ([]domain.CurrencyExchangeRate, error)
}

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) Create(ctx domain.AppContext, from, to domain.Currency, rate domain.ExchangeRate, uow ...unit_of_work.UnitOfWork) (*domain.CurrencyExchangeRate, error) {
	exchangeRate := domain.CurrencyExchangeRate{
		FromCurrency: from,
		ToCurrency:   to,
		Rate:         rate,
		RecordedAt:   time.Now().UTC(),
	}
	if err := exchangeRate.Validate(); err != nil {
		return nil, err
	}
	if err := s.repository.Create(ctx, exchangeRate, uow...); err != nil {
		return nil, err
	}
	return &exchangeRate, nil
}

func (s *Service) GetByCurrencies(ctx domain.AppContext, from, to domain.Currency, asOf time.Time, uow ...unit_of_work.UnitOfWork) (*domain.CurrencyExchangeRate, error) {
	exchangeRate, err := s.repository.GetByCurrencies(ctx, from, to, asOf, uow...)
	if err != nil {
		return nil, err
	}
	if exchangeRate == nil {
		return nil, domain.NewCurrencyExchangeRateNotFoundError(from.Code, to.Code)
	}
	return exchangeRate, nil
}

func (s *Service) List(ctx domain.AppContext, uow ...unit_of_work.UnitOfWork) ([]domain.CurrencyExchangeRate, error) {
	return s.repository.List(ctx, uow...)
}
