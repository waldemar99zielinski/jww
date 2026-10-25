package bootstrap

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"jww/internal/domain"
)

func TestGeneratedRatesCoverEveryDayAndDirectedPair(t *testing.T) {
	start := time.Date(2026, 8, 25, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 10, 25, 0, 0, 0, 0, time.UTC)
	currencies := []domain.Currency{{Code: "PLN"}, {Code: "USD"}, {Code: "EUR"}}

	rates := generatedRates(currencies, start, end)

	assert.Len(t, rates, daysInclusive(start, end)*6)
	assert.Equal(t, start, rates[0].RecordedAt)
	assert.Equal(t, end, rates[len(rates)-1].RecordedAt)
	for _, rate := range rates {
		assert.NotEqual(t, rate.FromCurrencyCode, rate.ToCurrencyCode)
		assert.Positive(t, rate.Rate)
	}
}

func TestTransactionPlansAreDeterministicAndValid(t *testing.T) {
	now := time.Date(2026, 10, 25, 18, 30, 0, 0, time.UTC)
	opts := Options{TransactionCount: 250, RandomSeed: 42, Now: now}
	users := []domain.User{{ID: uuid.New(), Username: "dev", DisplayName: "Dev"}}
	currencies := []domain.Currency{{Code: "PLN", Name: "Polish Zloty"}, {Code: "EUR", Name: "Euro"}}
	expenses := []domain.TransactionCategory{{ID: uuid.New(), Type: domain.TransactionCategoryTypeExpense, Name: "Groceries"}}
	incomes := []domain.TransactionCategory{{ID: uuid.New(), Type: domain.TransactionCategoryTypeIncome, Name: "Salary"}}

	first := transactionPlans(opts, users, currencies, expenses, incomes)
	second := transactionPlans(opts, users, currencies, expenses, incomes)

	require.Equal(t, first, second)
	start, _ := seedRange(now)
	seenIncome, seenExpense := false, false
	for i, plan := range first {
		assert.False(t, plan.CreatedAt.Before(start))
		assert.True(t, plan.CreatedAt.Before(now))
		index, ok := markerIndex(plan.Description)
		assert.True(t, ok)
		assert.Equal(t, i, index)
		assert.Equal(t, string(plan.Type), string(plan.Category.Type))
		assert.Positive(t, plan.Amount)
		seenIncome = seenIncome || plan.Type == domain.TransactionTypeIncome
		seenExpense = seenExpense || plan.Type == domain.TransactionTypeExpense
	}
	assert.True(t, seenIncome)
	assert.True(t, seenExpense)
}

func TestOptionsRejectImpossibleAttachmentCount(t *testing.T) {
	opts := DefaultOptions()
	opts.TransactionCount = 2
	opts.AttachmentCount = 3
	assert.Error(t, opts.validate())
}

func TestMarkerIndexRejectsUnrelatedDescriptions(t *testing.T) {
	index, ok := markerIndex("ordinary transaction")
	assert.False(t, ok)
	assert.Zero(t, index)
}
