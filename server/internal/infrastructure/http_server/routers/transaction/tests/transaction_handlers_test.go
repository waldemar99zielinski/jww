package transaction_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"jww/internal/adapter/schema"
	"jww/internal/domain"
	"jww/internal/infrastructure/http_server/kernel"
	transactionrouter "jww/internal/infrastructure/http_server/routers/transaction"
	"jww/internal/service/currency"
	"jww/internal/service/transaction"
	"jww/internal/service/transaction_attachment"
	"jww/internal/service/transaction_category"
	"jww/internal/service/unit_of_work"
	"jww/internal/service/user"
	"jww/internal/service/workflow"
)

type fakeCurrencyRepo struct{ byCode map[string]domain.Currency }

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

type fakeCategoryRepo struct {
	byID map[uuid.UUID]domain.TransactionCategory
}

func (f *fakeCategoryRepo) Create(domain.AppContext, domain.TransactionCategory, ...unit_of_work.UnitOfWork) error {
	return nil
}
func (f *fakeCategoryRepo) GetByID(_ domain.AppContext, id uuid.UUID, _ ...unit_of_work.UnitOfWork) (*domain.TransactionCategory, error) {
	c, ok := f.byID[id]
	if !ok {
		return nil, nil
	}
	return &c, nil
}
func (f *fakeCategoryRepo) ListByType(domain.AppContext, domain.TransactionCategoryType, bool, ...unit_of_work.UnitOfWork) ([]domain.TransactionCategory, error) {
	return nil, nil
}
func (f *fakeCategoryRepo) Update(domain.AppContext, domain.TransactionCategory, ...unit_of_work.UnitOfWork) error {
	return nil
}
func (f *fakeCategoryRepo) Delete(domain.AppContext, domain.TransactionCategory, ...unit_of_work.UnitOfWork) error {
	return nil
}

type fakeUserRepo struct{ byID map[uuid.UUID]domain.User }

func (f *fakeUserRepo) Create(domain.AppContext, domain.User, ...unit_of_work.UnitOfWork) error {
	return nil
}
func (f *fakeUserRepo) GetByID(_ domain.AppContext, id uuid.UUID, _ ...unit_of_work.UnitOfWork) (*domain.User, error) {
	u, ok := f.byID[id]
	if !ok {
		return nil, nil
	}
	return &u, nil
}
func (f *fakeUserRepo) List(domain.AppContext, ...unit_of_work.UnitOfWork) ([]domain.User, error) {
	return nil, nil
}
func (f *fakeUserRepo) Update(domain.AppContext, domain.User, ...unit_of_work.UnitOfWork) error {
	return nil
}
func (f *fakeUserRepo) Delete(domain.AppContext, domain.User, ...unit_of_work.UnitOfWork) error {
	return nil
}

type fakePasswordHasher struct{}

func (fakePasswordHasher) Hash(password string) (string, error) { return "hashed:" + password, nil }
func (fakePasswordHasher) Verify(password, hash string) bool    { return "hashed:"+password == hash }

type fakeTransactionRepo struct {
	byID   map[int64]domain.Transaction
	nextID int64
}

func (f *fakeTransactionRepo) Create(
	_ domain.AppContext,
	transactionType domain.TransactionType,
	amount domain.Money,
	currency domain.Currency,
	category domain.TransactionCategory,
	user domain.User,
	source domain.TransactionSource,
	createdAt time.Time,
	name string,
	description string,
	_ ...unit_of_work.UnitOfWork,
) (*domain.Transaction, error) {
	f.nextID++
	t := domain.Transaction{
		ID: f.nextID, Type: transactionType, Amount: amount, Currency: currency,
		Category: category, User: user, Source: source, Name: name,
		Description: description, CreatedAt: createdAt,
	}
	f.byID[t.ID] = t
	return &t, nil
}

func (f *fakeTransactionRepo) GetByID(_ domain.AppContext, id int64, _ ...unit_of_work.UnitOfWork) (*domain.Transaction, error) {
	t, ok := f.byID[id]
	if !ok {
		return nil, nil
	}
	return &t, nil
}

func (f *fakeTransactionRepo) List(_ domain.AppContext, filter *transaction.Filter, _ ...unit_of_work.UnitOfWork) ([]domain.Transaction, error) {
	var out []domain.Transaction
	for _, t := range f.byID {
		if filter != nil && filter.UserID != nil && t.User.ID != *filter.UserID {
			continue
		}
		out = append(out, t)
	}
	return out, nil
}

func (f *fakeTransactionRepo) Update(_ domain.AppContext, t domain.Transaction, _ ...unit_of_work.UnitOfWork) error {
	f.byID[t.ID] = t
	return nil
}

func (f *fakeTransactionRepo) Delete(_ domain.AppContext, t domain.Transaction, _ ...unit_of_work.UnitOfWork) error {
	delete(f.byID, t.ID)
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

type testFixtures struct {
	currency          domain.Currency
	category          domain.TransactionCategory
	user              domain.User
	transactionRepo   *fakeTransactionRepo
	attachmentRepo    *fakeAttachmentRepo
	attachmentStorage *fakeAttachmentStorage
}

func newTestRouter(t *testing.T) (*chi.Mux, *testFixtures) {
	t.Helper()
	curr := domain.Currency{Code: "USD", Name: "US Dollar"}
	cat := domain.TransactionCategory{ID: uuid.New(), Type: domain.TransactionCategoryTypeExpense, Name: "Groceries", Active: true}
	usr := domain.User{ID: uuid.New(), Username: "wozi", PasswordHash: "hashed:x", DisplayName: "Wozi"}

	currencyRepo := &fakeCurrencyRepo{byCode: map[string]domain.Currency{curr.Code: curr}}
	categoryRepo := &fakeCategoryRepo{byID: map[uuid.UUID]domain.TransactionCategory{cat.ID: cat}}
	userRepo := &fakeUserRepo{byID: map[uuid.UUID]domain.User{usr.ID: usr}}
	transactionRepo := &fakeTransactionRepo{byID: map[int64]domain.Transaction{}}
	attachmentRepo := newFakeAttachmentRepo()
	attachmentStorage := newFakeAttachmentStorage()

	currencyService := currency.NewService(currencyRepo, fakeUoWManager{})
	categoryService := transaction_category.NewService(categoryRepo, fakeUoWManager{})
	userService := user.NewService(userRepo, fakePasswordHasher{}, fakeUoWManager{})
	transactionService := transaction.NewService(transactionRepo, fakeUoWManager{})
	attachmentService := transaction_attachment.NewService(attachmentRepo, attachmentStorage, nil)

	deps := &kernel.Deps{
		TransactionService:           transactionService,
		TransactionAttachmentService: attachmentService,
		TransactionWorkflow:          workflow.NewTransactionWorkflow(currencyService, categoryService, userService, transactionService, attachmentService, fakeUoWManager{}),
	}
	r := chi.NewRouter()
	transactionrouter.Mount(r, r, deps)
	return r, &testFixtures{
		currency:          curr,
		category:          cat,
		user:              usr,
		transactionRepo:   transactionRepo,
		attachmentRepo:    attachmentRepo,
		attachmentStorage: attachmentStorage,
	}
}

func decodeData(t *testing.T, rec *httptest.ResponseRecorder, v any) {
	t.Helper()
	var body struct {
		Data json.RawMessage `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.NoError(t, json.Unmarshal(body.Data, v))
}

func TestCreateTransactionHandler(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		r, fx := newTestRouter(t)
		body := `{"type":"expense","amount":"10.00","currency_code":"` + fx.currency.Code + `","category_id":"` + fx.category.ID.String() + `","user_id":"` + fx.user.ID.String() + `","source":"manual","name":"Groceries run","description":""}`
		req := httptest.NewRequest(http.MethodPost, "/v1/transactions", bytes.NewBufferString(body))
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		require.Equal(t, http.StatusCreated, rec.Code)
		var resp schema.TransactionResponse
		decodeData(t, rec, &resp)
		assert.Equal(t, "Groceries run", resp.Name)
		assert.Equal(t, domain.Money(1000), resp.Amount)
	})

	t.Run("unknown category is not found", func(t *testing.T) {
		r, fx := newTestRouter(t)
		body := `{"type":"expense","amount":"10.00","currency_code":"` + fx.currency.Code + `","category_id":"` + uuid.NewString() + `","user_id":"` + fx.user.ID.String() + `","source":"manual","name":"X"}`
		req := httptest.NewRequest(http.MethodPost, "/v1/transactions", bytes.NewBufferString(body))
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("failed Validate (zero amount)", func(t *testing.T) {
		r, fx := newTestRouter(t)
		body := `{"type":"expense","amount":"0","currency_code":"` + fx.currency.Code + `","category_id":"` + fx.category.ID.String() + `","user_id":"` + fx.user.ID.String() + `","source":"manual","name":"X"}`
		req := httptest.NewRequest(http.MethodPost, "/v1/transactions", bytes.NewBufferString(body))
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func TestGetTransactionHandler(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		r, fx := newTestRouter(t)
		tr, err := fx.transactionRepo.Create(domain.AppContext{}, domain.TransactionTypeExpense, domain.Money(500), fx.currency, fx.category, fx.user, domain.TransactionSourceManual, time.Now(), "Coffee", "")
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/v1/transactions/"+intToStr(tr.ID), nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("bad id", func(t *testing.T) {
		r, _ := newTestRouter(t)
		req := httptest.NewRequest(http.MethodGet, "/v1/transactions/not-a-number", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("not found", func(t *testing.T) {
		r, _ := newTestRouter(t)
		req := httptest.NewRequest(http.MethodGet, "/v1/transactions/999999", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}

func TestListTransactionsHandler(t *testing.T) {
	t.Run("no filters returns all transactions", func(t *testing.T) {
		r, fx := newTestRouter(t)
		_, err := fx.transactionRepo.Create(domain.AppContext{}, domain.TransactionTypeExpense, domain.Money(500), fx.currency, fx.category, fx.user, domain.TransactionSourceManual, time.Now().Add(-time.Hour), "Coffee", "")
		require.NoError(t, err)
		_, err = fx.transactionRepo.Create(domain.AppContext{}, domain.TransactionTypeExpense, domain.Money(800), fx.currency, fx.category, fx.user, domain.TransactionSourceManual, time.Now(), "Lunch", "")
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/v1/transactions", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		var resp []schema.TransactionResponse
		decodeData(t, rec, &resp)
		assert.Len(t, resp, 2)
	})

	t.Run("happy path with user_id filter", func(t *testing.T) {
		r, fx := newTestRouter(t)
		_, err := fx.transactionRepo.Create(domain.AppContext{}, domain.TransactionTypeExpense, domain.Money(500), fx.currency, fx.category, fx.user, domain.TransactionSourceManual, time.Now(), "Coffee", "")
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/v1/transactions?user_id="+fx.user.ID.String(), nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		var resp []schema.TransactionResponse
		decodeData(t, rec, &resp)
		assert.Len(t, resp, 1)
	})

	t.Run("invalid type query param", func(t *testing.T) {
		r, _ := newTestRouter(t)
		req := httptest.NewRequest(http.MethodGet, "/v1/transactions?type=bogus", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("invalid source query param", func(t *testing.T) {
		r, _ := newTestRouter(t)
		req := httptest.NewRequest(http.MethodGet, "/v1/transactions?source=bogus", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func TestDeleteTransactionHandler(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		r, fx := newTestRouter(t)
		tr, err := fx.transactionRepo.Create(domain.AppContext{}, domain.TransactionTypeExpense, domain.Money(500), fx.currency, fx.category, fx.user, domain.TransactionSourceManual, time.Now(), "Coffee", "")
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodDelete, "/v1/transactions/"+intToStr(tr.ID), nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNoContent, rec.Code)
	})

	t.Run("not found", func(t *testing.T) {
		r, _ := newTestRouter(t)
		req := httptest.NewRequest(http.MethodDelete, "/v1/transactions/999999", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}

func intToStr(id int64) string {
	return json.Number(string(rune('0' + id))).String() // placeholder replaced below
}
