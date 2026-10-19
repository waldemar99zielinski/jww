//go:build integration

package gateway_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"jww/internal/adapter/gateway"
	"jww/internal/domain"
	"jww/internal/service/transaction"
	"jww/internal/service/unit_of_work"
	"jww/internal/tests/pgtest"
)

// transactionFixtures creates the currency, category, and user every
// transaction row's FKs require, all inside uow's transaction.
type transactionFixtures struct {
	currency domain.Currency
	category domain.TransactionCategory
	user     domain.User
}

func seedTransactionFixtures(t *testing.T, appCtx domain.AppContext, uow unit_of_work.UnitOfWork, currencyCode string, categoryType domain.TransactionCategoryType) transactionFixtures {
	t.Helper()
	currencyRepo := gateway.NewPostgresCurrencyRepository(pool)
	categoryRepo := gateway.NewPostgresTransactionCategoryRepository(pool)
	userRepo := gateway.NewPostgresUserRepository(pool)

	currency := domain.Currency{Code: currencyCode, Name: currencyCode + " currency", IsBase: false}
	require.NoError(t, currencyRepo.Create(appCtx, currency, uow))

	category := newTestCategory(categoryType, "Zed Tx Category "+currencyCode, 0)
	require.NoError(t, categoryRepo.Create(appCtx, category, uow))

	user := newTestUser("tx-user-" + currencyCode)
	require.NoError(t, userRepo.Create(appCtx, user, uow))

	return transactionFixtures{currency: currency, category: category, user: user}
}

func TestPostgresTransactionRepository_CreateGetByID_RoundTrip(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	appCtx := pgtest.AppContext(ctx)
	uow := pgtest.BeginRollback(t, pool, appCtx)
	fx := seedTransactionFixtures(t, appCtx, uow, "ZTA", domain.TransactionCategoryTypeExpense)
	repo := gateway.NewPostgresTransactionRepository(pool)

	createdAt := time.Now().UTC().Truncate(time.Microsecond)
	// 1,234,567.89 — several significant digits, chosen so any accidental
	// float64 round trip in the scan path would be caught by exact equality.
	amount := domain.Money(123456789)

	created, err := repo.Create(appCtx, domain.TransactionTypeExpense, amount, fx.currency, fx.category, fx.user,
		domain.TransactionSourceManual, createdAt, "Groceries run", "weekly shop", uow)
	require.NoError(t, err)
	require.NotNil(t, created)
	assert.Greater(t, created.ID, int64(0))

	got, err := repo.GetByID(appCtx, created.ID, uow)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, amount, got.Amount, "amount must round-trip exactly, not drift via float64")
	assert.Equal(t, fx.currency.Code, got.Currency.Code)
	assert.Equal(t, fx.category.ID, got.Category.ID)
	assert.Equal(t, fx.user.ID, got.User.ID)
	assert.Equal(t, "Groceries run", got.Name)
	assert.Equal(t, "weekly shop", got.Description)
	assert.Equal(t, domain.TransactionSourceManual, got.Source)
	assert.WithinDuration(t, createdAt, got.CreatedAt, time.Millisecond)
}

func TestPostgresTransactionRepository_GetByID_NotFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	appCtx := pgtest.AppContext(ctx)
	uow := pgtest.BeginRollback(t, pool, appCtx)
	repo := gateway.NewPostgresTransactionRepository(pool)

	got, err := repo.GetByID(appCtx, 9_999_999_999, uow)
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestPostgresTransactionRepository_Create_MismatchedCategoryTypeViolatesCompositeFK(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	appCtx := pgtest.AppContext(ctx)
	uow := pgtest.BeginRollback(t, pool, appCtx)
	// An expense-typed category, but we Create with TransactionTypeIncome —
	// the service layer's Validate() would normally reject this first; this
	// proves transactions_category_type_fk still refuses it at the DB level.
	fx := seedTransactionFixtures(t, appCtx, uow, "ZTB", domain.TransactionCategoryTypeExpense)
	repo := gateway.NewPostgresTransactionRepository(pool)

	_, err := repo.Create(appCtx, domain.TransactionTypeIncome, domain.Money(100), fx.currency, fx.category, fx.user,
		domain.TransactionSourceManual, time.Now(), "Mismatched", "", uow)
	assert.Error(t, err, "transactions_category_type_fk should reject type/category_id mismatch")
}

func TestPostgresTransactionRepository_List_FiltersByExplicitFields(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	appCtx := pgtest.AppContext(ctx)
	uow := pgtest.BeginRollback(t, pool, appCtx)
	expenseFx := seedTransactionFixtures(t, appCtx, uow, "ZTC", domain.TransactionCategoryTypeExpense)
	incomeFx := seedTransactionFixtures(t, appCtx, uow, "ZTD", domain.TransactionCategoryTypeIncome)
	repo := gateway.NewPostgresTransactionRepository(pool)

	old := time.Now().UTC().Add(-48 * time.Hour).Truncate(time.Microsecond)
	recent := time.Now().UTC().Add(-1 * time.Hour).Truncate(time.Microsecond)

	oldExpense, err := repo.Create(appCtx, domain.TransactionTypeExpense, 1000, expenseFx.currency, expenseFx.category, expenseFx.user,
		domain.TransactionSourceManual, old, "Old expense", "", uow)
	require.NoError(t, err)
	recentExpense, err := repo.Create(appCtx, domain.TransactionTypeExpense, 2000, expenseFx.currency, expenseFx.category, expenseFx.user,
		domain.TransactionSourceManual, recent, "Recent expense", "", uow)
	require.NoError(t, err)
	recentIncome, err := repo.Create(appCtx, domain.TransactionTypeIncome, 3000, incomeFx.currency, incomeFx.category, incomeFx.user,
		domain.TransactionSourceManual, recent, "Recent income", "", uow)
	require.NoError(t, err)

	t.Run("filter by type", func(t *testing.T) {
		expenseType := domain.TransactionTypeExpense
		list, err := repo.List(appCtx, &transaction.Filter{Type: &expenseType}, uow)
		require.NoError(t, err)
		ids := idsOf(list)
		assert.ElementsMatch(t, []int64{oldExpense.ID, recentExpense.ID}, ids)
	})

	t.Run("filter by user", func(t *testing.T) {
		list, err := repo.List(appCtx, &transaction.Filter{UserID: &incomeFx.user.ID}, uow)
		require.NoError(t, err)
		ids := idsOf(list)
		assert.ElementsMatch(t, []int64{recentIncome.ID}, ids)
	})

	t.Run("filter by category", func(t *testing.T) {
		list, err := repo.List(appCtx, &transaction.Filter{CategoryID: &expenseFx.category.ID}, uow)
		require.NoError(t, err)
		ids := idsOf(list)
		assert.ElementsMatch(t, []int64{oldExpense.ID, recentExpense.ID}, ids)
	})

	t.Run("filter by currency code", func(t *testing.T) {
		list, err := repo.List(appCtx, &transaction.Filter{CurrencyCode: incomeFx.currency.Code}, uow)
		require.NoError(t, err)
		ids := idsOf(list)
		assert.ElementsMatch(t, []int64{recentIncome.ID}, ids)
	})

	t.Run("filter by created_from excludes the old one", func(t *testing.T) {
		from := recent.Add(-time.Minute)
		list, err := repo.List(appCtx, &transaction.Filter{CreatedFrom: &from}, uow)
		require.NoError(t, err)
		ids := idsOf(list)
		assert.ElementsMatch(t, []int64{recentExpense.ID, recentIncome.ID}, ids)
		assert.NotContains(t, ids, oldExpense.ID)
	})

	t.Run("no filter returns everything by date descending", func(t *testing.T) {
		list, err := repo.List(appCtx, nil, uow)
		require.NoError(t, err)
		require.Len(t, list, 3)
		assert.Equal(t, []int64{recentIncome.ID, recentExpense.ID, oldExpense.ID}, idsOf(list))
	})

	t.Run("empty filter returns everything by date descending", func(t *testing.T) {
		list, err := repo.List(appCtx, &transaction.Filter{}, uow)
		require.NoError(t, err)
		require.Len(t, list, 3)
		assert.Equal(t, []int64{recentIncome.ID, recentExpense.ID, oldExpense.ID}, idsOf(list))
	})

	t.Run("sort by date ascending", func(t *testing.T) {
		list, err := repo.List(appCtx, &transaction.Filter{
			Sort: transaction.Sort{Key: transaction.SortKeyDate, Order: transaction.SortOrderAsc},
		}, uow)
		require.NoError(t, err)
		assert.Equal(t, []int64{oldExpense.ID, recentExpense.ID, recentIncome.ID}, idsOf(list))
	})

	t.Run("sort key defaults to date when only the order is supplied", func(t *testing.T) {
		list, err := repo.List(appCtx, &transaction.Filter{
			Sort: transaction.Sort{Order: transaction.SortOrderAsc},
		}, uow)
		require.NoError(t, err)
		assert.Equal(t, []int64{oldExpense.ID, recentExpense.ID, recentIncome.ID}, idsOf(list))
	})

	t.Run("sort order defaults to descending when only the key is supplied", func(t *testing.T) {
		list, err := repo.List(appCtx, &transaction.Filter{
			Sort: transaction.Sort{Key: transaction.SortKeyDate},
		}, uow)
		require.NoError(t, err)
		assert.Equal(t, []int64{recentIncome.ID, recentExpense.ID, oldExpense.ID}, idsOf(list))
	})
}

func TestPostgresTransactionRepository_List_DefaultsToDateDescending(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	appCtx := pgtest.AppContext(ctx)
	uow := pgtest.BeginRollback(t, pool, appCtx)
	fx := seedTransactionFixtures(t, appCtx, uow, "ZTG", domain.TransactionCategoryTypeExpense)
	repo := gateway.NewPostgresTransactionRepository(pool)

	now := time.Now().UTC().Truncate(time.Microsecond)
	// Created first → lower id, but the newer created_at.
	lowerIDNewerTime, err := repo.Create(appCtx, domain.TransactionTypeExpense, 1000, fx.currency, fx.category, fx.user,
		domain.TransactionSourceManual, now, "Lower id, newer timestamp", "", uow)
	require.NoError(t, err)
	// Created second → higher id, but an older created_at, so id-descending
	// and created_at-descending disagree.
	higherIDOlderTime, err := repo.Create(appCtx, domain.TransactionTypeExpense, 2000, fx.currency, fx.category, fx.user,
		domain.TransactionSourceManual, now.Add(-time.Hour), "Higher id, older timestamp", "", uow)
	require.NoError(t, err)

	list, err := repo.List(appCtx, nil, uow)
	require.NoError(t, err)
	require.Len(t, list, 2)
	assert.Equal(t, []int64{lowerIDNewerTime.ID, higherIDOlderTime.ID}, idsOf(list),
		"default order must follow created_at, not id")
}

func idsOf(list []domain.Transaction) []int64 {
	ids := make([]int64, len(list))
	for i, tr := range list {
		ids[i] = tr.ID
	}
	return ids
}

func TestPostgresTransactionRepository_Update(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	appCtx := pgtest.AppContext(ctx)
	uow := pgtest.BeginRollback(t, pool, appCtx)
	fx := seedTransactionFixtures(t, appCtx, uow, "ZTE", domain.TransactionCategoryTypeExpense)
	repo := gateway.NewPostgresTransactionRepository(pool)

	created, err := repo.Create(appCtx, domain.TransactionTypeExpense, 1000, fx.currency, fx.category, fx.user,
		domain.TransactionSourceManual, time.Now(), "Original name", "", uow)
	require.NoError(t, err)

	updated := *created
	updated.Amount = 2000
	updated.Name = "Updated name"
	// created_at is the transaction's date and is mutable.
	newCreatedAt := created.CreatedAt.Add(-72 * time.Hour).Truncate(time.Microsecond)
	updated.CreatedAt = newCreatedAt
	require.NoError(t, repo.Update(appCtx, updated, uow))

	got, err := repo.GetByID(appCtx, created.ID, uow)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, domain.Money(2000), got.Amount)
	assert.Equal(t, "Updated name", got.Name)
	assert.WithinDuration(t, newCreatedAt, got.CreatedAt, time.Microsecond)
}

func TestPostgresTransactionRepository_Delete(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	appCtx := pgtest.AppContext(ctx)
	uow := pgtest.BeginRollback(t, pool, appCtx)
	fx := seedTransactionFixtures(t, appCtx, uow, "ZTF", domain.TransactionCategoryTypeExpense)
	repo := gateway.NewPostgresTransactionRepository(pool)

	created, err := repo.Create(appCtx, domain.TransactionTypeExpense, 1000, fx.currency, fx.category, fx.user,
		domain.TransactionSourceManual, time.Now(), "Doomed", "", uow)
	require.NoError(t, err)

	require.NoError(t, repo.Delete(appCtx, *created, uow))

	got, err := repo.GetByID(appCtx, created.ID, uow)
	require.NoError(t, err)
	assert.Nil(t, got)
}
