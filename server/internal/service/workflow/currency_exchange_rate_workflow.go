// Package workflow composes several internal/service packages behind one
// operation — resolving referenced resources and/or demarcating a
// transaction across more than one service call. It is the only layer
// permitted to depend on sibling service packages.
package workflow

import (
	"fmt"
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
	rateSource          currency_exchange_rate.RateSource
	uowManager          unit_of_work.Manager
}

func NewCurrencyExchangeRateWorkflow(
	currencyService *currency.Service,
	exchangeRateService *currency_exchange_rate.Service,
	rateSource currency_exchange_rate.RateSource,
	uowManager unit_of_work.Manager,
) *CurrencyExchangeRateWorkflow {
	return &CurrencyExchangeRateWorkflow{
		currencyService:     currencyService,
		exchangeRateService: exchangeRateService,
		rateSource:          rateSource,
		uowManager:          uowManager,
	}
}

// SyncResult summarizes one sync run for the boundary to log: what was fetched,
// what actually changed, and every skip/drop that must not be swallowed.
type SyncResult struct {
	Currencies int
	Fetched    int
	Inserted   int
	Updated    int
	Skipped    int
	Warnings   []string
}

// SyncDaily records the previous calendar day's rates for every directed pair
// of the currencies the app holds. It is the scheduled entry point: the tick
// fires just after midnight, when the previous day is fully published.
func (w *CurrencyExchangeRateWorkflow) SyncDaily(ctx domain.AppContext) (SyncResult, error) {
	return w.Sync(ctx, time.Now().UTC().AddDate(0, 0, -1))
}

// Sync records a specific day's rates for the full directed matrix among every
// currency the app holds (N×(N−1) rows). It is base-agnostic — the base flag is
// never consulted, because the base can change and every pair is fetched
// regardless. The whole matrix is written by a single UpsertMany call, so it
// needs no unit of work. Nothing is dropped silently: every provider skip and
// every unconfigured code is appended to Warnings. A zero asOf asks the
// provider for its latest day.
func (w *CurrencyExchangeRateWorkflow) Sync(ctx domain.AppContext, asOf time.Time) (SyncResult, error) {
	currencies, err := w.currencyService.List(ctx)
	if err != nil {
		return SyncResult{}, err
	}

	result := SyncResult{Currencies: len(currencies)}
	if len(currencies) < 2 {
		return result, nil // nothing to pair
	}

	codes := make([]string, 0, len(currencies))
	known := make(map[string]struct{}, len(currencies))
	for _, c := range currencies {
		codes = append(codes, c.Code)
		known[c.Code] = struct{}{}
	}

	fetch, err := w.rateSource.Rates(ctx, codes, asOf) // fetch before writing
	if err != nil {
		return SyncResult{}, err
	}
	result.Fetched = len(fetch.Rates)
	result.Skipped = len(fetch.Skipped)

	for _, skip := range fetch.Skipped {
		result.Warnings = append(result.Warnings, fmt.Sprintf("no rate for %s→%s on %s: %s",
			skip.From, skip.To, asOf.Format("2006-01-02"), skip.Reason))
	}
	if len(fetch.Rates) == 0 {
		result.Warnings = append(result.Warnings,
			fmt.Sprintf("provider returned no rates for %s", asOf.Format("2006-01-02")))
	}

	// Keep only pairs whose codes are currencies we actually hold (FK guard).
	rates := make([]domain.CurrencyExchangeRateFlatten, 0, len(fetch.Rates))
	for _, rate := range fetch.Rates {
		_, okFrom := known[rate.FromCurrencyCode]
		_, okTo := known[rate.ToCurrencyCode]
		if !okFrom || !okTo {
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("dropped rate %s→%s: currency not configured", rate.FromCurrencyCode, rate.ToCurrencyCode))
			continue
		}
		rates = append(rates, rate)
	}

	// A single write needs no unit of work: UpsertMany is one statement.
	counts, err := w.exchangeRateService.UpsertMany(ctx, rates)
	if err != nil {
		return SyncResult{}, err
	}

	result.Inserted = counts.Inserted
	result.Updated = counts.Updated
	return result, nil
}

func (w *CurrencyExchangeRateWorkflow) Create(ctx domain.AppContext, fromCode, toCode string, rate domain.ExchangeRate) (*domain.CurrencyExchangeRate, error) {
	uow, err := w.uowManager.Create(ctx)
	if err != nil {
		return nil, err
	}
	from, err := w.currencyService.GetByCode(ctx, fromCode, uow)
	if err != nil {
		_ = uow.Restore(ctx)
		return nil, err
	}
	to, err := w.currencyService.GetByCode(ctx, toCode, uow)
	if err != nil {
		_ = uow.Restore(ctx)
		return nil, err
	}
	exchangeRate, err := w.exchangeRateService.Create(ctx, *from, *to, rate, uow)
	if err != nil {
		_ = uow.Restore(ctx)
		return nil, err
	}
	if err := uow.Finish(ctx); err != nil {
		return nil, err
	}
	return exchangeRate, nil
}

func (w *CurrencyExchangeRateWorkflow) GetByCurrencies(ctx domain.AppContext, fromCode, toCode string, asOf time.Time) (*domain.CurrencyExchangeRate, error) {
	uow, err := w.uowManager.Create(ctx)
	if err != nil {
		return nil, err
	}
	from, err := w.currencyService.GetByCode(ctx, fromCode, uow)
	if err != nil {
		_ = uow.Restore(ctx)
		return nil, err
	}
	to, err := w.currencyService.GetByCode(ctx, toCode, uow)
	if err != nil {
		_ = uow.Restore(ctx)
		return nil, err
	}
	exchangeRate, err := w.exchangeRateService.GetByCurrencies(ctx, *from, *to, asOf, uow)
	if err != nil {
		_ = uow.Restore(ctx)
		return nil, err
	}
	if err := uow.Finish(ctx); err != nil {
		return nil, err
	}
	return exchangeRate, nil
}

func (w *CurrencyExchangeRateWorkflow) List(ctx domain.AppContext, filter currency_exchange_rate.Filter) ([]domain.CurrencyExchangeRate, error) {
	return w.exchangeRateService.List(ctx, &filter)
}
