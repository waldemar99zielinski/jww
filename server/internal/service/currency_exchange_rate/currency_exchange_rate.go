// Package currency_exchange_rate implements create/read use cases for
// exchange-rate records, plus the batched write path the scheduled sync uses.
// It owns no unit of work manager: every method is a single repository call,
// so callers supply the transaction. It also declares the outbound RateSource
// port the workflow depends on.
package currency_exchange_rate

import (
	"time"

	"jww/internal/domain"
	"jww/internal/service/unit_of_work"
)

// RateSkip is a directed pair the source could not produce a rate for, with a
// human-readable reason. It is returned (never dropped) so the boundary can
// log it.
type RateSkip struct {
	From   string
	To     string
	Reason string
}

// RateFetch is what a RateSource returns for one day: the directed pairs it
// read and the pairs it skipped.
type RateFetch struct {
	Rates   []domain.CurrencyExchangeRateFlatten
	Skipped []RateSkip
}

// RateSource is the outbound port for fetching exchange rates among a set of
// currencies for one day. Implementations return every directed pair they
// could read (N×(N−1) rows) plus a RateSkip for anything the provider omitted.
// A zero asOf asks for the provider's latest day. Implemented in
// internal/adapter/frankfurter.
type RateSource interface {
	Rates(ctx domain.AppContext, currencies []string, asOf time.Time) (RateFetch, error)
}

// UpsertCounts reports how a batch changed the table: rows newly inserted vs.
// rows whose rate was replaced on a (from, to, recorded_at) conflict.
type UpsertCounts struct {
	Inserted int
	Updated  int
}

type Repository interface {
	Create(ctx domain.AppContext, exchangeRate domain.CurrencyExchangeRate, uow ...unit_of_work.UnitOfWork) error
	// UpsertMany writes a whole batch in one statement, replacing rate on a
	// (from_currency_code, to_currency_code, recorded_at) conflict.
	UpsertMany(ctx domain.AppContext, rates []domain.CurrencyExchangeRateFlatten, uow ...unit_of_work.UnitOfWork) (UpsertCounts, error)
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

func (s *Service) UpsertMany(ctx domain.AppContext, rates []domain.CurrencyExchangeRateFlatten, uow ...unit_of_work.UnitOfWork) (UpsertCounts, error) {
	if len(rates) == 0 {
		return UpsertCounts{}, nil
	}
	normalized := make([]domain.CurrencyExchangeRateFlatten, 0, len(rates))
	for _, rate := range rates {
		rate.RecordedAt = rate.RecordedAt.UTC()
		if err := rate.Validate(); err != nil {
			return UpsertCounts{}, err
		}
		normalized = append(normalized, rate)
	}
	return s.repository.UpsertMany(ctx, normalized, uow...)
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
