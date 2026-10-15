package currency_exchange_rate_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"jww/internal/adapter/schema"
	"jww/internal/domain"
	"jww/internal/infrastructure/http_server/kernel"
	exchangeraterouter "jww/internal/infrastructure/http_server/routers/currency_exchange_rate"
	"jww/internal/service/currency"
	ratesvc "jww/internal/service/currency_exchange_rate"
	"jww/internal/service/unit_of_work"
	"jww/internal/service/workflow"
)

type fakeCurrencyRepo struct {
	byCode map[string]domain.Currency
}

func (f *fakeCurrencyRepo) Create(domain.AppContext, domain.Currency, ...unit_of_work.UnitOfWork) error {
	return nil
}
func (f *fakeCurrencyRepo) GetByCode(_ domain.AppContext, code string, _ ...unit_of_work.UnitOfWork) (*domain.Currency, error) {
	c, ok := f.byCode[code]
	if !ok {
		return nil, nil
	}
	return &c, nil
}
func (f *fakeCurrencyRepo) List(domain.AppContext, ...unit_of_work.UnitOfWork) ([]domain.Currency, error) {
	return nil, nil
}
func (f *fakeCurrencyRepo) Update(domain.AppContext, domain.Currency, ...unit_of_work.UnitOfWork) error {
	return nil
}
func (f *fakeCurrencyRepo) Delete(domain.AppContext, domain.Currency, ...unit_of_work.UnitOfWork) error {
	return nil
}

type fakeRateRepo struct {
	rates []domain.CurrencyExchangeRate
}

func (f *fakeRateRepo) Create(_ domain.AppContext, r domain.CurrencyExchangeRate, _ ...unit_of_work.UnitOfWork) error {
	f.rates = append(f.rates, r)
	return nil
}

func (f *fakeRateRepo) GetByCurrencies(_ domain.AppContext, from, to domain.Currency, asOf time.Time, _ ...unit_of_work.UnitOfWork) (*domain.CurrencyExchangeRate, error) {
	for _, r := range f.rates {
		if r.FromCurrency.Code == from.Code && r.ToCurrency.Code == to.Code {
			return &r, nil
		}
	}
	return nil, nil
}

func (f *fakeRateRepo) List(domain.AppContext, ...unit_of_work.UnitOfWork) ([]domain.CurrencyExchangeRate, error) {
	return f.rates, nil
}

type fakeUoWManager struct{}

func (fakeUoWManager) Create(domain.AppContext) (unit_of_work.UnitOfWork, error) {
	return fakeUoW{}, nil
}

type fakeUoW struct{}

func (fakeUoW) Finish(domain.AppContext) error  { return nil }
func (fakeUoW) Restore(domain.AppContext) error { return nil }
func (fakeUoW) GetValue() interface{}           { return nil }

func newTestRouter(t *testing.T) (*chi.Mux, *fakeCurrencyRepo, *fakeRateRepo) {
	t.Helper()
	currencyRepo := &fakeCurrencyRepo{byCode: map[string]domain.Currency{
		"USD": {Code: "USD", Name: "US Dollar"},
		"EUR": {Code: "EUR", Name: "Euro"},
	}}
	rateRepo := &fakeRateRepo{}
	currencyService := currency.NewService(currencyRepo, fakeUoWManager{})
	rateService := ratesvc.NewService(rateRepo)
	deps := &kernel.Deps{
		CurrencyExchangeRateWorkflow: workflow.NewCurrencyExchangeRateWorkflow(currencyService, rateService, fakeUoWManager{}),
	}
	r := chi.NewRouter()
	exchangeraterouter.Mount(r, r, deps)
	return r, currencyRepo, rateRepo
}

func decodeData(t *testing.T, rec *httptest.ResponseRecorder, v any) {
	t.Helper()
	var body struct {
		Data json.RawMessage `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.NoError(t, json.Unmarshal(body.Data, v))
}

func TestCreateExchangeRateHandler(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		r, _, _ := newTestRouter(t)
		req := httptest.NewRequest(http.MethodPost, "/v1/currency-exchange-rates", bytes.NewBufferString(`{"from_currency_code":"USD","to_currency_code":"EUR","rate":"0.9100"}`))
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		require.Equal(t, http.StatusCreated, rec.Code)
		var resp schema.CurrencyExchangeRateResponse
		decodeData(t, rec, &resp)
		assert.Equal(t, "USD", resp.FromCurrencyCode)
		assert.Equal(t, "EUR", resp.ToCurrencyCode)
	})

	t.Run("unknown from currency is not found, not a 500", func(t *testing.T) {
		r, _, _ := newTestRouter(t)
		req := httptest.NewRequest(http.MethodPost, "/v1/currency-exchange-rates", bytes.NewBufferString(`{"from_currency_code":"ZZZ","to_currency_code":"EUR","rate":"1.0000"}`))
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("failed Validate (non-positive rate)", func(t *testing.T) {
		r, _, _ := newTestRouter(t)
		req := httptest.NewRequest(http.MethodPost, "/v1/currency-exchange-rates", bytes.NewBufferString(`{"from_currency_code":"USD","to_currency_code":"EUR","rate":"0"}`))
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func TestGetExchangeRateHandler(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		r, _, rateRepo := newTestRouter(t)
		rateRepo.rates = append(rateRepo.rates, domain.CurrencyExchangeRate{
			FromCurrency: domain.Currency{Code: "USD"}, ToCurrency: domain.Currency{Code: "EUR"},
			Rate: domain.ExchangeRate(9100), RecordedAt: time.Now(),
		})

		req := httptest.NewRequest(http.MethodGet, "/v1/currency-exchange-rates?from_currency_code=USD&to_currency_code=EUR&as_of=2026-01-01", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("missing query params", func(t *testing.T) {
		r, _, _ := newTestRouter(t)
		req := httptest.NewRequest(http.MethodGet, "/v1/currency-exchange-rates?as_of=2026-01-01", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("not found", func(t *testing.T) {
		r, _, _ := newTestRouter(t)
		req := httptest.NewRequest(http.MethodGet, "/v1/currency-exchange-rates?from_currency_code=USD&to_currency_code=EUR&as_of=2026-01-01", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}

func TestListExchangeRatesHandler(t *testing.T) {
	r, _, rateRepo := newTestRouter(t)
	rateRepo.rates = append(rateRepo.rates, domain.CurrencyExchangeRate{
		FromCurrency: domain.Currency{Code: "USD"}, ToCurrency: domain.Currency{Code: "EUR"},
		Rate: domain.ExchangeRate(9100), RecordedAt: time.Now(),
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/currency-exchange-rates/all", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var resp []schema.CurrencyExchangeRateResponse
	decodeData(t, rec, &resp)
	assert.Len(t, resp, 1)
}
