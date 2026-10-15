package domain

import "fmt"

// CurrencyStats aggregates income/expense totals for a single currency
// within a Summary.
type CurrencyStats struct {
	Currency     Currency
	TotalIncome  Money
	TotalExpense Money
	Net          Money
}

func (s CurrencyStats) Validate() error {
	if s.TotalIncome < 0 {
		return fmt.Errorf("total_income must not be negative")
	}
	if s.TotalExpense < 0 {
		return fmt.Errorf("total_expense must not be negative")
	}
	if s.Net != s.TotalIncome-s.TotalExpense {
		return fmt.Errorf("net must equal total_income minus total_expense")
	}
	return nil
}
