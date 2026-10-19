//go:build integration

package gateway_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"jww/internal/adapter/gateway"
	"jww/internal/domain"
	"jww/internal/service/unit_of_work"
	"jww/internal/tests/pgtest"
)

// createSubTransactionHost inserts a transaction to own line items and returns
// its id. currencyCode must be unique per test binary run (FK on currencies).
func createSubTransactionHost(t *testing.T, appCtx domain.AppContext, uow unit_of_work.UnitOfWork, currencyCode string) int64 {
	t.Helper()
	fx := seedTransactionFixtures(t, appCtx, uow, currencyCode, domain.TransactionCategoryTypeExpense)
	txRepo := gateway.NewPostgresTransactionRepository(pool)
	tx, err := txRepo.Create(appCtx, domain.TransactionTypeExpense, domain.Money(100), fx.currency, fx.category, fx.user,
		domain.TransactionSourceManual, time.Now(), "sub-transaction host", "", uow)
	require.NoError(t, err)
	return tx.ID
}

func newSubTransaction(txID int64, name string, amount domain.Money) domain.TransactionSubTransaction {
	return domain.TransactionSubTransaction{
		ID:            uuid.New(),
		TransactionID: txID,
		Name:          name,
		Label:         "dairy",
		Amount:        amount,
		CreatedAt:     time.Now().UTC().Truncate(time.Microsecond),
	}
}

// TestPostgresTransactionSubTransactionRepository_RoundTrip pins the one thing a
// fake repo cannot: numeric(12,2) <-> domain.Money round-trips exactly, with no
// float64 drift, even for a value with several significant digits.
func TestPostgresTransactionSubTransactionRepository_RoundTrip(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	appCtx := pgtest.AppContext(ctx)
	uow := pgtest.BeginRollback(t, pool, appCtx)
	repo := gateway.NewPostgresTransactionSubTransactionRepository(pool)

	txID := createSubTransactionHost(t, appCtx, uow, "ZSA")

	// 1,234,567.89 — chosen so an accidental float64 conversion would fail.
	amount := domain.Money(123456789)
	row := newSubTransaction(txID, "Milk 2%", amount)

	require.NoError(t, repo.Create(appCtx, row, uow))

	got, err := repo.GetByID(appCtx, row.ID, uow)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, row.ID, got.ID)
	assert.Equal(t, txID, got.TransactionID)
	assert.Equal(t, "Milk 2%", got.Name)
	assert.Equal(t, "dairy", got.Label)
	assert.Equal(t, amount, got.Amount, "amount must round-trip exactly, never via float64")
	assert.True(t, row.CreatedAt.Equal(got.CreatedAt), "created_at must round-trip")
}

func TestPostgresTransactionSubTransactionRepository_GetByID_NotFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	appCtx := pgtest.AppContext(ctx)
	uow := pgtest.BeginRollback(t, pool, appCtx)
	repo := gateway.NewPostgresTransactionSubTransactionRepository(pool)

	got, err := repo.GetByID(appCtx, uuid.New(), uow)
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestPostgresTransactionSubTransactionRepository_Update(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	appCtx := pgtest.AppContext(ctx)
	uow := pgtest.BeginRollback(t, pool, appCtx)
	repo := gateway.NewPostgresTransactionSubTransactionRepository(pool)

	txID := createSubTransactionHost(t, appCtx, uow, "ZSB")
	row := newSubTransaction(txID, "Coffee", domain.Money(250))
	require.NoError(t, repo.Create(appCtx, row, uow))

	row.Name = "Espresso"
	row.Label = "drinks"
	row.Amount = domain.Money(300)
	require.NoError(t, repo.Update(appCtx, row, uow))

	got, err := repo.GetByID(appCtx, row.ID, uow)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "Espresso", got.Name)
	assert.Equal(t, "drinks", got.Label)
	assert.Equal(t, domain.Money(300), got.Amount)
}

func TestPostgresTransactionSubTransactionRepository_Delete(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	appCtx := pgtest.AppContext(ctx)
	uow := pgtest.BeginRollback(t, pool, appCtx)
	repo := gateway.NewPostgresTransactionSubTransactionRepository(pool)

	txID := createSubTransactionHost(t, appCtx, uow, "ZSC")
	row := newSubTransaction(txID, "ToRemove", domain.Money(100))
	require.NoError(t, repo.Create(appCtx, row, uow))

	require.NoError(t, repo.Delete(appCtx, row, uow))

	got, err := repo.GetByID(appCtx, row.ID, uow)
	require.NoError(t, err)
	assert.Nil(t, got)
}

// TestPostgresTransactionSubTransactionRepository_ListIsParentScoped proves a
// parent-scoped lookup never crosses transactions and is ordered
// deterministically (created_at, then id).
func TestPostgresTransactionSubTransactionRepository_ListIsParentScoped(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	appCtx := pgtest.AppContext(ctx)
	uow := pgtest.BeginRollback(t, pool, appCtx)
	repo := gateway.NewPostgresTransactionSubTransactionRepository(pool)

	hostA := createSubTransactionHost(t, appCtx, uow, "ZSD")
	hostB := createSubTransactionHost(t, appCtx, uow, "ZSE")

	earlier := newSubTransaction(hostA, "A-earlier", domain.Money(100))
	earlier.CreatedAt = time.Now().UTC().Add(-time.Hour).Truncate(time.Microsecond)
	later := newSubTransaction(hostA, "A-later", domain.Money(200))
	other := newSubTransaction(hostB, "B-only", domain.Money(300))
	require.NoError(t, repo.Create(appCtx, earlier, uow))
	require.NoError(t, repo.Create(appCtx, later, uow))
	require.NoError(t, repo.Create(appCtx, other, uow))

	listA, err := repo.ListByTransactionIDs(appCtx, []int64{hostA}, uow)
	require.NoError(t, err)
	require.Len(t, listA, 2, "rows from another transaction must never leak in")
	assert.Equal(t, earlier.ID, listA[0].ID)
	assert.Equal(t, later.ID, listA[1].ID)
	for _, s := range listA {
		assert.Equal(t, hostA, s.TransactionID)
	}

	listB, err := repo.ListByTransactionIDs(appCtx, []int64{hostB}, uow)
	require.NoError(t, err)
	require.Len(t, listB, 1)
	assert.Equal(t, other.ID, listB[0].ID)
}

// TestPostgresTransactionSubTransactionRepository_AmountCheckEnforced proves the
// amount <> 0 CHECK is real at the database, not just in the domain validator.
func TestPostgresTransactionSubTransactionRepository_AmountCheckEnforced(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	appCtx := pgtest.AppContext(ctx)
	uow := pgtest.BeginRollback(t, pool, appCtx)
	repo := gateway.NewPostgresTransactionSubTransactionRepository(pool)

	txID := createSubTransactionHost(t, appCtx, uow, "ZSF")

	zero := newSubTransaction(txID, "Zero", domain.Money(0))
	assert.Error(t, repo.Create(appCtx, zero, uow), "amount = 0 must violate the CHECK")
}

// TestPostgresTransactionSubTransactionRepository_NegativeDiscountAccepted
// proves the replacement CHECK permits negative discount rows.
func TestPostgresTransactionSubTransactionRepository_NegativeDiscountAccepted(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	appCtx := pgtest.AppContext(ctx)
	uow := pgtest.BeginRollback(t, pool, appCtx)
	repo := gateway.NewPostgresTransactionSubTransactionRepository(pool)

	txID := createSubTransactionHost(t, appCtx, uow, "ZSH")

	discount := newSubTransaction(txID, "Coupon", domain.Money(-100))
	assert.NoError(t, repo.Create(appCtx, discount, uow), "a negative discount must be accepted")
}

func TestPostgresTransactionSubTransactionRepository_NameCheckEnforced(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	appCtx := pgtest.AppContext(ctx)
	uow := pgtest.BeginRollback(t, pool, appCtx)
	repo := gateway.NewPostgresTransactionSubTransactionRepository(pool)

	txID := createSubTransactionHost(t, appCtx, uow, "ZSG")

	blank := newSubTransaction(txID, "   ", domain.Money(100))
	assert.Error(t, repo.Create(appCtx, blank, uow), "a blank name must violate the CHECK")
}

// TestPostgresTransactionSubTransactionRepository_CascadeOnTransactionDelete
// proves line items are removed with their parent via the FK, so no orphan rows
// survive a transaction delete.
func TestPostgresTransactionSubTransactionRepository_CascadeOnTransactionDelete(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	appCtx := pgtest.AppContext(ctx)
	uow := pgtest.BeginRollback(t, pool, appCtx)
	repo := gateway.NewPostgresTransactionSubTransactionRepository(pool)
	txRepo := gateway.NewPostgresTransactionRepository(pool)

	fx := seedTransactionFixtures(t, appCtx, uow, "ZSH", domain.TransactionCategoryTypeExpense)
	tx, err := txRepo.Create(appCtx, domain.TransactionTypeExpense, domain.Money(100), fx.currency, fx.category, fx.user,
		domain.TransactionSourceManual, time.Now(), "cascade host", "", uow)
	require.NoError(t, err)

	row := newSubTransaction(tx.ID, "Cascade me", domain.Money(100))
	require.NoError(t, repo.Create(appCtx, row, uow))

	require.NoError(t, txRepo.Delete(appCtx, *tx, uow), "a transaction with only line items must be deletable")

	got, err := repo.GetByID(appCtx, row.ID, uow)
	require.NoError(t, err)
	assert.Nil(t, got, "the line item must be gone with its parent")
}
