package transaction_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"jww/internal/adapter/schema"
	jwtadapter "jww/internal/adapter/security/jwt"
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
	otherCurrency     domain.Currency
	incomeCategory    domain.TransactionCategory
	otherUser         domain.User
	currencyRepo      *fakeCurrencyRepo
	categoryRepo      *fakeCategoryRepo
	userRepo          *fakeUserRepo
	transactionRepo   *fakeTransactionRepo
	attachmentRepo    *fakeAttachmentRepo
	attachmentStorage *fakeAttachmentStorage
}

func newTestRouter(t *testing.T) (*chi.Mux, *testFixtures) {
	t.Helper()
	curr := domain.Currency{Code: "USD", Name: "US Dollar"}
	otherCurr := domain.Currency{Code: "EUR", Name: "Euro"}
	cat := domain.TransactionCategory{ID: uuid.New(), Type: domain.TransactionCategoryTypeExpense, Name: "Groceries", Active: true}
	incomeCat := domain.TransactionCategory{ID: uuid.New(), Type: domain.TransactionCategoryTypeIncome, Name: "Salary", Active: true}
	usr := domain.User{ID: uuid.New(), Username: "wozi", PasswordHash: "hashed:x", DisplayName: "Wozi"}
	otherUsr := domain.User{ID: uuid.New(), Username: "partner", PasswordHash: "hashed:y", DisplayName: "Partner"}

	currencyRepo := &fakeCurrencyRepo{byCode: map[string]domain.Currency{curr.Code: curr, otherCurr.Code: otherCurr}}
	categoryRepo := &fakeCategoryRepo{byID: map[uuid.UUID]domain.TransactionCategory{cat.ID: cat, incomeCat.ID: incomeCat}}
	userRepo := &fakeUserRepo{byID: map[uuid.UUID]domain.User{usr.ID: usr, otherUsr.ID: otherUsr}}
	transactionRepo := &fakeTransactionRepo{byID: map[int64]domain.Transaction{}}
	attachmentRepo := newFakeAttachmentRepo()
	attachmentStorage := newFakeAttachmentStorage()

	currencyService := currency.NewService(currencyRepo, fakeUoWManager{})
	categoryService := transaction_category.NewService(categoryRepo, fakeUoWManager{})
	userService := user.NewService(userRepo, fakePasswordHasher{}, fakeUoWManager{})
	transactionService := transaction.NewService(transactionRepo, fakeUoWManager{})
	grants := jwtadapter.NewAttachmentGrantService(jwtadapter.Config{
		SecretKey: "transaction-handler-test-secret", Algorithm: "HS256", AccessTokenExpireMinutes: 60,
	}, 5*time.Minute)
	attachmentService := transaction_attachment.NewService(attachmentRepo, attachmentStorage, nil, grants)

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
		otherCurrency:     otherCurr,
		incomeCategory:    incomeCat,
		otherUser:         otherUsr,
		currencyRepo:      currencyRepo,
		categoryRepo:      categoryRepo,
		userRepo:          userRepo,
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

	t.Run("accepts an explicit created_at", func(t *testing.T) {
		r, fx := newTestRouter(t)
		const when = "2024-03-15T12:30:00Z"
		body := `{"type":"expense","amount":"10.00","currency_code":"` + fx.currency.Code + `","category_id":"` + fx.category.ID.String() + `","user_id":"` + fx.user.ID.String() + `","source":"manual","name":"Dated","created_at":"` + when + `"}`
		req := httptest.NewRequest(http.MethodPost, "/v1/transactions", bytes.NewBufferString(body))
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		require.Equal(t, http.StatusCreated, rec.Code)
		var resp schema.TransactionResponse
		decodeData(t, rec, &resp)
		want, err := time.Parse(time.RFC3339, when)
		require.NoError(t, err)
		assert.WithinDuration(t, want, resp.CreatedAt, time.Second)
	})

	t.Run("omitting created_at defaults to now", func(t *testing.T) {
		r, fx := newTestRouter(t)
		body := `{"type":"expense","amount":"10.00","currency_code":"` + fx.currency.Code + `","category_id":"` + fx.category.ID.String() + `","user_id":"` + fx.user.ID.String() + `","source":"manual","name":"Undated"}`
		req := httptest.NewRequest(http.MethodPost, "/v1/transactions", bytes.NewBufferString(body))
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		require.Equal(t, http.StatusCreated, rec.Code)
		var resp schema.TransactionResponse
		decodeData(t, rec, &resp)
		assert.WithinDuration(t, time.Now(), resp.CreatedAt, time.Minute)
	})

	t.Run("zero or malformed created_at is bad request", func(t *testing.T) {
		r, fx := newTestRouter(t)
		base := `{"type":"expense","amount":"10.00","currency_code":"` + fx.currency.Code + `","category_id":"` + fx.category.ID.String() + `","user_id":"` + fx.user.ID.String() + `","source":"manual","name":"X"`
		assert.Equal(t, http.StatusBadRequest, postCreate(t, r, base+`,"created_at":"0001-01-01T00:00:00Z"}`).Code)
		assert.Equal(t, http.StatusBadRequest, postCreate(t, r, base+`,"created_at":"not-a-date"}`).Code)
	})
}

func TestGetTransactionHandler(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		r, fx := newTestRouter(t)
		tr, err := fx.transactionRepo.Create(domain.AppContext{}, domain.TransactionTypeExpense, domain.Money(500), fx.currency, fx.category, fx.user, domain.TransactionSourceManual, time.Now(), "Coffee", "")
		require.NoError(t, err)
		previewable := fx.attachmentRepo.seedDraft()
		previewable.TransactionID = &tr.ID
		fx.attachmentRepo.byID[previewable.ID] = previewable
		downloadOnly := fx.attachmentRepo.seedDraft()
		downloadOnly.TransactionID = &tr.ID
		downloadOnly.MediaType = "text/html"
		fx.attachmentRepo.byID[downloadOnly.ID] = downloadOnly

		req := httptest.NewRequest(http.MethodGet, "/v1/transactions/"+intToStr(tr.ID), nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)
		var response schema.TransactionDetailsResponse
		decodeData(t, rec, &response)
		require.Len(t, response.Attachments, 2)
		for _, attachment := range response.Attachments {
			switch attachment.ID {
			case previewable.ID:
				require.NotNil(t, attachment.PreviewGrant)
				assert.NotEmpty(t, *attachment.PreviewGrant)
			case downloadOnly.ID:
				assert.Nil(t, attachment.PreviewGrant)
			default:
				t.Fatalf("unexpected attachment %s", attachment.ID)
			}
		}
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
		first, err := fx.transactionRepo.Create(domain.AppContext{}, domain.TransactionTypeExpense, domain.Money(500), fx.currency, fx.category, fx.user, domain.TransactionSourceManual, time.Now().Add(-time.Hour), "Coffee", "")
		require.NoError(t, err)
		_, err = fx.transactionRepo.Create(domain.AppContext{}, domain.TransactionTypeExpense, domain.Money(800), fx.currency, fx.category, fx.user, domain.TransactionSourceManual, time.Now(), "Lunch", "")
		require.NoError(t, err)
		attachment := fx.attachmentRepo.seedDraft()
		attachment.TransactionID = &first.ID
		fx.attachmentRepo.byID[attachment.ID] = attachment

		req := httptest.NewRequest(http.MethodGet, "/v1/transactions", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		assert.NotContains(t, rec.Body.String(), "preview_grant", "list responses must not mint or expose grants")
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

func seedTransactionRecord(t *testing.T, fx *testFixtures) domain.Transaction {
	t.Helper()
	tr, err := fx.transactionRepo.Create(domain.AppContext{}, domain.TransactionTypeExpense, domain.Money(500), fx.currency, fx.category, fx.user, domain.TransactionSourceManual, time.Now(), "Coffee", "note")
	require.NoError(t, err)
	return *tr
}

func patchTransaction(t *testing.T, r http.Handler, id int64, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPatch, "/v1/transactions/"+intToStr(id), bytes.NewBufferString(body))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func postCreate(t *testing.T, r http.Handler, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/v1/transactions", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func TestUpdateTransactionHandler(t *testing.T) {
	t.Run("applies every provided field and returns the detail shape", func(t *testing.T) {
		r, fx := newTestRouter(t)
		tr := seedTransactionRecord(t, fx)
		previewable := fx.attachmentRepo.seedDraft()
		previewable.TransactionID = &tr.ID
		fx.attachmentRepo.byID[previewable.ID] = previewable

		body := `{"type":"income","amount":"12.34","currency_code":"EUR","category_id":"` + fx.incomeCategory.ID.String() +
			`","user_id":"` + fx.otherUser.ID.String() + `","source":"recurring","name":"Salary","description":"march"}`
		rec := patchTransaction(t, r, tr.ID, body)

		require.Equal(t, http.StatusOK, rec.Code)
		var resp schema.TransactionDetailsResponse
		decodeData(t, rec, &resp)
		assert.Equal(t, domain.TransactionTypeIncome, resp.Type)
		assert.Equal(t, domain.Money(1234), resp.Amount, "a bare JSON number parses as money")
		assert.Equal(t, "EUR", resp.CurrencyCode)
		assert.Equal(t, fx.incomeCategory.ID, resp.CategoryID)
		assert.Equal(t, fx.otherUser.ID, resp.UserID)
		assert.Equal(t, domain.TransactionSourceRecurring, resp.Source)
		assert.Equal(t, "Salary", resp.Name)
		assert.Equal(t, "march", resp.Description)
		assert.WithinDuration(t, tr.CreatedAt, resp.CreatedAt, time.Millisecond, "created_at is immutable")
		require.Len(t, resp.Attachments, 1)
		require.NotNil(t, resp.Attachments[0].PreviewGrant)

		stored := fx.transactionRepo.byID[tr.ID]
		assert.Equal(t, domain.TransactionTypeIncome, stored.Type)
		assert.Equal(t, domain.Money(1234), stored.Amount)
		assert.Equal(t, "EUR", stored.Currency.Code)
		assert.Equal(t, fx.incomeCategory.ID, stored.Category.ID)
		assert.Equal(t, fx.otherUser.ID, stored.User.ID)
		assert.Equal(t, domain.TransactionSourceRecurring, stored.Source)
		assert.Equal(t, "Salary", stored.Name)
		assert.Equal(t, "march", stored.Description)
	})

	t.Run("only provided fields change", func(t *testing.T) {
		r, fx := newTestRouter(t)
		tr := seedTransactionRecord(t, fx)

		rec := patchTransaction(t, r, tr.ID, `{"description":"updated"}`)

		require.Equal(t, http.StatusOK, rec.Code)
		var resp schema.TransactionDetailsResponse
		decodeData(t, rec, &resp)
		assert.Equal(t, "updated", resp.Description)
		assert.Equal(t, "Coffee", resp.Name, "unpatched fields keep their value")
		assert.Equal(t, domain.Money(500), resp.Amount)
		assert.Equal(t, "USD", resp.CurrencyCode)
		assert.Equal(t, fx.category.ID, resp.CategoryID)
		assert.Equal(t, fx.user.ID, resp.UserID)
	})

	t.Run("an explicit empty string clears the description", func(t *testing.T) {
		r, fx := newTestRouter(t)
		tr := seedTransactionRecord(t, fx)

		rec := patchTransaction(t, r, tr.ID, `{"description":""}`)

		require.Equal(t, http.StatusOK, rec.Code)
		var resp schema.TransactionDetailsResponse
		decodeData(t, rec, &resp)
		assert.Equal(t, "", resp.Description)
		assert.Equal(t, "", fx.transactionRepo.byID[tr.ID].Description)
	})

	t.Run("empty patch is a no-op", func(t *testing.T) {
		r, fx := newTestRouter(t)
		tr := seedTransactionRecord(t, fx)

		rec := patchTransaction(t, r, tr.ID, `{}`)

		require.Equal(t, http.StatusOK, rec.Code)
		var resp schema.TransactionDetailsResponse
		decodeData(t, rec, &resp)
		assert.Equal(t, "Coffee", resp.Name)
		assert.Equal(t, domain.Money(500), resp.Amount)
	})

	t.Run("unknown category is not found", func(t *testing.T) {
		r, fx := newTestRouter(t)
		tr := seedTransactionRecord(t, fx)
		rec := patchTransaction(t, r, tr.ID, `{"category_id":"`+uuid.NewString()+`"}`)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("unknown user is not found", func(t *testing.T) {
		r, fx := newTestRouter(t)
		tr := seedTransactionRecord(t, fx)
		rec := patchTransaction(t, r, tr.ID, `{"user_id":"`+uuid.NewString()+`"}`)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("unknown currency is not found", func(t *testing.T) {
		r, fx := newTestRouter(t)
		tr := seedTransactionRecord(t, fx)
		rec := patchTransaction(t, r, tr.ID, `{"currency_code":"GBP"}`)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("invalid type is bad request", func(t *testing.T) {
		r, fx := newTestRouter(t)
		tr := seedTransactionRecord(t, fx)
		rec := patchTransaction(t, r, tr.ID, `{"type":"bogus"}`)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("zero amount is bad request", func(t *testing.T) {
		r, fx := newTestRouter(t)
		tr := seedTransactionRecord(t, fx)
		rec := patchTransaction(t, r, tr.ID, `{"amount":0}`)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("blank name is bad request", func(t *testing.T) {
		r, fx := newTestRouter(t)
		tr := seedTransactionRecord(t, fx)
		rec := patchTransaction(t, r, tr.ID, `{"name":""}`)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("type no longer matching the category is bad request", func(t *testing.T) {
		r, fx := newTestRouter(t)
		tr := seedTransactionRecord(t, fx)
		rec := patchTransaction(t, r, tr.ID, `{"type":"income"}`)
		assert.Equal(t, http.StatusBadRequest, rec.Code, "switching type without a matching category must be rejected")
	})

	t.Run("updates the transaction date", func(t *testing.T) {
		r, fx := newTestRouter(t)
		tr := seedTransactionRecord(t, fx)
		const when = "2024-03-15T12:30:00Z"

		rec := patchTransaction(t, r, tr.ID, `{"created_at":"`+when+`"}`)

		require.Equal(t, http.StatusOK, rec.Code)
		var resp schema.TransactionDetailsResponse
		decodeData(t, rec, &resp)
		want, err := time.Parse(time.RFC3339, when)
		require.NoError(t, err)
		assert.WithinDuration(t, want, resp.CreatedAt, time.Second)
		assert.WithinDuration(t, want, fx.transactionRepo.byID[tr.ID].CreatedAt, time.Second)
	})

	t.Run("malformed or zero date is bad request", func(t *testing.T) {
		r, fx := newTestRouter(t)
		tr := seedTransactionRecord(t, fx)
		assert.Equal(t, http.StatusBadRequest, patchTransaction(t, r, tr.ID, `{"created_at":"not-a-date"}`).Code)
		assert.Equal(t, http.StatusBadRequest, patchTransaction(t, r, tr.ID, `{"created_at":"0001-01-01T00:00:00Z"}`).Code)
	})

	t.Run("unknown fields are rejected", func(t *testing.T) {
		r, fx := newTestRouter(t)
		tr := seedTransactionRecord(t, fx)
		assert.Equal(t, http.StatusBadRequest, patchTransaction(t, r, tr.ID, `{"attachment_ids":[]}`).Code)
		assert.Equal(t, http.StatusBadRequest, patchTransaction(t, r, tr.ID, `{"id":7}`).Code)
	})

	t.Run("not found", func(t *testing.T) {
		r, _ := newTestRouter(t)
		rec := patchTransaction(t, r, 999999, `{"name":"X"}`)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("bad id", func(t *testing.T) {
		r, _ := newTestRouter(t)
		req := httptest.NewRequest(http.MethodPatch, "/v1/transactions/not-a-number", bytes.NewBufferString(`{"name":"X"}`))
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func intToStr(id int64) string {
	return strconv.FormatInt(id, 10)
}
