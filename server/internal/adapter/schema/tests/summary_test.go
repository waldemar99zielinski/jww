package schema_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"jww/internal/adapter/schema"
	"jww/internal/domain"
)

func TestSummaryToResponse(t *testing.T) {
	period := domain.SummaryPeriodMonth
	periodStart := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	periodEnd := time.Date(2026, 2, 28, 0, 0, 0, 0, time.UTC)
	usd := domain.Currency{Code: "USD", Name: "US Dollar"}

	s := domain.Summary{
		Period:       &period,
		PeriodStart:  &periodStart,
		PeriodEnd:    periodEnd,
		Currency:     usd,
		TotalIncome:  domain.Money(5000),
		TotalExpense: domain.Money(2000),
		Net:          domain.Money(3000),
		ByCurrency: map[string]domain.CurrencyStats{
			"USD": {Currency: usd, TotalIncome: 5000, TotalExpense: 2000, Net: 3000},
		},
		Transactions: []domain.Transaction{{ID: 1}, {ID: 2}},
	}

	resp := schema.SummaryToResponse(s)

	require.NotNil(t, resp.Period)
	assert.Equal(t, period, *resp.Period)
	require.NotNil(t, resp.PeriodStart)
	assert.Equal(t, "2026-02-01", resp.PeriodStart.String())
	assert.Equal(t, "2026-02-28", resp.PeriodEnd.String())
	assert.Equal(t, "USD", resp.CurrencyCode)
	assert.Equal(t, domain.Money(5000), resp.TotalIncome)
	assert.Equal(t, domain.Money(2000), resp.TotalExpense)
	assert.Equal(t, domain.Money(3000), resp.Net)
	assert.Equal(t, 2, resp.TransactionCount)

	require.Contains(t, resp.ByCurrency, "USD")
	usdStats := resp.ByCurrency["USD"]
	assert.Equal(t, "USD", usdStats.CurrencyCode)
	assert.Equal(t, domain.Money(5000), usdStats.TotalIncome)
	assert.Equal(t, domain.Money(2000), usdStats.TotalExpense)
	assert.Equal(t, domain.Money(3000), usdStats.Net)
}

func TestSummaryToResponse_NilPeriodStart(t *testing.T) {
	// A custom range with no explicit start ("everything up to period_end")
	// must not panic and must serialize PeriodStart as nil.
	s := domain.Summary{
		PeriodEnd: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		Currency:  domain.Currency{Code: "USD", Name: "US Dollar"},
	}

	resp := schema.SummaryToResponse(s)

	assert.Nil(t, resp.PeriodStart)
	assert.Nil(t, resp.Period)
}
