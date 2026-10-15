//go:build integration

package pgtest

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// One container is started per test in this file — Start is cheap enough
// (~1-2s) that a dedicated TestMain isn't worth it here, unlike in the
// gateway/e2e suites which share one container across many tests.

func TestStart_AppliesMigrations(t *testing.T) {
	ctx := context.Background()
	pool, err := Start(ctx)
	require.NoError(t, err)
	defer pool.Close(ctx)

	var version int64
	err = pool.QueryRow(ctx, "select version_id from goose_db_version order by id desc limit 1").Scan(&version)
	require.NoError(t, err)
	assert.Greater(t, version, int64(0), "expected at least one migration applied")
}

func TestBeginRollback_IsolatesWrites(t *testing.T) {
	ctx := context.Background()
	pool, err := Start(ctx)
	require.NoError(t, err)
	defer pool.Close(ctx)

	appCtx := AppContext(ctx)
	uow := BeginRollback(t, pool.Pool, appCtx)
	tx, ok := uow.GetValue().(pgx.Tx)
	require.True(t, ok, "GetValue() should return a pgx.Tx")

	_, err = tx.Exec(ctx, "insert into currencies (code, name, is_base) values ('ZZZ', 'Test', false)")
	require.NoError(t, err)

	var countInTx int
	require.NoError(t, tx.QueryRow(ctx, "select count(*) from currencies where code = 'ZZZ'").Scan(&countInTx))
	assert.Equal(t, 1, countInTx, "the insert should be visible inside its own transaction")

	require.NoError(t, uow.Restore(appCtx), "manually trigger the rollback BeginRollback's cleanup will also attempt")

	var countAfterRollback int
	require.NoError(t, pool.QueryRow(ctx, "select count(*) from currencies where code = 'ZZZ'").Scan(&countAfterRollback))
	assert.Equal(t, 0, countAfterRollback, "rollback should have erased the insert")
}
