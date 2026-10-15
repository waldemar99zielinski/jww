package currency_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"jww/internal/adapter/schema"
	"jww/internal/domain"
	"jww/internal/infrastructure/http_server/kernel"
	currencyrouter "jww/internal/infrastructure/http_server/routers/currency"
	"jww/internal/service/currency"
	"jww/internal/service/unit_of_work"
)

// fakeRepository implements currency.Repository in memory.
type fakeRepository struct {
	byCode map[string]domain.Currency
}

func newFakeRepository() *fakeRepository {
	return &fakeRepository{byCode: map[string]domain.Currency{}}
}

func (f *fakeRepository) Create(_ domain.AppContext, c domain.Currency, _ ...unit_of_work.UnitOfWork) error {
	f.byCode[c.Code] = c
	return nil
}

func (f *fakeRepository) GetByCode(_ domain.AppContext, code string, _ ...unit_of_work.UnitOfWork) (*domain.Currency, error) {
	c, ok := f.byCode[code]
	if !ok {
		return nil, nil
	}
	return &c, nil
}

func (f *fakeRepository) List(_ domain.AppContext, _ ...unit_of_work.UnitOfWork) ([]domain.Currency, error) {
	out := make([]domain.Currency, 0, len(f.byCode))
	for _, c := range f.byCode {
		out = append(out, c)
	}
	return out, nil
}

func (f *fakeRepository) Update(_ domain.AppContext, c domain.Currency, _ ...unit_of_work.UnitOfWork) error {
	f.byCode[c.Code] = c
	return nil
}

func (f *fakeRepository) Delete(_ domain.AppContext, c domain.Currency, _ ...unit_of_work.UnitOfWork) error {
	delete(f.byCode, c.Code)
	return nil
}

// fakeUoWManager is a no-op Unit of Work — router tests exercise the HTTP
// plumbing, not transaction demarcation (that's covered by the service-layer
// tests), so Create/Finish/Restore never fail and never need to coordinate
// with the fake repository above.
type fakeUoWManager struct{}

func (fakeUoWManager) Create(domain.AppContext) (unit_of_work.UnitOfWork, error) {
	return fakeUoW{}, nil
}

type fakeUoW struct{}

func (fakeUoW) Finish(domain.AppContext) error  { return nil }
func (fakeUoW) Restore(domain.AppContext) error { return nil }
func (fakeUoW) GetValue() interface{}           { return nil }

func newTestRouter(t *testing.T) (*chi.Mux, *fakeRepository) {
	t.Helper()
	repo := newFakeRepository()
	deps := &kernel.Deps{
		CurrencyService: currency.NewService(repo, fakeUoWManager{}),
	}
	r := chi.NewRouter()
	currencyrouter.Mount(r, r, deps)
	return r, repo
}

func decodeData(t *testing.T, rec *httptest.ResponseRecorder, v any) {
	t.Helper()
	var body struct {
		Data json.RawMessage `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.NoError(t, json.Unmarshal(body.Data, v))
}

func TestCreateCurrencyHandler(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		r, _ := newTestRouter(t)
		req := httptest.NewRequest(http.MethodPost, "/v1/currencies", bytes.NewBufferString(`{"code":"USD","name":"US Dollar","is_base":true}`))
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		require.Equal(t, http.StatusCreated, rec.Code)
		var resp schema.CurrencyResponse
		decodeData(t, rec, &resp)
		assert.Equal(t, "USD", resp.Code)
		assert.Equal(t, "US Dollar", resp.Name)
		assert.True(t, resp.IsBase)
	})

	t.Run("bad JSON", func(t *testing.T) {
		r, _ := newTestRouter(t)
		req := httptest.NewRequest(http.MethodPost, "/v1/currencies", bytes.NewBufferString(`{"code":`))
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("failed Validate", func(t *testing.T) {
		r, _ := newTestRouter(t)
		req := httptest.NewRequest(http.MethodPost, "/v1/currencies", bytes.NewBufferString(`{"code":"US","name":"US Dollar"}`))
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func TestGetCurrencyHandler(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		r, repo := newTestRouter(t)
		_ = repo.Create(domain.AppContext{}, domain.Currency{Code: "EUR", Name: "Euro"})

		req := httptest.NewRequest(http.MethodGet, "/v1/currencies/EUR", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		var resp schema.CurrencyResponse
		decodeData(t, rec, &resp)
		assert.Equal(t, "EUR", resp.Code)
	})

	t.Run("not found", func(t *testing.T) {
		r, _ := newTestRouter(t)
		req := httptest.NewRequest(http.MethodGet, "/v1/currencies/ZZZ", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}

func TestListCurrenciesHandler(t *testing.T) {
	r, repo := newTestRouter(t)
	_ = repo.Create(domain.AppContext{}, domain.Currency{Code: "USD", Name: "US Dollar"})
	_ = repo.Create(domain.AppContext{}, domain.Currency{Code: "EUR", Name: "Euro"})

	req := httptest.NewRequest(http.MethodGet, "/v1/currencies", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var resp []schema.CurrencyResponse
	decodeData(t, rec, &resp)
	assert.Len(t, resp, 2)
}

func TestUpdateCurrencyHandler(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		r, repo := newTestRouter(t)
		_ = repo.Create(domain.AppContext{}, domain.Currency{Code: "USD", Name: "US Dollar"})

		req := httptest.NewRequest(http.MethodPut, "/v1/currencies/USD", bytes.NewBufferString(`{"name":"United States Dollar","is_base":true}`))
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		var resp schema.CurrencyResponse
		decodeData(t, rec, &resp)
		assert.Equal(t, "United States Dollar", resp.Name)
	})

	t.Run("not found", func(t *testing.T) {
		r, _ := newTestRouter(t)
		req := httptest.NewRequest(http.MethodPut, "/v1/currencies/ZZZ", bytes.NewBufferString(`{"name":"X"}`))
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("failed Validate", func(t *testing.T) {
		r, repo := newTestRouter(t)
		_ = repo.Create(domain.AppContext{}, domain.Currency{Code: "USD", Name: "US Dollar"})
		req := httptest.NewRequest(http.MethodPut, "/v1/currencies/USD", bytes.NewBufferString(`{"name":""}`))
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func TestDeleteCurrencyHandler(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		r, repo := newTestRouter(t)
		_ = repo.Create(domain.AppContext{}, domain.Currency{Code: "USD", Name: "US Dollar"})

		req := httptest.NewRequest(http.MethodDelete, "/v1/currencies/USD", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNoContent, rec.Code)
	})

	t.Run("not found", func(t *testing.T) {
		r, _ := newTestRouter(t)
		req := httptest.NewRequest(http.MethodDelete, "/v1/currencies/ZZZ", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}
