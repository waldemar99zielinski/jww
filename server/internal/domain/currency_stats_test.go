package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func validCurrencyStats() CurrencyStats {
	return CurrencyStats{
		Currency:     Currency{Code: "USD", Name: "US Dollar"},
		TotalIncome:  Money(1000),
		TotalExpense: Money(400),
		Net:          Money(600),
	}
}

func TestCurrencyStats_Validate(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(s CurrencyStats) CurrencyStats
		wantErr string
	}{
		{"valid", func(s CurrencyStats) CurrencyStats { return s }, ""},
		{"negative total income", func(s CurrencyStats) CurrencyStats { s.TotalIncome = -1; return s }, "total_income"},
		{"negative total expense", func(s CurrencyStats) CurrencyStats { s.TotalExpense = -1; return s }, "total_expense"},
		{"net does not match income minus expense", func(s CurrencyStats) CurrencyStats { s.Net = 0; return s }, "net"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.mutate(validCurrencyStats()).Validate()
			if tt.wantErr == "" {
				assert.NoError(t, err)
				return
			}
			assert.ErrorContains(t, err, tt.wantErr)
		})
	}
}
