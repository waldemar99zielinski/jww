package workflow_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"jww/internal/domain"
	"jww/internal/service/currency"
	currency_exchange_rate "jww/internal/service/currency_exchange_rate"
	"jww/internal/service/workflow"
)

var (
	usd = domain.Currency{Code: "USD", Name: "US Dollar", IsBase: true}
	eur = domain.Currency{Code: "EUR", Name: "Euro"}
)

func newExchangeRateWorkflow(currencyRepo *fakeCurrencyRepo, rateRepo *fakeExchangeRateRepo, uowManager *fakeUoWManager) *workflow.CurrencyExchangeRateWorkflow {
	currencyService := currency.NewService(currencyRepo, uowManager)
	rateService := currency_exchange_rate.NewService(rateRepo)
	return workflow.NewCurrencyExchangeRateWorkflow(currencyService, rateService, uowManager)
}

func TestCurrencyExchangeRateWorkflow_Create(t *testing.T) {
	t.Run("success commits once", func(t *testing.T) {
		currencyRepo := newFakeCurrencyRepo(usd, eur)
		rateRepo := newFakeExchangeRateRepo()
		uowManager := &fakeUoWManager{}
		w := newExchangeRateWorkflow(currencyRepo, rateRepo, uowManager)

		rate, err := w.Create(newAppContext(), "USD", "EUR", domain.ExchangeRate(9_000))

		require.NoError(t, err)
		assert.Equal(t, "USD", rate.FromCurrency.Code)
		assert.Equal(t, "EUR", rate.ToCurrency.Code)
		assert.Len(t, rateRepo.created, 1)
		assert.True(t, uowManager.last().finished)
		assert.False(t, uowManager.last().restored)
	})

	t.Run("unknown from-currency rolls back before touching the to-currency or the rate", func(t *testing.T) {
		currencyRepo := newFakeCurrencyRepo(eur) // USD deliberately missing
		rateRepo := newFakeExchangeRateRepo()
		uowManager := &fakeUoWManager{}
		w := newExchangeRateWorkflow(currencyRepo, rateRepo, uowManager)

		_, err := w.Create(newAppContext(), "USD", "EUR", domain.ExchangeRate(9_000))

		require.Error(t, err)
		var notFound *domain.CurrencyNotFoundError
		assert.ErrorAs(t, err, &notFound)
		assert.True(t, uowManager.last().restored)
		assert.False(t, uowManager.last().finished)
		assert.Empty(t, rateRepo.created, "the rate must not be created when a referenced currency is missing")
	})

	t.Run("unknown to-currency rolls back", func(t *testing.T) {
		currencyRepo := newFakeCurrencyRepo(usd) // EUR deliberately missing
		rateRepo := newFakeExchangeRateRepo()
		uowManager := &fakeUoWManager{}
		w := newExchangeRateWorkflow(currencyRepo, rateRepo, uowManager)

		_, err := w.Create(newAppContext(), "USD", "EUR", domain.ExchangeRate(9_000))

		require.Error(t, err)
		assert.True(t, uowManager.last().restored)
		assert.Empty(t, rateRepo.created)
	})

	t.Run("invalid rate rolls back", func(t *testing.T) {
		currencyRepo := newFakeCurrencyRepo(usd, eur)
		rateRepo := newFakeExchangeRateRepo()
		uowManager := &fakeUoWManager{}
		w := newExchangeRateWorkflow(currencyRepo, rateRepo, uowManager)

		_, err := w.Create(newAppContext(), "USD", "EUR", domain.ExchangeRate(0))

		require.Error(t, err)
		assert.True(t, uowManager.last().restored)
	})
}

func TestCurrencyExchangeRateWorkflow_GetByCurrencies(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		currencyRepo := newFakeCurrencyRepo(usd, eur)
		rateRepo := newFakeExchangeRateRepo(domain.CurrencyExchangeRate{
			FromCurrency: usd, ToCurrency: eur, Rate: domain.ExchangeRate(9_000), RecordedAt: time.Now().UTC(),
		})
		uowManager := &fakeUoWManager{}
		w := newExchangeRateWorkflow(currencyRepo, rateRepo, uowManager)

		rate, err := w.GetByCurrencies(newAppContext(), "USD", "EUR", time.Now())

		require.NoError(t, err)
		assert.Equal(t, domain.ExchangeRate(9_000), rate.Rate)
		assert.True(t, uowManager.last().finished)
	})

	t.Run("rate not found rolls back", func(t *testing.T) {
		currencyRepo := newFakeCurrencyRepo(usd, eur)
		rateRepo := newFakeExchangeRateRepo() // no rate recorded
		uowManager := &fakeUoWManager{}
		w := newExchangeRateWorkflow(currencyRepo, rateRepo, uowManager)

		_, err := w.GetByCurrencies(newAppContext(), "USD", "EUR", time.Now())

		require.Error(t, err)
		assert.True(t, uowManager.last().restored)
	})
}

func TestCurrencyExchangeRateWorkflow_List_IsAPassthroughWithNoTransaction(t *testing.T) {
	currencyRepo := newFakeCurrencyRepo()
	rateRepo := newFakeExchangeRateRepo(domain.CurrencyExchangeRate{FromCurrency: usd, ToCurrency: eur, Rate: 9_000, RecordedAt: time.Now()})
	uowManager := &fakeUoWManager{}
	w := newExchangeRateWorkflow(currencyRepo, rateRepo, uowManager)

	list, err := w.List(newAppContext())

	require.NoError(t, err)
	assert.Len(t, list, 1)
	assert.Empty(t, uowManager.created, "List is a plain passthrough and must not open a transaction")
}
