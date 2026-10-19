package workflow_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"jww/internal/domain"
	"jww/internal/service/transaction"
)

// subTxFixture builds the workflow over fakes with one already-valid expense
// transaction (id 7) loaded, so the sub-transaction CRUD/recompute paths have
// a parent to load and re-persist. The transaction's fields are all valid so
// the amount-accumulation write passes domain.Transaction.Validate.
func subTxFixture() (*transactionWorkflowFixture, domain.TransactionCategory, domain.User) {
	category := groceries()
	user := julcia()
	f := newTransactionWorkflowFixture([]domain.Currency{usd}, []domain.TransactionCategory{category}, []domain.User{user})
	f.transactionRepo.getResult = &domain.Transaction{
		ID:       7,
		Type:     domain.TransactionTypeExpense,
		Amount:   domain.Money(500),
		Currency: usd,
		Category: category,
		User:     user,
		Source:   domain.TransactionSourceManual,
		Name:     "Weekly shop",
	}
	return f, category, user
}

// seedSub puts one line item directly into the fake sub-transaction repository,
// as if it had been persisted earlier.
func seedSub(f *transactionWorkflowFixture, transactionID int64, name string, amount domain.Money) domain.TransactionSubTransaction {
	sub := domain.TransactionSubTransaction{
		ID:            uuid.New(),
		TransactionID: transactionID,
		Name:          name,
		Label:         "",
		Amount:        amount,
	}
	f.subTransactionRepo.byID[sub.ID] = sub
	return sub
}

// TestTransactionWorkflow_Create_WithSubTransactions pins the feature's central
// rule: when items are supplied, the transaction amount is their sum and any
// caller-supplied amount is ignored; with no items the passed amount is kept.
func TestTransactionWorkflow_Create_WithSubTransactions(t *testing.T) {
	t.Run("non-empty items derive the amount as their sum, ignoring the passed amount", func(t *testing.T) {
		category := groceries()
		user := julcia()
		f := newTransactionWorkflowFixture([]domain.Currency{usd}, []domain.TransactionCategory{category}, []domain.User{user})

		created, err := f.workflow.Create(newAppContext(), domain.TransactionTypeExpense, domain.Money(999999), "USD",
			category.ID, user.ID, domain.TransactionSourceManual, "Weekly shop", "", nil, nil,
			[]domain.NewSubTransaction{
				{Name: "Milk", Label: "dairy", Amount: domain.Money(300)},
				{Name: "Bread", Label: "bakery", Amount: domain.Money(450)},
			})

		require.NoError(t, err)
		require.NotNil(t, created)
		assert.Equal(t, domain.Money(750), created.Amount, "the sum of the items wins over the passed amount")
		require.Len(t, created.SubTransactions, 2)
		assert.Equal(t, []string{"Milk", "Bread"}, []string{created.SubTransactions[0].Name, created.SubTransactions[1].Name})
		assert.Equal(t, 2, len(f.subTransactionRepo.byID), "both items are persisted")
		assert.Equal(t, 1, f.transactionRepo.createCalled)
		assert.True(t, f.uowManager.last().finished)
		assert.False(t, f.uowManager.last().restored)
	})

	t.Run("an empty item list keeps the passed amount and a non-nil empty slice", func(t *testing.T) {
		category := groceries()
		user := julcia()
		f := newTransactionWorkflowFixture([]domain.Currency{usd}, []domain.TransactionCategory{category}, []domain.User{user})

		created, err := f.workflow.Create(newAppContext(), domain.TransactionTypeExpense, domain.Money(1500), "USD",
			category.ID, user.ID, domain.TransactionSourceManual, "Weekly shop", "", nil, nil,
			[]domain.NewSubTransaction{})

		require.NoError(t, err)
		require.NotNil(t, created)
		assert.Equal(t, domain.Money(1500), created.Amount, "with no items the passed amount is used unchanged")
		assert.NotNil(t, created.SubTransactions)
		assert.Empty(t, created.SubTransactions)
		assert.Empty(t, f.subTransactionRepo.byID)
		assert.True(t, f.uowManager.last().finished)
	})

	t.Run("a nil item list keeps the passed amount", func(t *testing.T) {
		category := groceries()
		user := julcia()
		f := newTransactionWorkflowFixture([]domain.Currency{usd}, []domain.TransactionCategory{category}, []domain.User{user})

		created, err := f.workflow.Create(newAppContext(), domain.TransactionTypeExpense, domain.Money(1500), "USD",
			category.ID, user.ID, domain.TransactionSourceManual, "Weekly shop", "", nil, nil, nil)

		require.NoError(t, err)
		require.NotNil(t, created)
		assert.Equal(t, domain.Money(1500), created.Amount)
		assert.NotNil(t, created.SubTransactions)
		assert.Empty(t, created.SubTransactions)
	})

	t.Run("an invalid item rolls the whole create back after the transaction insert", func(t *testing.T) {
		category := groceries()
		user := julcia()
		f := newTransactionWorkflowFixture([]domain.Currency{usd}, []domain.TransactionCategory{category}, []domain.User{user})

		// The sum (700) is positive so the transaction insert succeeds; the
		// second item is blank-named and fails validation mid-loop.
		_, err := f.workflow.Create(newAppContext(), domain.TransactionTypeExpense, domain.Money(0), "USD",
			category.ID, user.ID, domain.TransactionSourceManual, "Weekly shop", "", nil, nil,
			[]domain.NewSubTransaction{
				{Name: "Milk", Amount: domain.Money(500)},
				{Name: "  ", Amount: domain.Money(200)},
			})

		require.Error(t, err)
		assert.Equal(t, 1, f.transactionRepo.createCalled, "the transaction insert happened before the bad item")
		assert.True(t, f.uowManager.last().restored)
		assert.False(t, f.uowManager.last().finished)
	})
}

func TestTransactionWorkflow_CreateSubTransaction(t *testing.T) {
	t.Run("adds a row and recomputes the parent amount as the sum", func(t *testing.T) {
		f, _, _ := subTxFixture()
		existing := seedSub(f, 7, "Milk", domain.Money(200))

		updated, err := f.workflow.CreateSubTransaction(newAppContext(), 7, domain.NewSubTransaction{
			Name: "Bread", Label: "bakery", Amount: domain.Money(450),
		})

		require.NoError(t, err)
		require.NotNil(t, updated)
		assert.Equal(t, domain.Money(650), updated.Amount, "200 existing + 450 new")
		require.Len(t, updated.SubTransactions, 2)
		assert.Len(t, f.subTransactionRepo.byID, 2)
		_, stillThere := f.subTransactionRepo.byID[existing.ID]
		assert.True(t, stillThere)
		require.NotEmpty(t, f.transactionRepo.updated)
		assert.Equal(t, domain.Money(650), f.transactionRepo.updated[len(f.transactionRepo.updated)-1].Amount)
		assert.True(t, f.uowManager.last().finished)
		assert.False(t, f.uowManager.last().restored)
	})

	t.Run("a negative discount row lowers the derived total", func(t *testing.T) {
		f, _, _ := subTxFixture()
		seedSub(f, 7, "Milk", domain.Money(500))

		updated, err := f.workflow.CreateSubTransaction(newAppContext(), 7, domain.NewSubTransaction{
			Name: "Coupon", Label: "discount", Amount: domain.Money(-100),
		})

		require.NoError(t, err)
		assert.Equal(t, domain.Money(400), updated.Amount)
		require.Len(t, updated.SubTransactions, 2)
		assert.Equal(t, domain.Money(-100), updated.SubTransactions[1].Amount)
	})

	t.Run("a discount that makes the parent total non-positive rolls back", func(t *testing.T) {
		f, _, _ := subTxFixture()
		seedSub(f, 7, "Milk", domain.Money(100))

		_, err := f.workflow.CreateSubTransaction(newAppContext(), 7, domain.NewSubTransaction{
			Name: "Coupon", Label: "discount", Amount: domain.Money(-100),
		})

		require.Error(t, err)
		assert.True(t, f.uowManager.last().restored)
		assert.False(t, f.uowManager.last().finished)
	})

	t.Run("a missing parent transaction is a not-found and rolls back", func(t *testing.T) {
		f, _, _ := subTxFixture()
		f.transactionRepo.getResult = nil

		_, err := f.workflow.CreateSubTransaction(newAppContext(), 7, domain.NewSubTransaction{Name: "Milk", Amount: domain.Money(200)})

		require.Error(t, err)
		assertKind(t, err, domain.ErrorKindNotFound)
		assert.True(t, f.uowManager.last().restored)
		assert.False(t, f.uowManager.last().finished)
	})

	t.Run("an invalid item (blank name) rolls back before persisting", func(t *testing.T) {
		f, _, _ := subTxFixture()

		_, err := f.workflow.CreateSubTransaction(newAppContext(), 7, domain.NewSubTransaction{Name: "   ", Amount: domain.Money(200)})

		require.Error(t, err)
		assert.True(t, f.uowManager.last().restored)
		assert.False(t, f.uowManager.last().finished)
		assert.Empty(t, f.subTransactionRepo.byID)
	})
}

func TestTransactionWorkflow_UpdateSubTransaction(t *testing.T) {
	t.Run("applies a partial patch and recomputes the sum", func(t *testing.T) {
		f, _, _ := subTxFixture()
		a := seedSub(f, 7, "Milk", domain.Money(300))
		seedSub(f, 7, "Bread", domain.Money(450))

		updated, err := f.workflow.UpdateSubTransaction(newAppContext(), 7, a.ID, domain.SubTransactionPatch{
			Name:   ptr("Oat milk"),
			Label:  ptr("dairy"),
			Amount: ptr(domain.Money(100)),
		})

		require.NoError(t, err)
		require.NotNil(t, updated)
		assert.Equal(t, domain.Money(550), updated.Amount, "100 patched + 450 untouched")
		stored := f.subTransactionRepo.byID[a.ID]
		assert.Equal(t, "Oat milk", stored.Name)
		assert.Equal(t, "dairy", stored.Label)
		assert.Equal(t, domain.Money(100), stored.Amount)
		require.NotEmpty(t, f.transactionRepo.updated)
		assert.Equal(t, domain.Money(550), f.transactionRepo.updated[len(f.transactionRepo.updated)-1].Amount)
		assert.True(t, f.uowManager.last().finished)
	})

	t.Run("a nil patch field leaves that field unchanged", func(t *testing.T) {
		f, _, _ := subTxFixture()
		a := seedSub(f, 7, "Milk", domain.Money(300))

		_, err := f.workflow.UpdateSubTransaction(newAppContext(), 7, a.ID, domain.SubTransactionPatch{
			Label: ptr("dairy"),
		})

		require.NoError(t, err)
		stored := f.subTransactionRepo.byID[a.ID]
		assert.Equal(t, "Milk", stored.Name, "name is untouched")
		assert.Equal(t, "dairy", stored.Label)
		assert.Equal(t, domain.Money(300), stored.Amount)
	})

	t.Run("an item belonging to a different transaction is a not-found", func(t *testing.T) {
		f, _, _ := subTxFixture()
		other := seedSub(f, 999, "Milk", domain.Money(300))

		_, err := f.workflow.UpdateSubTransaction(newAppContext(), 7, other.ID, domain.SubTransactionPatch{Amount: ptr(domain.Money(100))})

		require.Error(t, err)
		var notFound *domain.TransactionSubTransactionNotFoundError
		require.ErrorAs(t, err, &notFound)
		assertKind(t, err, domain.ErrorKindNotFound)
		assert.True(t, f.uowManager.last().restored)
		assert.False(t, f.uowManager.last().finished)
		assert.Empty(t, f.transactionRepo.updated)
	})

	t.Run("a missing item is a not-found", func(t *testing.T) {
		f, _, _ := subTxFixture()

		_, err := f.workflow.UpdateSubTransaction(newAppContext(), 7, uuid.New(), domain.SubTransactionPatch{Amount: ptr(domain.Money(100))})

		require.Error(t, err)
		assertKind(t, err, domain.ErrorKindNotFound)
		assert.True(t, f.uowManager.last().restored)
	})

	t.Run("a missing parent transaction is a not-found", func(t *testing.T) {
		f, _, _ := subTxFixture()
		f.transactionRepo.getResult = nil

		_, err := f.workflow.UpdateSubTransaction(newAppContext(), 7, uuid.New(), domain.SubTransactionPatch{Name: ptr("X")})

		require.Error(t, err)
		assertKind(t, err, domain.ErrorKindNotFound)
		assert.True(t, f.uowManager.last().restored)
	})

	t.Run("an invalid patched amount rolls back", func(t *testing.T) {
		f, _, _ := subTxFixture()
		a := seedSub(f, 7, "Milk", domain.Money(300))

		_, err := f.workflow.UpdateSubTransaction(newAppContext(), 7, a.ID, domain.SubTransactionPatch{Amount: ptr(domain.Money(0))})

		require.Error(t, err)
		assert.True(t, f.uowManager.last().restored)
		assert.False(t, f.uowManager.last().finished)
		assert.Equal(t, domain.Money(300), f.subTransactionRepo.byID[a.ID].Amount, "the invalid edit is not persisted")
	})
}

func TestTransactionWorkflow_DeleteSubTransaction(t *testing.T) {
	t.Run("removes a row and recomputes the sum", func(t *testing.T) {
		f, _, _ := subTxFixture()
		a := seedSub(f, 7, "Milk", domain.Money(300))
		seedSub(f, 7, "Bread", domain.Money(450))

		updated, err := f.workflow.DeleteSubTransaction(newAppContext(), 7, a.ID)

		require.NoError(t, err)
		require.NotNil(t, updated)
		assert.Equal(t, domain.Money(450), updated.Amount)
		assert.NotContains(t, f.subTransactionRepo.byID, a.ID)
		require.NotEmpty(t, f.transactionRepo.updated)
		assert.Equal(t, domain.Money(450), f.transactionRepo.updated[len(f.transactionRepo.updated)-1].Amount)
		assert.True(t, f.uowManager.last().finished)
	})

	t.Run("deleting the last item leaves the amount at its last computed value and releases it", func(t *testing.T) {
		f, _, _ := subTxFixture()
		a := seedSub(f, 7, "Milk", domain.Money(500))

		updated, err := f.workflow.DeleteSubTransaction(newAppContext(), 7, a.ID)

		require.NoError(t, err)
		require.NotNil(t, updated)
		assert.Equal(t, domain.Money(500), updated.Amount, "the last computed amount is kept, not zeroed")
		assert.NotNil(t, updated.SubTransactions)
		assert.Empty(t, updated.SubTransactions)
		assert.True(t, f.uowManager.last().finished)

		// With no items left, an explicit transaction-level amount patch is
		// allowed again.
		result, err := f.workflow.Update(newAppContext(), 7, transaction.Patch{Amount: ptr(domain.Money(999))})
		require.NoError(t, err)
		assert.Equal(t, domain.Money(999), result.Transaction.Amount)
	})

	t.Run("an item belonging to a different transaction is a not-found", func(t *testing.T) {
		f, _, _ := subTxFixture()
		other := seedSub(f, 999, "Milk", domain.Money(300))

		_, err := f.workflow.DeleteSubTransaction(newAppContext(), 7, other.ID)

		require.Error(t, err)
		var notFound *domain.TransactionSubTransactionNotFoundError
		require.ErrorAs(t, err, &notFound)
		assertKind(t, err, domain.ErrorKindNotFound)
		assert.True(t, f.uowManager.last().restored)
		assert.Contains(t, f.subTransactionRepo.byID, other.ID, "nothing is deleted on a mismatch")
	})

	t.Run("a missing item is a not-found", func(t *testing.T) {
		f, _, _ := subTxFixture()

		_, err := f.workflow.DeleteSubTransaction(newAppContext(), 7, uuid.New())

		require.Error(t, err)
		assertKind(t, err, domain.ErrorKindNotFound)
		assert.True(t, f.uowManager.last().restored)
	})

	t.Run("a missing parent transaction is a not-found", func(t *testing.T) {
		f, _, _ := subTxFixture()
		f.transactionRepo.getResult = nil

		_, err := f.workflow.DeleteSubTransaction(newAppContext(), 7, uuid.New())

		require.Error(t, err)
		assertKind(t, err, domain.ErrorKindNotFound)
		assert.True(t, f.uowManager.last().restored)
	})
}

// TestTransactionWorkflow_Update_RejectsAmountWhileItemsExist pins the "amount
// is derived" guard: an explicit amount patch on a transaction with items is a
// validation (400) error; with no items the same patch succeeds.
func TestTransactionWorkflow_Update_RejectsAmountWhileItemsExist(t *testing.T) {
	t.Run("explicit amount patch with items is a validation error and writes nothing", func(t *testing.T) {
		f, _, _ := subTxFixture()
		seedSub(f, 7, "Milk", domain.Money(300))

		_, err := f.workflow.Update(newAppContext(), 7, transaction.Patch{Amount: ptr(domain.Money(999))})

		require.Error(t, err)
		var validation *domain.ValidationError
		require.ErrorAs(t, err, &validation)
		assertKind(t, err, domain.ErrorKindValidation)
		assert.Empty(t, f.transactionRepo.updated)
	})

	t.Run("an explicit amount patch with no items succeeds", func(t *testing.T) {
		f, _, _ := subTxFixture()

		result, err := f.workflow.Update(newAppContext(), 7, transaction.Patch{Amount: ptr(domain.Money(999))})

		require.NoError(t, err)
		assert.Equal(t, domain.Money(999), result.Transaction.Amount)
		require.Len(t, f.transactionRepo.updated, 1)
		assert.Equal(t, domain.Money(999), f.transactionRepo.updated[0].Amount)
	})

	t.Run("a non-amount patch with items is still allowed", func(t *testing.T) {
		f, _, _ := subTxFixture()
		seedSub(f, 7, "Milk", domain.Money(300))

		result, err := f.workflow.Update(newAppContext(), 7, transaction.Patch{Name: ptr("Renamed")})

		require.NoError(t, err)
		assert.Equal(t, "Renamed", result.Transaction.Name)
		assert.Equal(t, domain.Money(500), result.Transaction.Amount, "the amount is untouched by a non-amount patch")
	})
}

// TestTransactionWorkflow_Get_HydratesSubTransactions proves reads always carry
// a non-nil SubTransactions slice, empty when there are none.
func TestTransactionWorkflow_Get_HydratesSubTransactions(t *testing.T) {
	t.Run("empty when the transaction has no items", func(t *testing.T) {
		f, _, _ := subTxFixture()

		got, err := f.workflow.Get(newAppContext(), 7)

		require.NoError(t, err)
		assert.NotNil(t, got.SubTransactions)
		assert.Empty(t, got.SubTransactions)
		assert.True(t, f.uowManager.last().finished)
	})

	t.Run("hydrated with the transaction's items", func(t *testing.T) {
		f, _, _ := subTxFixture()
		seedSub(f, 7, "Milk", domain.Money(300))
		seedSub(f, 7, "Bread", domain.Money(450))

		got, err := f.workflow.Get(newAppContext(), 7)

		require.NoError(t, err)
		require.Len(t, got.SubTransactions, 2)
		amounts := []domain.Money{got.SubTransactions[0].Amount, got.SubTransactions[1].Amount}
		assert.ElementsMatch(t, []domain.Money{domain.Money(300), domain.Money(450)}, amounts)
	})
}

// assertKind asserts err is a domain.DomainError of the expected kind.
func assertKind(t *testing.T, err error, kind domain.ErrorKind) {
	t.Helper()
	var domainErr domain.DomainError
	require.ErrorAs(t, err, &domainErr)
	assert.Equal(t, kind, domainErr.Kind())
}
