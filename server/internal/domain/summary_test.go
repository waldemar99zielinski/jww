package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func validSummary() Summary {
	return Summary{
		PeriodEnd:    time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC),
		Currency:     Currency{Code: "USD", Name: "US Dollar"},
		TotalIncome:  Money(1000),
		TotalExpense: Money(400),
		Net:          Money(600),
	}
}

func day(y int, m time.Month, d int) time.Time {
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func TestSummary_Validate_PeriodRange(t *testing.T) {
	t.Run("custom range: period start nil is fine", func(t *testing.T) {
		s := validSummary()
		s.Period = nil
		s.PeriodStart = nil
		assert.NoError(t, s.Validate())
	})

	t.Run("custom range: end on/after start is fine", func(t *testing.T) {
		s := validSummary()
		start := day(2026, 1, 1)
		s.Period = nil
		s.PeriodStart = &start
		s.PeriodEnd = day(2026, 1, 31)
		assert.NoError(t, s.Validate())
	})

	t.Run("custom range: end before start is rejected", func(t *testing.T) {
		s := validSummary()
		start := day(2026, 2, 1)
		s.Period = nil
		s.PeriodStart = &start
		s.PeriodEnd = day(2026, 1, 1)
		assert.ErrorContains(t, s.Validate(), "period_end must not be before period_start")
	})

	t.Run("discrete period requires a period start", func(t *testing.T) {
		s := validSummary()
		period := SummaryPeriodMonth
		s.Period = &period
		s.PeriodStart = nil
		assert.ErrorContains(t, s.Validate(), "period_start is required")
	})

	t.Run("discrete period: end before start is rejected", func(t *testing.T) {
		s := validSummary()
		period := SummaryPeriodMonth
		start := day(2026, 2, 1)
		s.Period = &period
		s.PeriodStart = &start
		s.PeriodEnd = day(2026, 1, 1)
		assert.ErrorContains(t, s.Validate(), "period_end must not be before period_start")
	})

	t.Run("discrete period: valid range", func(t *testing.T) {
		s := validSummary()
		period := SummaryPeriodMonth
		start := day(2026, 1, 1)
		s.Period = &period
		s.PeriodStart = &start
		s.PeriodEnd = day(2026, 1, 31)
		assert.NoError(t, s.Validate())
	})
}

func TestSummary_Validate_Totals(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(s Summary) Summary
		wantErr string
	}{
		{"valid", func(s Summary) Summary { return s }, ""},
		{"negative total income", func(s Summary) Summary { s.TotalIncome = -1; return s }, "total_income"},
		{"negative total expense", func(s Summary) Summary { s.TotalExpense = -1; return s }, "total_expense"},
		{"net does not match income minus expense", func(s Summary) Summary { s.Net = 0; return s }, "net"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.mutate(validSummary()).Validate()
			if tt.wantErr == "" {
				assert.NoError(t, err)
				return
			}
			assert.ErrorContains(t, err, tt.wantErr)
		})
	}
}

func TestSummary_Validate_ByCurrencyKeyMustMatchItsCurrencyCode(t *testing.T) {
	s := validSummary()
	s.ByCurrency = map[string]CurrencyStats{
		"EUR": {Currency: Currency{Code: "USD", Name: "US Dollar"}, Net: 0},
	}
	assert.ErrorContains(t, s.Validate(), "by_currency")
}

func TestSummary_Validate_ByCurrencyKeyMatchingItsCodeIsFine(t *testing.T) {
	s := validSummary()
	s.ByCurrency = map[string]CurrencyStats{
		"USD": {Currency: Currency{Code: "USD", Name: "US Dollar"}, TotalIncome: 1000, TotalExpense: 400, Net: 600},
	}
	assert.NoError(t, s.Validate())
}

func TestSummary_Validate_TransactionsMustFallWithinPeriod(t *testing.T) {
	start := day(2026, 1, 1)
	base := validSummary()
	base.PeriodStart = &start
	base.PeriodEnd = day(2026, 1, 31)

	t.Run("transaction inside the period is fine", func(t *testing.T) {
		s := base
		s.Transactions = []Transaction{{CreatedAt: day(2026, 1, 15)}}
		assert.NoError(t, s.Validate())
	})

	t.Run("transaction before period start is rejected", func(t *testing.T) {
		s := base
		s.Transactions = []Transaction{{CreatedAt: day(2025, 12, 31)}}
		assert.ErrorContains(t, s.Validate(), "fall within the summary period")
	})

	t.Run("transaction after period end is rejected", func(t *testing.T) {
		s := base
		s.Transactions = []Transaction{{CreatedAt: day(2026, 2, 1)}}
		assert.ErrorContains(t, s.Validate(), "fall within the summary period")
	})
}
