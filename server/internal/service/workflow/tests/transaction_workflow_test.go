package workflow_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"jww/internal/domain"
	"jww/internal/service/currency"
	"jww/internal/service/transaction"
	"jww/internal/service/transactioncategory"
	"jww/internal/service/user"
	"jww/internal/service/workflow"
)

type transactionWorkflowFixture struct {
	currencyRepo    *fakeCurrencyRepo
	categoryRepo    *fakeCategoryRepo
	userRepo        *fakeUserRepo
	transactionRepo *fakeTransactionRepo
	uowManager      *fakeUoWManager
	workflow        *workflow.TransactionWorkflow
}

func newTransactionWorkflowFixture(currencies []domain.Currency, categories []domain.TransactionCategory, users []domain.User) *transactionWorkflowFixture {
	f := &transactionWorkflowFixture{
		currencyRepo:    newFakeCurrencyRepo(currencies...),
		categoryRepo:    newFakeCategoryRepo(categories...),
		userRepo:        newFakeUserRepo(users...),
		transactionRepo: &fakeTransactionRepo{},
		uowManager:      &fakeUoWManager{},
	}
	f.workflow = workflow.NewTransactionWorkflow(
		currency.NewService(f.currencyRepo, f.uowManager),
		transactioncategory.NewService(f.categoryRepo, f.uowManager),
		user.NewService(f.userRepo, fakePasswordHasher{}, f.uowManager),
		transaction.NewService(f.transactionRepo, f.uowManager),
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

		created, err := f.workflow.Create(newAppContext(), domain.TransactionTypeExpense, domain.Money(1500), "USD", category.ID, user.ID, domain.TransactionSourceManual, "Weekly shop", "")

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

		_, err := f.workflow.Create(newAppContext(), domain.TransactionTypeExpense, domain.Money(1500), "USD", category.ID, user.ID, domain.TransactionSourceManual, "Weekly shop", "")

		require.Error(t, err)
		assert.True(t, f.uowManager.last().restored)
		assert.Zero(t, f.transactionRepo.createCalled)
	})

	t.Run("unknown category rolls back before resolving user", func(t *testing.T) {
		user := julcia()
		f := newTransactionWorkflowFixture([]domain.Currency{usd}, nil, []domain.User{user})

		_, err := f.workflow.Create(newAppContext(), domain.TransactionTypeExpense, domain.Money(1500), "USD", uuid.New(), user.ID, domain.TransactionSourceManual, "Weekly shop", "")

		require.Error(t, err)
		assert.True(t, f.uowManager.last().restored)
		assert.Zero(t, f.transactionRepo.createCalled)
	})

	t.Run("unknown user rolls back", func(t *testing.T) {
		category := groceries()
		f := newTransactionWorkflowFixture([]domain.Currency{usd}, []domain.TransactionCategory{category}, nil)

		_, err := f.workflow.Create(newAppContext(), domain.TransactionTypeExpense, domain.Money(1500), "USD", category.ID, uuid.New(), domain.TransactionSourceManual, "Weekly shop", "")

		require.Error(t, err)
		assert.True(t, f.uowManager.last().restored)
		assert.Zero(t, f.transactionRepo.createCalled)
	})

	t.Run("mismatched category/transaction type rolls back (entity-level validation surfaces through the workflow)", func(t *testing.T) {
		// category is an expense category, but the request claims income.
		category := groceries()
		user := julcia()
		f := newTransactionWorkflowFixture([]domain.Currency{usd}, []domain.TransactionCategory{category}, []domain.User{user})

		_, err := f.workflow.Create(newAppContext(), domain.TransactionTypeIncome, domain.Money(1500), "USD", category.ID, user.ID, domain.TransactionSourceManual, "Salary", "")

		require.Error(t, err)
		assert.True(t, f.uowManager.last().restored)
	})
}

func TestTransactionWorkflow_List(t *testing.T) {
	t.Run("success with no filters opens and commits a transaction", func(t *testing.T) {
		f := newTransactionWorkflowFixture(nil, nil, nil)
		f.transactionRepo.listResult = []domain.Transaction{{ID: 1}}

		got, err := f.workflow.List(newAppContext(), workflow.TransactionListFilter{})

		require.NoError(t, err)
		assert.Len(t, got, 1)
		assert.Equal(t, 1, f.transactionRepo.listCalled)
		assert.True(t, f.uowManager.last().finished)
	})

	t.Run("unknown user filter rolls back and never calls List", func(t *testing.T) {
		f := newTransactionWorkflowFixture(nil, nil, nil)
		missing := uuid.New()

		_, err := f.workflow.List(newAppContext(), workflow.TransactionListFilter{UserID: &missing})

		require.Error(t, err)
		assert.True(t, f.uowManager.last().restored)
		assert.Zero(t, f.transactionRepo.listCalled)
	})

	t.Run("filter that fails its own Validate (category/type mismatch) rolls back", func(t *testing.T) {
		category := groceries() // expense
		f := newTransactionWorkflowFixture(nil, []domain.TransactionCategory{category}, nil)
		incomeType := domain.TransactionTypeIncome

		_, err := f.workflow.List(newAppContext(), workflow.TransactionListFilter{CategoryID: &category.ID, Type: &incomeType})

		require.Error(t, err)
		assert.True(t, f.uowManager.last().restored)
		assert.Zero(t, f.transactionRepo.listCalled)
	})
}
