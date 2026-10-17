package workflow_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"jww/internal/domain"
	"jww/internal/service/currency"
	"jww/internal/service/transaction"
	"jww/internal/service/transaction_attachment"
	transactioncategory "jww/internal/service/transaction_category"
	"jww/internal/service/user"
	"jww/internal/service/workflow"
)

type transactionWorkflowFixture struct {
	currencyRepo      *fakeCurrencyRepo
	categoryRepo      *fakeCategoryRepo
	userRepo          *fakeUserRepo
	transactionRepo   *fakeTransactionRepo
	attachmentRepo    *fakeAttachmentRepo
	attachmentStorage *fakeAttachmentStorage
	uowManager        *fakeUoWManager
	workflow          *workflow.TransactionWorkflow
}

func newTransactionWorkflowFixture(currencies []domain.Currency, categories []domain.TransactionCategory, users []domain.User) *transactionWorkflowFixture {
	f := &transactionWorkflowFixture{
		currencyRepo:      newFakeCurrencyRepo(currencies...),
		categoryRepo:      newFakeCategoryRepo(categories...),
		userRepo:          newFakeUserRepo(users...),
		transactionRepo:   &fakeTransactionRepo{},
		attachmentRepo:    newFakeAttachmentRepo(),
		attachmentStorage: newFakeAttachmentStorage(),
		uowManager:        &fakeUoWManager{},
	}
	attachmentService := transaction_attachment.NewService(f.attachmentRepo, f.attachmentStorage, nil)
	f.workflow = workflow.NewTransactionWorkflow(
		currency.NewService(f.currencyRepo, f.uowManager),
		transactioncategory.NewService(f.categoryRepo, f.uowManager),
		user.NewService(f.userRepo, fakePasswordHasher{}, f.uowManager),
		transaction.NewService(f.transactionRepo, f.uowManager),
		attachmentService,
		f.uowManager,
	)
	return f
}

func groceries() domain.TransactionCategory {
	return domain.TransactionCategory{ID: uuid.New(), Type: domain.TransactionCategoryTypeExpense, Name: "Groceries", Active: true}
}

func julcia() domain.User {
	return domain.User{ID: uuid.New(), Username: "julcia", PasswordHash: "x", DisplayName: "Julcia"}
}

func TestTransactionWorkflow_Create(t *testing.T) {
	t.Run("success resolves currency, category, and user before creating", func(t *testing.T) {
		category := groceries()
		user := julcia()
		f := newTransactionWorkflowFixture([]domain.Currency{usd}, []domain.TransactionCategory{category}, []domain.User{user})

		created, err := f.workflow.Create(newAppContext(), domain.TransactionTypeExpense, domain.Money(1500), "USD", category.ID, user.ID, domain.TransactionSourceManual, "Weekly shop", "", nil)

		require.NoError(t, err)
		assert.Equal(t, domain.Money(1500), created.Amount)
		assert.Equal(t, "USD", created.Currency.Code)
		assert.Equal(t, category.ID, created.Category.ID)
		assert.Equal(t, user.ID, created.User.ID)
		assert.Equal(t, 1, f.transactionRepo.createCalled)
		assert.True(t, f.uowManager.last().finished)
		assert.False(t, f.uowManager.last().restored)
	})

	t.Run("unknown currency rolls back before resolving category or user", func(t *testing.T) {
		category := groceries()
		user := julcia()
		f := newTransactionWorkflowFixture(nil, []domain.TransactionCategory{category}, []domain.User{user})

		_, err := f.workflow.Create(newAppContext(), domain.TransactionTypeExpense, domain.Money(1500), "USD", category.ID, user.ID, domain.TransactionSourceManual, "Weekly shop", "", nil)

		require.Error(t, err)
		assert.True(t, f.uowManager.last().restored)
		assert.Zero(t, f.transactionRepo.createCalled)
	})

	t.Run("unknown category rolls back before resolving user", func(t *testing.T) {
		user := julcia()
		f := newTransactionWorkflowFixture([]domain.Currency{usd}, nil, []domain.User{user})

		_, err := f.workflow.Create(newAppContext(), domain.TransactionTypeExpense, domain.Money(1500), "USD", uuid.New(), user.ID, domain.TransactionSourceManual, "Weekly shop", "", nil)

		require.Error(t, err)
		assert.True(t, f.uowManager.last().restored)
		assert.Zero(t, f.transactionRepo.createCalled)
	})

	t.Run("unknown user rolls back", func(t *testing.T) {
		category := groceries()
		f := newTransactionWorkflowFixture([]domain.Currency{usd}, []domain.TransactionCategory{category}, nil)

		_, err := f.workflow.Create(newAppContext(), domain.TransactionTypeExpense, domain.Money(1500), "USD", category.ID, uuid.New(), domain.TransactionSourceManual, "Weekly shop", "", nil)

		require.Error(t, err)
		assert.True(t, f.uowManager.last().restored)
		assert.Zero(t, f.transactionRepo.createCalled)
	})

	t.Run("mismatched category/transaction type rolls back (entity-level validation surfaces through the workflow)", func(t *testing.T) {
		// category is an expense category, but the request claims income.
		category := groceries()
		user := julcia()
		f := newTransactionWorkflowFixture([]domain.Currency{usd}, []domain.TransactionCategory{category}, []domain.User{user})

		_, err := f.workflow.Create(newAppContext(), domain.TransactionTypeIncome, domain.Money(1500), "USD", category.ID, user.ID, domain.TransactionSourceManual, "Salary", "", nil)

		require.Error(t, err)
		assert.True(t, f.uowManager.last().restored)
	})
}

func TestTransactionWorkflow_List(t *testing.T) {
	t.Run("success with no filters does not open a transaction", func(t *testing.T) {
		f := newTransactionWorkflowFixture(nil, nil, nil)
		f.transactionRepo.listResult = []domain.Transaction{{ID: 1}}

		got, err := f.workflow.List(newAppContext(), transaction.Filter{})

		require.NoError(t, err)
		assert.Len(t, got, 1)
		assert.Equal(t, 1, f.transactionRepo.listCalled)
		assert.Nil(t, f.uowManager.last())
	})

	t.Run("passes explicit related-resource fields directly to the repository", func(t *testing.T) {
		currency := domain.Currency{Code: "USD", Name: "US Dollar"}
		category := groceries()
		user := julcia()
		f := newTransactionWorkflowFixture([]domain.Currency{currency}, []domain.TransactionCategory{category}, []domain.User{user})

		_, err := f.workflow.List(newAppContext(), transaction.Filter{
			UserID:       &user.ID,
			CategoryID:   &category.ID,
			CurrencyCode: "USD",
		})

		require.NoError(t, err)
		require.NotNil(t, f.transactionRepo.listFilter)
		assert.Equal(t, user.ID, *f.transactionRepo.listFilter.UserID)
		assert.Equal(t, category.ID, *f.transactionRepo.listFilter.CategoryID)
		assert.Equal(t, "USD", f.transactionRepo.listFilter.CurrencyCode)
		assert.Nil(t, f.uowManager.last())
	})

	t.Run("unknown user filter never lists", func(t *testing.T) {
		f := newTransactionWorkflowFixture(nil, nil, nil)
		missing := uuid.New()

		_, err := f.workflow.List(newAppContext(), transaction.Filter{UserID: &missing})

		require.Error(t, err)
		assert.Nil(t, f.uowManager.last())
		assert.Zero(t, f.transactionRepo.listCalled)
	})

	t.Run("category type mismatch never lists", func(t *testing.T) {
		category := groceries()
		f := newTransactionWorkflowFixture(nil, []domain.TransactionCategory{category}, nil)
		incomeType := domain.TransactionTypeIncome

		_, err := f.workflow.List(newAppContext(), transaction.Filter{CategoryID: &category.ID, Type: &incomeType})

		require.Error(t, err)
		assert.Nil(t, f.uowManager.last())
		assert.Zero(t, f.transactionRepo.listCalled)
	})

	t.Run("invalid date range fails before listing", func(t *testing.T) {
		f := newTransactionWorkflowFixture(nil, nil, nil)
		from := time.Now().UTC()
		to := from.Add(-time.Hour)

		_, err := f.workflow.List(newAppContext(), transaction.Filter{CreatedFrom: &from, CreatedTo: &to})

		require.Error(t, err)
		assert.Nil(t, f.uowManager.last())
		assert.Zero(t, f.transactionRepo.listCalled)
	})
}
