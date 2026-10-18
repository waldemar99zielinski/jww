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

		created, err := f.workflow.Create(newAppContext(), domain.TransactionTypeExpense, domain.Money(1500), "USD", category.ID, user.ID, domain.TransactionSourceManual, "Weekly shop", "", nil, nil)

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

		_, err := f.workflow.Create(newAppContext(), domain.TransactionTypeExpense, domain.Money(1500), "USD", category.ID, user.ID, domain.TransactionSourceManual, "Weekly shop", "", nil, nil)

		require.Error(t, err)
		assert.True(t, f.uowManager.last().restored)
		assert.Zero(t, f.transactionRepo.createCalled)
	})

	t.Run("unknown category rolls back before resolving user", func(t *testing.T) {
		user := julcia()
		f := newTransactionWorkflowFixture([]domain.Currency{usd}, nil, []domain.User{user})

		_, err := f.workflow.Create(newAppContext(), domain.TransactionTypeExpense, domain.Money(1500), "USD", uuid.New(), user.ID, domain.TransactionSourceManual, "Weekly shop", "", nil, nil)

		require.Error(t, err)
		assert.True(t, f.uowManager.last().restored)
		assert.Zero(t, f.transactionRepo.createCalled)
	})

	t.Run("unknown user rolls back", func(t *testing.T) {
		category := groceries()
		f := newTransactionWorkflowFixture([]domain.Currency{usd}, []domain.TransactionCategory{category}, nil)

		_, err := f.workflow.Create(newAppContext(), domain.TransactionTypeExpense, domain.Money(1500), "USD", category.ID, uuid.New(), domain.TransactionSourceManual, "Weekly shop", "", nil, nil)

		require.Error(t, err)
		assert.True(t, f.uowManager.last().restored)
		assert.Zero(t, f.transactionRepo.createCalled)
	})

	t.Run("mismatched category/transaction type rolls back (entity-level validation surfaces through the workflow)", func(t *testing.T) {
		// category is an expense category, but the request claims income.
		category := groceries()
		user := julcia()
		f := newTransactionWorkflowFixture([]domain.Currency{usd}, []domain.TransactionCategory{category}, []domain.User{user})

		_, err := f.workflow.Create(newAppContext(), domain.TransactionTypeIncome, domain.Money(1500), "USD", category.ID, user.ID, domain.TransactionSourceManual, "Salary", "", nil, nil)

		require.Error(t, err)
		assert.True(t, f.uowManager.last().restored)
	})
}

// ptr returns a pointer to v, for building optional patch fields.
func ptr[T any](v T) *T { return &v }

type updateFixture struct {
	*transactionWorkflowFixture
	expenseCategory domain.TransactionCategory
	incomeCategory  domain.TransactionCategory
	euro            domain.Currency
}

func newUpdateFixture() *updateFixture {
	expenseCategory := groceries()
	incomeCategory := domain.TransactionCategory{ID: uuid.New(), Type: domain.TransactionCategoryTypeIncome, Name: "Salary", Active: true}
	user := julcia()
	euro := domain.Currency{Code: "EUR", Name: "Euro"}
	f := newTransactionWorkflowFixture([]domain.Currency{usd, euro}, []domain.TransactionCategory{expenseCategory, incomeCategory}, []domain.User{user})
	f.transactionRepo.getResult = &domain.Transaction{
		ID: 7, Type: domain.TransactionTypeExpense, Amount: domain.Money(500),
		Currency: usd, Category: expenseCategory, User: user,
		Source: domain.TransactionSourceManual, Name: "Coffee", Description: "note",
	}
	return &updateFixture{transactionWorkflowFixture: f, expenseCategory: expenseCategory, incomeCategory: incomeCategory, euro: euro}
}

func TestTransactionWorkflow_Update(t *testing.T) {
	t.Run("merges provided fields and leaves the rest untouched", func(t *testing.T) {
		f := newUpdateFixture()

		result, err := f.workflow.Update(newAppContext(), 7, transaction.Patch{Amount: ptr(domain.Money(999)), Name: ptr("Espresso")})

		require.NoError(t, err)
		assert.Equal(t, domain.Money(999), result.Transaction.Amount)
		assert.Equal(t, "Espresso", result.Transaction.Name)
		assert.Equal(t, "note", result.Transaction.Description, "unpatched fields keep their value")
		require.Len(t, f.transactionRepo.updated, 1)
		stored := f.transactionRepo.updated[0]
		assert.Equal(t, domain.Money(999), stored.Amount)
		assert.Equal(t, "Espresso", stored.Name)
		assert.Equal(t, "note", stored.Description)
		assert.Equal(t, usd.Code, stored.Currency.Code)
	})

	t.Run("resolves a changed currency, category, user, and source before writing", func(t *testing.T) {
		f := newUpdateFixture()
		otherUser := julcia()
		f.userRepo.byID[otherUser.ID] = otherUser

		result, err := f.workflow.Update(newAppContext(), 7, transaction.Patch{
			Type:         ptr(domain.TransactionTypeIncome),
			CurrencyCode: ptr(f.euro.Code),
			CategoryID:   ptr(f.incomeCategory.ID),
			UserID:       ptr(otherUser.ID),
			Source:       ptr(domain.TransactionSourceRecurring),
		})

		require.NoError(t, err)
		require.Len(t, f.transactionRepo.updated, 1)
		stored := f.transactionRepo.updated[0]
		assert.Equal(t, domain.TransactionTypeIncome, stored.Type)
		assert.Equal(t, "EUR", stored.Currency.Code)
		assert.Equal(t, f.incomeCategory.ID, stored.Category.ID)
		assert.Equal(t, otherUser.ID, stored.User.ID)
		assert.Equal(t, domain.TransactionSourceRecurring, stored.Source)
		assert.Equal(t, f.euro.Code, result.Transaction.Currency.Code)
	})

	t.Run("updates the transaction date", func(t *testing.T) {
		f := newUpdateFixture()
		when := time.Date(2024, 3, 15, 12, 30, 0, 0, time.UTC)

		_, err := f.workflow.Update(newAppContext(), 7, transaction.Patch{CreatedAt: ptr(when)})

		require.NoError(t, err)
		require.Len(t, f.transactionRepo.updated, 1)
		assert.True(t, when.Equal(f.transactionRepo.updated[0].CreatedAt))
	})

	t.Run("an empty patch performs no write", func(t *testing.T) {
		f := newUpdateFixture()

		_, err := f.workflow.Update(newAppContext(), 7, transaction.Patch{})

		require.NoError(t, err)
		assert.Empty(t, f.transactionRepo.updated)
	})

	t.Run("unknown transaction is not found", func(t *testing.T) {
		f := newUpdateFixture()
		f.transactionRepo.getResult = nil

		_, err := f.workflow.Update(newAppContext(), 7, transaction.Patch{Name: ptr("X")})

		require.Error(t, err)
		assert.Empty(t, f.transactionRepo.updated)
	})

	t.Run("unknown currency fails before writing", func(t *testing.T) {
		f := newUpdateFixture()

		_, err := f.workflow.Update(newAppContext(), 7, transaction.Patch{CurrencyCode: ptr("GBP")})

		require.Error(t, err)
		assert.Empty(t, f.transactionRepo.updated)
	})

	t.Run("a type no longer matching the category is a validation error before writing", func(t *testing.T) {
		f := newUpdateFixture()

		_, err := f.workflow.Update(newAppContext(), 7, transaction.Patch{Type: ptr(domain.TransactionTypeIncome)})

		require.Error(t, err)
		var domainErr domain.DomainError
		require.ErrorAs(t, err, &domainErr)
		assert.Equal(t, domain.ErrorKindValidation, domainErr.Kind())
		assert.Empty(t, f.transactionRepo.updated)
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
