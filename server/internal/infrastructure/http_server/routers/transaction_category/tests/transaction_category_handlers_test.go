package transaction_category_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"jww/internal/adapter/schema"
	"jww/internal/domain"
	"jww/internal/infrastructure/http_server/kernel"
	categoryrouter "jww/internal/infrastructure/http_server/routers/transaction_category"
	"jww/internal/service/transaction_category"
	"jww/internal/service/unit_of_work"
)

type fakeRepository struct {
	byID map[uuid.UUID]domain.TransactionCategory
}

func newFakeRepository() *fakeRepository {
	return &fakeRepository{byID: map[uuid.UUID]domain.TransactionCategory{}}
}

func (f *fakeRepository) Create(_ domain.AppContext, c domain.TransactionCategory, _ ...unit_of_work.UnitOfWork) error {
	f.byID[c.ID] = c
	return nil
}

func (f *fakeRepository) GetByID(_ domain.AppContext, id uuid.UUID, _ ...unit_of_work.UnitOfWork) (*domain.TransactionCategory, error) {
	c, ok := f.byID[id]
	if !ok {
		return nil, nil
	}
	return &c, nil
}

func (f *fakeRepository) ListByType(_ domain.AppContext, categoryType domain.TransactionCategoryType, activeOnly bool, _ ...unit_of_work.UnitOfWork) ([]domain.TransactionCategory, error) {
	var out []domain.TransactionCategory
	for _, c := range f.byID {
		if c.Type != categoryType {
			continue
		}
		if activeOnly && !c.Active {
			continue
		}
		out = append(out, c)
	}
	return out, nil
}

func (f *fakeRepository) Update(_ domain.AppContext, c domain.TransactionCategory, _ ...unit_of_work.UnitOfWork) error {
	f.byID[c.ID] = c
	return nil
}

func (f *fakeRepository) Delete(_ domain.AppContext, c domain.TransactionCategory, _ ...unit_of_work.UnitOfWork) error {
	delete(f.byID, c.ID)
	return nil
}

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
		TransactionCategoryService: transaction_category.NewService(repo, fakeUoWManager{}),
	}
	r := chi.NewRouter()
	categoryrouter.Mount(r, r, deps)
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

func seedCategory(repo *fakeRepository, categoryType domain.TransactionCategoryType, active bool) domain.TransactionCategory {
	c := domain.TransactionCategory{ID: uuid.New(), Type: categoryType, Name: "Groceries", Active: active, SortOrder: 0}
	_ = repo.Create(domain.AppContext{}, c)
	return c
}

func TestCreateCategoryHandler(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		r, _ := newTestRouter(t)
		req := httptest.NewRequest(http.MethodPost, "/v1/transaction-categories", bytes.NewBufferString(`{"type":"expense","name":"Groceries","sort_order":0}`))
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		require.Equal(t, http.StatusCreated, rec.Code)
		var resp schema.TransactionCategoryResponse
		decodeData(t, rec, &resp)
		assert.Equal(t, "Groceries", resp.Name)
		assert.Equal(t, domain.TransactionCategoryTypeExpense, resp.Type)
	})

	t.Run("invalid type fails Validate", func(t *testing.T) {
		r, _ := newTestRouter(t)
		req := httptest.NewRequest(http.MethodPost, "/v1/transaction-categories", bytes.NewBufferString(`{"type":"bogus","name":"X"}`))
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func TestGetCategoryHandler(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		r, repo := newTestRouter(t)
		c := seedCategory(repo, domain.TransactionCategoryTypeExpense, true)
		req := httptest.NewRequest(http.MethodGet, "/v1/transaction-categories/"+c.ID.String(), nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("bad uuid", func(t *testing.T) {
		r, _ := newTestRouter(t)
		req := httptest.NewRequest(http.MethodGet, "/v1/transaction-categories/not-a-uuid", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("not found", func(t *testing.T) {
		r, _ := newTestRouter(t)
		req := httptest.NewRequest(http.MethodGet, "/v1/transaction-categories/"+uuid.NewString(), nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}

func TestListCategoriesHandler(t *testing.T) {
	t.Run("happy path filters by type and active_only", func(t *testing.T) {
		r, repo := newTestRouter(t)
		seedCategory(repo, domain.TransactionCategoryTypeExpense, true)
		seedCategory(repo, domain.TransactionCategoryTypeExpense, false)
		seedCategory(repo, domain.TransactionCategoryTypeIncome, true)

		req := httptest.NewRequest(http.MethodGet, "/v1/transaction-categories?type=expense&active_only=true", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		var resp []schema.TransactionCategoryResponse
		decodeData(t, rec, &resp)
		assert.Len(t, resp, 1)
	})

	t.Run("invalid type", func(t *testing.T) {
		r, _ := newTestRouter(t)
		req := httptest.NewRequest(http.MethodGet, "/v1/transaction-categories?type=bogus", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func TestUpdateCategoryHandler(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		r, repo := newTestRouter(t)
		c := seedCategory(repo, domain.TransactionCategoryTypeExpense, true)
		req := httptest.NewRequest(http.MethodPut, "/v1/transaction-categories/"+c.ID.String(), bytes.NewBufferString(`{"name":"Food","active":false,"sort_order":1}`))
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		var resp schema.TransactionCategoryResponse
		decodeData(t, rec, &resp)
		assert.Equal(t, "Food", resp.Name)
		assert.False(t, resp.Active)
	})

	t.Run("not found", func(t *testing.T) {
		r, _ := newTestRouter(t)
		req := httptest.NewRequest(http.MethodPut, "/v1/transaction-categories/"+uuid.NewString(), bytes.NewBufferString(`{"name":"X","sort_order":0}`))
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}

func TestDeleteCategoryHandler(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		r, repo := newTestRouter(t)
		c := seedCategory(repo, domain.TransactionCategoryTypeExpense, true)
		req := httptest.NewRequest(http.MethodDelete, "/v1/transaction-categories/"+c.ID.String(), nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNoContent, rec.Code)
	})

	t.Run("not found", func(t *testing.T) {
		r, _ := newTestRouter(t)
		req := httptest.NewRequest(http.MethodDelete, "/v1/transaction-categories/"+uuid.NewString(), nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}
