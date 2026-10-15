//go:build integration

package gateway_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"jww/internal/adapter/gateway"
	"jww/internal/domain"
	"jww/internal/tests/pgtest"
)

func newTestCategory(categoryType domain.TransactionCategoryType, name string, sortOrder int) domain.TransactionCategory {
	return domain.TransactionCategory{
		ID:        uuid.New(),
		Type:      categoryType,
		Name:      name,
		Active:    true,
		SortOrder: sortOrder,
	}
}

func TestPostgresTransactionCategoryRepository_CreateGetByID_RoundTrip(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	appCtx := pgtest.AppContext(ctx)
	uow := pgtest.BeginRollback(t, pool, appCtx)
	repo := gateway.NewPostgresTransactionCategoryRepository(pool)

	c := newTestCategory(domain.TransactionCategoryTypeExpense, "Zed Groceries", 0)
	require.NoError(t, repo.Create(appCtx, c, uow))

	got, err := repo.GetByID(appCtx, c.ID, uow)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, c, *got)
}

func TestPostgresTransactionCategoryRepository_GetByID_NotFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	appCtx := pgtest.AppContext(ctx)
	uow := pgtest.BeginRollback(t, pool, appCtx)
	repo := gateway.NewPostgresTransactionCategoryRepository(pool)

	got, err := repo.GetByID(appCtx, uuid.New(), uow)
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestPostgresTransactionCategoryRepository_ListByType_FiltersTypeAndActive(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	appCtx := pgtest.AppContext(ctx)
	uow := pgtest.BeginRollback(t, pool, appCtx)
	repo := gateway.NewPostgresTransactionCategoryRepository(pool)

	activeExpense := newTestCategory(domain.TransactionCategoryTypeExpense, "Zed Active Expense", 0)
	inactiveExpense := newTestCategory(domain.TransactionCategoryTypeExpense, "Zed Inactive Expense", 1)
	inactiveExpense.Active = false
	income := newTestCategory(domain.TransactionCategoryTypeIncome, "Zed Income", 0)

	require.NoError(t, repo.Create(appCtx, activeExpense, uow))
	require.NoError(t, repo.Create(appCtx, inactiveExpense, uow))
	require.NoError(t, repo.Create(appCtx, income, uow))

	// migration 00006 seeds 15 pre-existing active "expense" categories, so
	// assert containment/exclusion rather than an exact count.
	activeOnly, err := repo.ListByType(appCtx, domain.TransactionCategoryTypeExpense, true, uow)
	require.NoError(t, err)
	assert.True(t, containsCategoryID(activeOnly, activeExpense.ID), "activeOnly=true must include the active row")
	assert.False(t, containsCategoryID(activeOnly, inactiveExpense.ID), "activeOnly=true must exclude the inactive row")
	assert.False(t, containsCategoryID(activeOnly, income.ID), "expense filter must exclude the income row")

	allExpense, err := repo.ListByType(appCtx, domain.TransactionCategoryTypeExpense, false, uow)
	require.NoError(t, err)
	assert.True(t, containsCategoryID(allExpense, activeExpense.ID), "activeOnly=false must include the active row")
	assert.True(t, containsCategoryID(allExpense, inactiveExpense.ID), "activeOnly=false must include the inactive row too")
	assert.False(t, containsCategoryID(allExpense, income.ID), "expense filter must exclude the income row")
}

func containsCategoryID(categories []domain.TransactionCategory, id uuid.UUID) bool {
	for _, c := range categories {
		if c.ID == id {
			return true
		}
	}
	return false
}

func TestPostgresTransactionCategoryRepository_Update(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	appCtx := pgtest.AppContext(ctx)
	uow := pgtest.BeginRollback(t, pool, appCtx)
	repo := gateway.NewPostgresTransactionCategoryRepository(pool)

	c := newTestCategory(domain.TransactionCategoryTypeExpense, "Zed Original", 0)
	require.NoError(t, repo.Create(appCtx, c, uow))

	c.Name = "Zed Renamed"
	c.Active = false
	require.NoError(t, repo.Update(appCtx, c, uow))

	got, err := repo.GetByID(appCtx, c.ID, uow)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "Zed Renamed", got.Name)
	assert.False(t, got.Active)
}

func TestPostgresTransactionCategoryRepository_Delete(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	appCtx := pgtest.AppContext(ctx)
	uow := pgtest.BeginRollback(t, pool, appCtx)
	repo := gateway.NewPostgresTransactionCategoryRepository(pool)

	c := newTestCategory(domain.TransactionCategoryTypeExpense, "Zed Doomed", 0)
	require.NoError(t, repo.Create(appCtx, c, uow))
	require.NoError(t, repo.Delete(appCtx, c, uow))

	got, err := repo.GetByID(appCtx, c.ID, uow)
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestPostgresTransactionCategoryRepository_Create_DuplicateTypeNameViolatesUniqueIndex(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	appCtx := pgtest.AppContext(ctx)
	uow := pgtest.BeginRollback(t, pool, appCtx)
	repo := gateway.NewPostgresTransactionCategoryRepository(pool)

	require.NoError(t, repo.Create(appCtx, newTestCategory(domain.TransactionCategoryTypeExpense, "Zed Repeat", 0), uow))
	err := repo.Create(appCtx, newTestCategory(domain.TransactionCategoryTypeExpense, "Zed Repeat", 1), uow)
	assert.Error(t, err, "transaction_categories_type_name_uq should reject the same (type, name) twice")
	// Note: a constraint violation aborts the rest of this Postgres
	// transaction (SQLSTATE 25P02 on any further statement) — don't try
	// another Create on this same uow after this point; see the next test.
}

func TestPostgresTransactionCategoryRepository_Create_SameNameDifferentTypeAllowed(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	appCtx := pgtest.AppContext(ctx)
	uow := pgtest.BeginRollback(t, pool, appCtx)
	repo := gateway.NewPostgresTransactionCategoryRepository(pool)

	// The same name under a DIFFERENT type is explicitly allowed (migration
	// comment: "Gifts" can be both an income and an expense category).
	require.NoError(t, repo.Create(appCtx, newTestCategory(domain.TransactionCategoryTypeExpense, "Zed Repeat", 0), uow))
	err := repo.Create(appCtx, newTestCategory(domain.TransactionCategoryTypeIncome, "Zed Repeat", 0), uow)
	assert.NoError(t, err, "the same name under a different type must be allowed")
}
