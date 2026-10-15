// Package workflow composes several internal/service packages behind one
// operation — resolving referenced resources and/or demarcating a
// transaction across more than one service call. It is the only layer
// permitted to depend on sibling service packages.
package workflow

import (
	"time"

	"jww/internal/domain"
	"jww/internal/service/currency"
	currency_exchange_rate "jww/internal/service/currency_exchange_rate"
	"jww/internal/service/unit_of_work"
)

// CurrencyExchangeRateWorkflow resolves the from/to currency codes an
// exchange rate request references before handing off to
// currency_exchange_rate.Service — orchestration that
// create_exchange_rate_handler.go and get_exchange_rate_handler.go used to
// do inline, composing two services directly in the HTTP layer. List is a
// plain passthrough, included here so infra always goes through one
// workflow type for this resource.
type CurrencyExchangeRateWorkflow struct {
	currencyService     *currency.Service
	exchangeRateService *currency_exchange_rate.Service
	uowManager          unit_of_work.Manager
}

func NewCurrencyExchangeRateWorkflow(
	currencyService *currency.Service,
	exchangeRateService *currency_exchange_rate.Service,
	uowManager unit_of_work.Manager,
) *CurrencyExchangeRateWorkflow {
	return &CurrencyExchangeRateWorkflow{
		currencyService:     currencyService,
		exchangeRateService: exchangeRateService,
		uowManager:          uowManager,
	}
}

func (w *CurrencyExchangeRateWorkflow) Create(ctx domain.AppContext, fromCode, toCode string, rate domain.ExchangeRate) (*domain.CurrencyExchangeRate, error) {
	active, err := w.uowManager.Create(ctx)
	if err != nil {
		return nil, err
	}
	from, err := w.currencyService.GetByCode(ctx, fromCode, active)
	if err != nil {
		_ = active.Restore(ctx)
		return nil, err
	}
	to, err := w.currencyService.GetByCode(ctx, toCode, active)
	if err != nil {
		_ = active.Restore(ctx)
		return nil, err
	}
	exchangeRate, err := w.exchangeRateService.Create(ctx, *from, *to, rate, active)
	if err != nil {
		_ = active.Restore(ctx)
		return nil, err
	}
	if err := active.Finish(ctx); err != nil {
		return nil, err
	}
	return exchangeRate, nil
}

func (w *CurrencyExchangeRateWorkflow) GetByCurrencies(ctx domain.AppContext, fromCode, toCode string, asOf time.Time) (*domain.CurrencyExchangeRate, error) {
	active, err := w.uowManager.Create(ctx)
	if err != nil {
		return nil, err
	}
	from, err := w.currencyService.GetByCode(ctx, fromCode, active)
	if err != nil {
		_ = active.Restore(ctx)
		return nil, err
	}
	to, err := w.currencyService.GetByCode(ctx, toCode, active)
	if err != nil {
		_ = active.Restore(ctx)
		return nil, err
	}
	exchangeRate, err := w.exchangeRateService.GetByCurrencies(ctx, *from, *to, asOf, active)
	if err != nil {
		_ = active.Restore(ctx)
		return nil, err
	}
	if err := active.Finish(ctx); err != nil {
		return nil, err
	}
	return exchangeRate, nil
}

func (w *CurrencyExchangeRateWorkflow) List(ctx domain.AppContext) ([]domain.CurrencyExchangeRate, error) {
	return w.exchangeRateService.List(ctx)
}
