package workflow_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"jww/internal/domain"
	"jww/internal/service/currency"
	currency_exchange_rate "jww/internal/service/currency_exchange_rate"
	"jww/internal/service/transaction"
	"jww/internal/service/user"
	"jww/internal/service/workflow"
)

type summaryWorkflowFixture struct {
	currencyRepo    *fakeCurrencyRepo
	userRepo        *fakeUserRepo
	transactionRepo *fakeTransactionRepo
	rateRepo        *fakeExchangeRateRepo
	uowManager      *fakeUoWManager
	workflow        *workflow.SummaryWorkflow
}

func newSummaryWorkflowFixture(currencies []domain.Currency, users []domain.User) *summaryWorkflowFixture {
	f := &summaryWorkflowFixture{
		currencyRepo:    newFakeCurrencyRepo(currencies...),
		userRepo:        newFakeUserRepo(users...),
		transactionRepo: &fakeTransactionRepo{},
		rateRepo:        newFakeExchangeRateRepo(),
		uowManager:      &fakeUoWManager{},
	}
	f.workflow = workflow.NewSummaryWorkflow(
		currency.NewService(f.currencyRepo, f.uowManager),
		user.NewService(f.userRepo, fakePasswordHasher{}, f.uowManager),
		transaction.NewService(f.transactionRepo, f.uowManager),
		currency_exchange_rate.NewService(f.rateRepo),
		f.uowManager,
	)
	return f
}

var periodEnd = time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)

func withinPeriod(days int) time.Time {
	return periodEnd.AddDate(0, 0, -days)
}

func TestSummaryWorkflow_Generate(t *testing.T) {
	t.Run("same currency transactions sum directly without an exchange rate lookup", func(t *testing.T) {
		f := newSummaryWorkflowFixture([]domain.Currency{usd}, nil)
		f.transactionRepo.listResult = []domain.Transaction{
			{Type: domain.TransactionTypeIncome, Amount: 10_000, Currency: usd, CreatedAt: withinPeriod(1)},
			{Type: domain.TransactionTypeExpense, Amount: 3_000, Currency: usd, CreatedAt: withinPeriod(2)},
		}

		summary, err := f.workflow.Generate(newAppContext(), nil, periodEnd, "USD", nil, nil)

		require.NoError(t, err)
		assert.Equal(t, domain.Money(10_000), summary.TotalIncome)
		assert.Equal(t, domain.Money(3_000), summary.TotalExpense)
		assert.Equal(t, domain.Money(7_000), summary.Net)
		assert.Equal(t, 0, f.rateRepo.getCalls, "converting to its own currency must not look up a rate")
		assert.True(t, f.uowManager.last().finished)
	})

	t.Run("cross-currency transactions are converted via the exchange rate service", func(t *testing.T) {
		f := newSummaryWorkflowFixture([]domain.Currency{usd, eur}, nil)
		f.rateRepo.byPair["EUR->USD"] = domain.CurrencyExchangeRate{FromCurrency: eur, ToCurrency: usd, Rate: domain.ExchangeRate(11_000)} // 1 EUR = 1.1 USD
		f.transactionRepo.listResult = []domain.Transaction{
			{Type: domain.TransactionTypeExpense, Amount: 1_000, Currency: eur, CreatedAt: withinPeriod(1)}, // 10.00 EUR
		}

		summary, err := f.workflow.Generate(newAppContext(), nil, periodEnd, "USD", nil, nil)

		require.NoError(t, err)
		assert.Equal(t, domain.Money(1_100), summary.TotalExpense, "10.00 EUR at 1.1 should convert to 11.00 USD")
		assert.Equal(t, 1, f.rateRepo.getCalls)
		stats, ok := summary.ByCurrency["EUR"]
		require.True(t, ok, "ByCurrency should still report the original, unconverted EUR totals")
		assert.Equal(t, domain.Money(1_000), stats.TotalExpense)
	})

	t.Run("unknown summary currency rolls back before listing transactions", func(t *testing.T) {
		f := newSummaryWorkflowFixture(nil, nil)

		_, err := f.workflow.Generate(newAppContext(), nil, periodEnd, "USD", nil, nil)

		require.Error(t, err)
		assert.True(t, f.uowManager.last().restored)
		assert.Zero(t, f.transactionRepo.listCalled)
	})

	t.Run("unknown user filter rolls back before listing transactions", func(t *testing.T) {
		f := newSummaryWorkflowFixture([]domain.Currency{usd}, nil)
		missing := uuid.New()

		_, err := f.workflow.Generate(newAppContext(), nil, periodEnd, "USD", nil, &missing)

		require.Error(t, err)
		assert.True(t, f.uowManager.last().restored)
		assert.Zero(t, f.transactionRepo.listCalled)
	})

	t.Run("transaction listing failure rolls back", func(t *testing.T) {
		f := newSummaryWorkflowFixture([]domain.Currency{usd}, nil)
		f.transactionRepo.listErr = assert.AnError

		_, err := f.workflow.Generate(newAppContext(), nil, periodEnd, "USD", nil, nil)

		require.Error(t, err)
		assert.True(t, f.uowManager.last().restored)
	})

	// Regression test for the bug an earlier audit found and fixed: a
	// missing active.Restore(ctx) on the conversion error path in
	// statsByCurrency/convert. Any future regression here must fail this
	// test.
	t.Run("a missing exchange rate mid-aggregation rolls back (regression: Restore was once skipped here)", func(t *testing.T) {
		f := newSummaryWorkflowFixture([]domain.Currency{usd, eur}, nil)
		// No EUR->USD rate registered, so convert() will fail.
		f.transactionRepo.listResult = []domain.Transaction{
			{Type: domain.TransactionTypeExpense, Amount: 1_000, Currency: eur, CreatedAt: withinPeriod(1)},
		}

		_, err := f.workflow.Generate(newAppContext(), nil, periodEnd, "USD", nil, nil)

		require.Error(t, err)
		assert.True(t, f.uowManager.last().restored, "Restore must fire when currency conversion fails partway through aggregation")
		assert.False(t, f.uowManager.last().finished)
	})

	t.Run("a period_end before period_start fails Summary.Validate and rolls back", func(t *testing.T) {
		f := newSummaryWorkflowFixture([]domain.Currency{usd}, nil)
		periodStart := periodEnd.AddDate(0, 1, 0) // after periodEnd

		_, err := f.workflow.Generate(newAppContext(), nil, periodEnd, "USD", &periodStart, nil)

		require.Error(t, err)
		assert.True(t, f.uowManager.last().restored)
	})
}
