package summary_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"jww/internal/domain"
	"jww/internal/infrastructure/http_server/kernel"
	summaryrouter "jww/internal/infrastructure/http_server/routers/summary"
	"jww/internal/service/currency"
	currency_exchange_rate "jww/internal/service/currency_exchange_rate"
	"jww/internal/service/transaction"
	"jww/internal/service/unit_of_work"
	"jww/internal/service/user"
	"jww/internal/service/workflow"
)

// fakeCurrencyRepo implements currency.Repository with a single pre-seeded
// currency — enough for SummaryWorkflow.Generate's GetByCode lookup.
type fakeCurrencyRepo struct{ byCode map[string]domain.Currency }

func (r *fakeCurrencyRepo) Create(domain.AppContext, domain.Currency, ...unit_of_work.UnitOfWork) error {
	return nil
}
func (r *fakeCurrencyRepo) GetByCode(_ domain.AppContext, code string, _ ...unit_of_work.UnitOfWork) (*domain.Currency, error) {
	if c, ok := r.byCode[code]; ok {
		return &c, nil
	}
	return nil, nil
}
func (r *fakeCurrencyRepo) List(domain.AppContext, ...unit_of_work.UnitOfWork) ([]domain.Currency, error) {
	return nil, nil
}
func (r *fakeCurrencyRepo) Update(domain.AppContext, domain.Currency, ...unit_of_work.UnitOfWork) error {
	return nil
}
func (r *fakeCurrencyRepo) Delete(domain.AppContext, domain.Currency, ...unit_of_work.UnitOfWork) error {
	return nil
}

// fakeUserRepo implements user.Repository; empty is fine unless a test
// passes user_id.
type fakeUserRepo struct{ byID map[uuid.UUID]domain.User }

func (r *fakeUserRepo) Create(domain.AppContext, domain.User, ...unit_of_work.UnitOfWork) error {
	return nil
}
func (r *fakeUserRepo) GetByID(_ domain.AppContext, id uuid.UUID, _ ...unit_of_work.UnitOfWork) (*domain.User, error) {
	if u, ok := r.byID[id]; ok {
		return &u, nil
	}
	return nil, nil
}
func (r *fakeUserRepo) List(domain.AppContext, ...unit_of_work.UnitOfWork) ([]domain.User, error) {
	return nil, nil
}
func (r *fakeUserRepo) Update(domain.AppContext, domain.User, ...unit_of_work.UnitOfWork) error {
	return nil
}
func (r *fakeUserRepo) Delete(domain.AppContext, domain.User, ...unit_of_work.UnitOfWork) error {
	return nil
}

// fakeTransactionRepo implements transaction.Repository; List always
// returns an empty slice — this test exercises the HTTP/query-param layer,
// not the summary aggregation itself (that's workflow's own test suite).
type fakeTransactionRepo struct{}

func (r *fakeTransactionRepo) Create(domain.AppContext, domain.TransactionType, domain.Money, domain.Currency, domain.TransactionCategory, domain.User, domain.TransactionSource, time.Time, string, string, ...unit_of_work.UnitOfWork) (*domain.Transaction, error) {
	return nil, nil
}
func (r *fakeTransactionRepo) GetByID(domain.AppContext, int64, ...unit_of_work.UnitOfWork) (*domain.Transaction, error) {
	return nil, nil
}
func (r *fakeTransactionRepo) List(domain.AppContext, *transaction.Filter, ...unit_of_work.UnitOfWork) ([]domain.Transaction, error) {
	return nil, nil
}
func (r *fakeTransactionRepo) Update(domain.AppContext, domain.Transaction, ...unit_of_work.UnitOfWork) error {
	return nil
}
func (r *fakeTransactionRepo) Delete(domain.AppContext, domain.Transaction, ...unit_of_work.UnitOfWork) error {
	return nil
}

// fakeExchangeRateRepo implements currency_exchange_rate.Repository; unused
// with no transactions, but required to construct the service.
type fakeExchangeRateRepo struct{}

func (r *fakeExchangeRateRepo) Create(domain.AppContext, domain.CurrencyExchangeRate, ...unit_of_work.UnitOfWork) error {
	return nil
}
func (r *fakeExchangeRateRepo) GetByCurrencies(domain.AppContext, domain.Currency, domain.Currency, time.Time, ...unit_of_work.UnitOfWork) (*domain.CurrencyExchangeRate, error) {
	return nil, nil
}
func (r *fakeExchangeRateRepo) UpsertMany(domain.AppContext, []domain.CurrencyExchangeRateFlatten, ...unit_of_work.UnitOfWork) (currency_exchange_rate.UpsertCounts, error) {
	return currency_exchange_rate.UpsertCounts{}, nil
}
func (r *fakeExchangeRateRepo) List(domain.AppContext, ...unit_of_work.UnitOfWork) ([]domain.CurrencyExchangeRate, error) {
	return nil, nil
}

type fakeUoWManager struct{}

func (fakeUoWManager) Create(domain.AppContext) (unit_of_work.UnitOfWork, error) {
	return fakeUoW{}, nil
}

type fakeUoW struct{}

func (fakeUoW) Finish(domain.AppContext) error  { return nil }
func (fakeUoW) Restore(domain.AppContext) error { return nil }
func (fakeUoW) GetValue() interface{}           { return nil }

func newTestRouter(t *testing.T) *chi.Mux {
	t.Helper()
	currencyService := currency.NewService(&fakeCurrencyRepo{byCode: map[string]domain.Currency{
		"USD": {Code: "USD", Name: "US Dollar"},
	}}, fakeUoWManager{})
	userService := user.NewService(&fakeUserRepo{byID: map[uuid.UUID]domain.User{}}, nil, fakeUoWManager{})
	transactionService := transaction.NewService(&fakeTransactionRepo{}, fakeUoWManager{})
	exchangeRateService := currency_exchange_rate.NewService(&fakeExchangeRateRepo{})

	deps := &kernel.Deps{
		SummaryWorkflow: workflow.NewSummaryWorkflow(currencyService, userService, transactionService, exchangeRateService, fakeUoWManager{}),
	}
	r := chi.NewRouter()
	summaryrouter.Mount(r, r, deps)
	return r
}

func TestGetSummaryHandler_HappyPath(t *testing.T) {
	r := newTestRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/v1/summary?currency_code=USD&period_end=2026-01-31", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
}

func TestGetSummaryHandler_MissingCurrencyCode(t *testing.T) {
	r := newTestRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/v1/summary?period_end=2026-01-31", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestGetSummaryHandler_InvalidPeriodEnd(t *testing.T) {
	r := newTestRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/v1/summary?currency_code=USD&period_end=not-a-date", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestGetSummaryHandler_InvalidPeriod(t *testing.T) {
	r := newTestRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/v1/summary?currency_code=USD&period_end=2026-01-31&period=decade", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestGetSummaryHandler_InvalidUserID(t *testing.T) {
	r := newTestRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/v1/summary?currency_code=USD&period_end=2026-01-31&user_id=not-a-uuid", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestGetSummaryHandler_UnknownCurrencyIsNotFound(t *testing.T) {
	r := newTestRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/v1/summary?currency_code=ZZZ&period_end=2026-01-31", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}
