package currency_exchange_rate

import (
	"time"

	"jww/internal/domain"
	"jww/internal/service/unit_of_work"
)

type Repository interface {
	Create(ctx domain.AppContext, exchangeRate domain.CurrencyExchangeRate, uow ...unit_of_work.UnitOfWork) error
	// UpsertMany writes a whole batch in one statement, replacing rate on a
	// (from_currency_code, to_currency_code, recorded_at) conflict.
	UpsertMany(ctx domain.AppContext, rates []domain.CurrencyExchangeRateFlatten, uow ...unit_of_work.UnitOfWork) (UpsertCounts, error)
	GetByCurrencies(ctx domain.AppContext, from, to domain.Currency, asOf time.Time, uow ...unit_of_work.UnitOfWork) (*domain.CurrencyExchangeRate, error)
	List(ctx domain.AppContext, filter *Filter, uow ...unit_of_work.UnitOfWork) ([]domain.CurrencyExchangeRate, error)
}
