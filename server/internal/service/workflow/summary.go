package workflow

import (
	"time"

	"github.com/google/uuid"

	"jww/internal/domain"
	"jww/internal/service/currency"
	currency_exchange_rate "jww/internal/service/currency_exchange_rate"
	"jww/internal/service/transaction"
	"jww/internal/service/unit_of_work"
	"jww/internal/service/user"
)

type SummaryWorkflow struct {
	currencyService             *currency.Service
	userService                 *user.Service
	transactionService          *transaction.Service
	currencyExchangeRateService *currency_exchange_rate.Service
	uowManager                  unit_of_work.Manager
}

func NewSummaryWorkflow(
	currencyService *currency.Service,
	userService *user.Service,
	transactionService *transaction.Service,
	currencyExchangeRateService *currency_exchange_rate.Service,
	uowManager unit_of_work.Manager,
) *SummaryWorkflow {
	return &SummaryWorkflow{
		currencyService:             currencyService,
		userService:                 userService,
		transactionService:          transactionService,
		currencyExchangeRateService: currencyExchangeRateService,
		uowManager:                  uowManager,
	}
}

func (w *SummaryWorkflow) Generate(
	ctx domain.AppContext,
	period *domain.SummaryPeriod,
	periodEnd time.Time,
	currencyCode string,
	periodStart *time.Time,
	userID *uuid.UUID,
) (*domain.Summary, error) {
	uow, err := w.uowManager.Create(ctx)
	if err != nil {
		return nil, err
	}
	currency, err := w.currencyService.GetByCode(ctx, currencyCode, uow)
	if err != nil {
		_ = uow.Restore(ctx)
		return nil, err
	}
	var u *domain.User
	if userID != nil {
		u, err = w.userService.GetByID(ctx, *userID, uow)
		if err != nil {
			_ = uow.Restore(ctx)
			return nil, err
		}
	}

	var createdFrom *time.Time
	if periodStart != nil {
		t := time.Date(periodStart.Year(), periodStart.Month(), periodStart.Day(), 0, 0, 0, 0, time.UTC)
		createdFrom = &t
	}
	createdTo := time.Date(periodEnd.Year(), periodEnd.Month(), periodEnd.Day(), 23, 59, 59, 999999999, time.UTC)

	transactions, err := w.transactionService.List(ctx, &transaction.Filter{
		UserID:      userID,
		CreatedFrom: createdFrom,
		CreatedTo:   &createdTo,
	}, uow)
	if err != nil {
		_ = uow.Restore(ctx)
		return nil, err
	}

	byCurrency := w.statsByCurrency(transactions)

	var totalIncome, totalExpense domain.Money
	for _, t := range transactions {
		converted, err := w.convert(ctx, t, *currency, uow)
		if err != nil {
			_ = uow.Restore(ctx)
			return nil, err
		}
		if t.Type == domain.TransactionTypeIncome {
			totalIncome += converted
		} else {
			totalExpense += converted
		}
	}

	summary := domain.Summary{
		Period:       period,
		PeriodStart:  periodStart,
		PeriodEnd:    periodEnd,
		Currency:     *currency,
		TotalIncome:  totalIncome,
		TotalExpense: totalExpense,
		Net:          totalIncome - totalExpense,
		ByCurrency:   byCurrency,
		Transactions: transactions,
		User:         u,
	}
	if err := summary.Validate(); err != nil {
		_ = uow.Restore(ctx)
		return nil, err
	}
	if err := uow.Finish(ctx); err != nil {
		return nil, err
	}
	return &summary, nil
}

func (w *SummaryWorkflow) convert(ctx domain.AppContext, t domain.Transaction, currency domain.Currency, uow unit_of_work.UnitOfWork) (domain.Money, error) {
	if t.Currency.Code == currency.Code {
		return t.Amount, nil
	}
	rate, err := w.currencyExchangeRateService.GetByCurrencies(ctx, t.Currency, currency, t.CreatedAt, uow)
	if err != nil {
		return 0, err
	}
	return rate.Rate.Convert(t.Amount), nil
}

func (w *SummaryWorkflow) statsByCurrency(transactions []domain.Transaction) map[string]domain.CurrencyStats {
	incomeByCode := map[string]domain.Money{}
	expenseByCode := map[string]domain.Money{}
	currencyByCode := map[string]domain.Currency{}

	for _, t := range transactions {
		code := t.Currency.Code
		currencyByCode[code] = t.Currency
		if t.Type == domain.TransactionTypeIncome {
			incomeByCode[code] += t.Amount
		} else {
			expenseByCode[code] += t.Amount
		}
	}

	byCurrency := make(map[string]domain.CurrencyStats, len(currencyByCode))
	for code, curr := range currencyByCode {
		income := incomeByCode[code]
		expense := expenseByCode[code]
		byCurrency[code] = domain.CurrencyStats{
			Currency:     curr,
			TotalIncome:  income,
			TotalExpense: expense,
			Net:          income - expense,
		}
	}
	return byCurrency
}
