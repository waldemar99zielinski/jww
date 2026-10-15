//go:build integration

package gateway_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"jww/internal/adapter/gateway"
	"jww/internal/domain"
	"jww/internal/tests/pgtest"
)

func TestPostgresCurrencyRepository_CreateGetByCode_RoundTrip(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	appCtx := pgtest.AppContext(ctx)
	uow := pgtest.BeginRollback(t, pool, appCtx)
	repo := gateway.NewPostgresCurrencyRepository(pool)

	c := domain.Currency{Code: "ZZQ", Name: "Zed Quid", IsBase: false}
	require.NoError(t, repo.Create(appCtx, c, uow))

	got, err := repo.GetByCode(appCtx, "ZZQ", uow)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, c, *got)
}

func TestPostgresCurrencyRepository_GetByCode_NotFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	appCtx := pgtest.AppContext(ctx)
	uow := pgtest.BeginRollback(t, pool, appCtx)
	repo := gateway.NewPostgresCurrencyRepository(pool)

	got, err := repo.GetByCode(appCtx, "NXX", uow)
	require.NoError(t, err)
	assert.Nil(t, got, "a genuine miss must be (nil, nil), never a domain error")
}

func TestPostgresCurrencyRepository_List(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	appCtx := pgtest.AppContext(ctx)
	uow := pgtest.BeginRollback(t, pool, appCtx)
	repo := gateway.NewPostgresCurrencyRepository(pool)

	require.NoError(t, repo.Create(appCtx, domain.Currency{Code: "ZZA", Name: "A", IsBase: false}, uow))
	require.NoError(t, repo.Create(appCtx, domain.Currency{Code: "ZZB", Name: "B", IsBase: false}, uow))

	list, err := repo.List(appCtx, uow)
	require.NoError(t, err)
	// This transaction is rolled back at the end of the test and isolated
	// from every other test's transaction, so only what it created is
	// visible here — asserting the exact set is safe.
	require.Len(t, list, 2)
	assert.Equal(t, "ZZA", list[0].Code, "List orders by code")
	assert.Equal(t, "ZZB", list[1].Code)
}

func TestPostgresCurrencyRepository_Update(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	appCtx := pgtest.AppContext(ctx)
	uow := pgtest.BeginRollback(t, pool, appCtx)
	repo := gateway.NewPostgresCurrencyRepository(pool)

	c := domain.Currency{Code: "ZZC", Name: "Original", IsBase: false}
	require.NoError(t, repo.Create(appCtx, c, uow))

	c.Name = "Renamed"
	require.NoError(t, repo.Update(appCtx, c, uow))

	got, err := repo.GetByCode(appCtx, "ZZC", uow)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "Renamed", got.Name)
}

func TestPostgresCurrencyRepository_Delete(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	appCtx := pgtest.AppContext(ctx)
	uow := pgtest.BeginRollback(t, pool, appCtx)
	repo := gateway.NewPostgresCurrencyRepository(pool)

	c := domain.Currency{Code: "ZZD", Name: "Doomed", IsBase: false}
	require.NoError(t, repo.Create(appCtx, c, uow))
	require.NoError(t, repo.Delete(appCtx, c, uow))

	got, err := repo.GetByCode(appCtx, "ZZD", uow)
	require.NoError(t, err)
	assert.Nil(t, got)
}

// TestPostgresCurrencyRepository_ConstraintViolations proves the gateway
// doesn't swallow or reinterpret a real Postgres constraint error — it's the
// service layer's job to validate first, but if it didn't, the database
// must still refuse. The gateway never invents a domain error here, it just
// surfaces whatever pgx returns.
func TestPostgresCurrencyRepository_ConstraintViolations(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	appCtx := pgtest.AppContext(ctx)

	t.Run("duplicate code violates the primary key", func(t *testing.T) {
		t.Parallel()
		uow := pgtest.BeginRollback(t, pool, appCtx)
		repo := gateway.NewPostgresCurrencyRepository(pool)

		c := domain.Currency{Code: "ZZE", Name: "First", IsBase: false}
		require.NoError(t, repo.Create(appCtx, c, uow))
		err := repo.Create(appCtx, domain.Currency{Code: "ZZE", Name: "Second", IsBase: false}, uow)
		assert.Error(t, err)
	})

	t.Run("lowercase code violates the uppercase CHECK", func(t *testing.T) {
		t.Parallel()
		uow := pgtest.BeginRollback(t, pool, appCtx)
		repo := gateway.NewPostgresCurrencyRepository(pool)

		// domain.Currency.Validate() would normally reject this before it
		// ever reaches the gateway; this proves the DB's own CHECK is a real
		// second line of defense, not just documentation.
		err := repo.Create(appCtx, domain.Currency{Code: "zzf", Name: "lowercase", IsBase: false}, uow)
		assert.Error(t, err)
	})

	t.Run("a second base currency violates the partial unique index", func(t *testing.T) {
		t.Parallel()
		uow := pgtest.BeginRollback(t, pool, appCtx)
		repo := gateway.NewPostgresCurrencyRepository(pool)

		require.NoError(t, repo.Create(appCtx, domain.Currency{Code: "ZZG", Name: "Base one", IsBase: true}, uow))
		err := repo.Create(appCtx, domain.Currency{Code: "ZZH", Name: "Base two", IsBase: true}, uow)
		assert.Error(t, err, "currencies_single_base_uq should reject a second is_base=true row")
	})
}
