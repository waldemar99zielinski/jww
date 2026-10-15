package schema

import (
	"jww/internal/domain"
)

type CurrencyStatsResponse struct {
	CurrencyCode string       `json:"currency_code"`
	TotalIncome  domain.Money `json:"total_income"`
	TotalExpense domain.Money `json:"total_expense"`
	Net          domain.Money `json:"net"`
}

type SummaryResponse struct {
	Period           *domain.SummaryPeriod            `json:"period"`
	PeriodStart      *Date                            `json:"period_start"`
	PeriodEnd        Date                             `json:"period_end"`
	CurrencyCode     string                           `json:"currency_code"`
	TotalIncome      domain.Money                     `json:"total_income"`
	TotalExpense     domain.Money                     `json:"total_expense"`
	Net              domain.Money                     `json:"net"`
	ByCurrency       map[string]CurrencyStatsResponse `json:"by_currency"`
	TransactionCount int                              `json:"transaction_count"`
}

func SummaryToResponse(s domain.Summary) SummaryResponse {
	byCurrency := make(map[string]CurrencyStatsResponse, len(s.ByCurrency))
	for code, stats := range s.ByCurrency {
		byCurrency[code] = CurrencyStatsResponse{
			CurrencyCode: stats.Currency.Code,
			TotalIncome:  stats.TotalIncome,
			TotalExpense: stats.TotalExpense,
			Net:          stats.Net,
		}
	}
	var periodStart *Date
	if s.PeriodStart != nil {
		d := Date(*s.PeriodStart)
		periodStart = &d
	}
	return SummaryResponse{
		Period:           s.Period,
		PeriodStart:      periodStart,
		PeriodEnd:        Date(s.PeriodEnd),
		CurrencyCode:     s.Currency.Code,
		TotalIncome:      s.TotalIncome,
		TotalExpense:     s.TotalExpense,
		Net:              s.Net,
		ByCurrency:       byCurrency,
		TransactionCount: len(s.Transactions),
	}
}
