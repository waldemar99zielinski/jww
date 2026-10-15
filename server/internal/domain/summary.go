package domain

import (
	"fmt"
	"time"
)

// Summary.PeriodStart and PeriodEnd are date-only values (time-of-day is
// ignored); callers should pass time.Time truncated to midnight UTC.
type Summary struct {
	Period       *SummaryPeriod
	PeriodStart  *time.Time
	PeriodEnd    time.Time
	Currency     Currency
	TotalIncome  Money
	TotalExpense Money
	Net          Money
	ByCurrency   map[string]CurrencyStats
	Transactions []Transaction
	User         *User
}

func dateOnly(t time.Time) time.Time {
	t = t.UTC()
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

func (s Summary) Validate() error {
	periodEnd := dateOnly(s.PeriodEnd)
	if s.Period == nil {
		// No discrete bucket (day/week/month/year): an explicit custom
		// range. PeriodStart may still be nil, meaning "everything up to
		// PeriodEnd."
		if s.PeriodStart != nil && periodEnd.Before(dateOnly(*s.PeriodStart)) {
			return fmt.Errorf("period_end must not be before period_start")
		}
	} else {
		if s.PeriodStart == nil {
			return fmt.Errorf("period_start is required for a discrete (day/week/month/year) period")
		}
		if periodEnd.Before(dateOnly(*s.PeriodStart)) {
			return fmt.Errorf("period_end must not be before period_start")
		}
	}
	if s.TotalIncome < 0 {
		return fmt.Errorf("total_income must not be negative")
	}
	if s.TotalExpense < 0 {
		return fmt.Errorf("total_expense must not be negative")
	}
	if s.Net != s.TotalIncome-s.TotalExpense {
		return fmt.Errorf("net must equal total_income minus total_expense")
	}
	for code, stats := range s.ByCurrency {
		if stats.Currency.Code != code {
			return fmt.Errorf("by_currency key must match its CurrencyStats.Currency.Code")
		}
	}
	for _, transaction := range s.Transactions {
		transactionDate := dateOnly(transaction.CreatedAt)
		if s.PeriodStart != nil && transactionDate.Before(dateOnly(*s.PeriodStart)) {
			return fmt.Errorf("all transactions must fall within the summary period")
		}
		if transactionDate.After(periodEnd) {
			return fmt.Errorf("all transactions must fall within the summary period")
		}
	}
	return nil
}
