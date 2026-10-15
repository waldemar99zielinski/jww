package workflow_test

import (
	"errors"
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
	return newExchangeRateWorkflowWithSource(currencyRepo, rateRepo, &fakeRateSource{}, uowManager)
}

func newExchangeRateWorkflowWithSource(currencyRepo *fakeCurrencyRepo, rateRepo *fakeExchangeRateRepo, rateSource *fakeRateSource, uowManager *fakeUoWManager) *workflow.CurrencyExchangeRateWorkflow {
	currencyService := currency.NewService(currencyRepo, uowManager)
	rateService := currency_exchange_rate.NewService(rateRepo)
	return workflow.NewCurrencyExchangeRateWorkflow(currencyService, rateService, rateSource, uowManager)
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

var pln = domain.Currency{Code: "PLN", Name: "Polish Złoty", IsBase: true}

func fullMatrix(asOf time.Time) []domain.CurrencyExchangeRateFlatten {
	return []domain.CurrencyExchangeRateFlatten{
		{FromCurrencyCode: "PLN", ToCurrencyCode: "USD", Rate: 2562, RecordedAt: asOf},
		{FromCurrencyCode: "PLN", ToCurrencyCode: "EUR", Rate: 2288, RecordedAt: asOf},
		{FromCurrencyCode: "USD", ToCurrencyCode: "PLN", Rate: 39030, RecordedAt: asOf},
		{FromCurrencyCode: "USD", ToCurrencyCode: "EUR", Rate: 9230, RecordedAt: asOf},
		{FromCurrencyCode: "EUR", ToCurrencyCode: "PLN", Rate: 43711, RecordedAt: asOf},
		{FromCurrencyCode: "EUR", ToCurrencyCode: "USD", Rate: 10834, RecordedAt: asOf},
	}
}

func TestCurrencyExchangeRateWorkflow_Sync(t *testing.T) {
	asOf := time.Date(2026, 10, 14, 0, 0, 0, 0, time.UTC)

	t.Run("success writes the whole matrix in one upsert and opens no transaction", func(t *testing.T) {
		currencyRepo := newFakeCurrencyRepo(pln, usd, eur)
		rateRepo := newFakeExchangeRateRepo()
		rateRepo.upsertCounts = currency_exchange_rate.UpsertCounts{Inserted: 6}
		source := &fakeRateSource{fetch: currency_exchange_rate.RateFetch{Rates: fullMatrix(asOf)}}
		uowManager := &fakeUoWManager{}
		w := newExchangeRateWorkflowWithSource(currencyRepo, rateRepo, source, uowManager)

		result, err := w.Sync(newAppContext(), asOf)

		require.NoError(t, err)
		assert.Equal(t, 3, result.Currencies)
		assert.Equal(t, 6, result.Fetched)
		assert.Equal(t, 6, result.Inserted)
		assert.Equal(t, 0, result.Updated)
		assert.Zero(t, result.Skipped)
		assert.Empty(t, result.Warnings)
		assert.Equal(t, 1, rateRepo.upsertCalls, "the whole matrix must go through one upsert call")
		assert.Len(t, rateRepo.upserted, 6)
		assert.Empty(t, uowManager.created, "a single write must not open a unit of work")
	})

	t.Run("a source error aborts without writing", func(t *testing.T) {
		currencyRepo := newFakeCurrencyRepo(pln, usd, eur)
		rateRepo := newFakeExchangeRateRepo()
		source := &fakeRateSource{err: errors.New("boom")}
		uowManager := &fakeUoWManager{}
		w := newExchangeRateWorkflowWithSource(currencyRepo, rateRepo, source, uowManager)

		_, err := w.Sync(newAppContext(), asOf)

		assert.ErrorIs(t, err, source.err)
		assert.Zero(t, rateRepo.upsertCalls)
		assert.Empty(t, uowManager.created)
	})

	t.Run("an upsert error propagates and opens no transaction", func(t *testing.T) {
		currencyRepo := newFakeCurrencyRepo(pln, usd, eur)
		rateRepo := newFakeExchangeRateRepo()
		rateRepo.upsertErr = errors.New("boom")
		source := &fakeRateSource{fetch: currency_exchange_rate.RateFetch{Rates: fullMatrix(asOf)}}
		uowManager := &fakeUoWManager{}
		w := newExchangeRateWorkflowWithSource(currencyRepo, rateRepo, source, uowManager)

		_, err := w.Sync(newAppContext(), asOf)

		assert.ErrorIs(t, err, rateRepo.upsertErr)
		assert.Empty(t, uowManager.created)
	})

	t.Run("fewer than two currencies short-circuits without calling the source", func(t *testing.T) {
		currencyRepo := newFakeCurrencyRepo(pln)
		rateRepo := newFakeExchangeRateRepo()
		source := &fakeRateSource{}
		uowManager := &fakeUoWManager{}
		w := newExchangeRateWorkflowWithSource(currencyRepo, rateRepo, source, uowManager)

		result, err := w.Sync(newAppContext(), asOf)

		require.NoError(t, err)
		assert.Equal(t, 1, result.Currencies)
		assert.Zero(t, source.calls)
		assert.Empty(t, uowManager.created)
	})

	t.Run("a skipped pair is surfaced as exactly one warning, never dropped", func(t *testing.T) {
		currencyRepo := newFakeCurrencyRepo(pln, usd, eur)
		rateRepo := newFakeExchangeRateRepo()
		source := &fakeRateSource{fetch: currency_exchange_rate.RateFetch{
			Rates:   []domain.CurrencyExchangeRateFlatten{{FromCurrencyCode: "USD", ToCurrencyCode: "EUR", Rate: 9230, RecordedAt: asOf}},
			Skipped: []currency_exchange_rate.RateSkip{{From: "EUR", To: "PLN", Reason: "provider returned no rate"}},
		}}
		uowManager := &fakeUoWManager{}
		w := newExchangeRateWorkflowWithSource(currencyRepo, rateRepo, source, uowManager)

		result, err := w.Sync(newAppContext(), asOf)

		require.NoError(t, err)
		assert.Equal(t, 1, result.Skipped)
		require.Len(t, result.Warnings, 1)
		assert.Contains(t, result.Warnings[0], "EUR→PLN")
		assert.Len(t, rateRepo.upserted, 1, "the readable pair still lands")
	})

	t.Run("an empty day is surfaced as exactly one warning", func(t *testing.T) {
		currencyRepo := newFakeCurrencyRepo(pln, usd)
		rateRepo := newFakeExchangeRateRepo()
		source := &fakeRateSource{}
		uowManager := &fakeUoWManager{}
		w := newExchangeRateWorkflowWithSource(currencyRepo, rateRepo, source, uowManager)

		result, err := w.Sync(newAppContext(), asOf)

		require.NoError(t, err)
		require.Len(t, result.Warnings, 1)
		assert.Contains(t, result.Warnings[0], "no rates")
	})

	t.Run("an unconfigured code is surfaced as exactly one warning and dropped", func(t *testing.T) {
		currencyRepo := newFakeCurrencyRepo(usd, eur)
		rateRepo := newFakeExchangeRateRepo()
		source := &fakeRateSource{fetch: currency_exchange_rate.RateFetch{
			Rates: []domain.CurrencyExchangeRateFlatten{{FromCurrencyCode: "GBP", ToCurrencyCode: "USD", Rate: 13000, RecordedAt: asOf}},
		}}
		uowManager := &fakeUoWManager{}
		w := newExchangeRateWorkflowWithSource(currencyRepo, rateRepo, source, uowManager)

		result, err := w.Sync(newAppContext(), asOf)

		require.NoError(t, err)
		require.Len(t, result.Warnings, 1)
		assert.Contains(t, result.Warnings[0], "not configured")
		assert.Empty(t, rateRepo.upserted, "an unconfigured code must not reach the repository")
	})

	t.Run("Sync forwards the supplied date unchanged", func(t *testing.T) {
		currencyRepo := newFakeCurrencyRepo(pln, usd)
		rateRepo := newFakeExchangeRateRepo()
		source := &fakeRateSource{}
		w := newExchangeRateWorkflowWithSource(currencyRepo, rateRepo, source, &fakeUoWManager{})

		explicit := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
		_, err := w.Sync(newAppContext(), explicit)

		require.NoError(t, err)
		assert.True(t, source.lastAsOf.Equal(explicit))
	})

	t.Run("SyncDaily asks for yesterday in UTC", func(t *testing.T) {
		currencyRepo := newFakeCurrencyRepo(pln, usd)
		rateRepo := newFakeExchangeRateRepo()
		source := &fakeRateSource{}
		w := newExchangeRateWorkflowWithSource(currencyRepo, rateRepo, source, &fakeUoWManager{})

		want := time.Now().UTC().AddDate(0, 0, -1).Format("2006-01-02")
		_, err := w.SyncDaily(newAppContext())

		require.NoError(t, err)
		assert.Equal(t, want, source.lastAsOf.Format("2006-01-02"))
	})
}
