//go:build integration

// Package gateway_test exercises every Postgres repository against a real,
// disposable database (see internal/tests/pgtest) — never a mocked DBTX.
// One container is started once for the whole package via TestMain; each
// test gets its own rolled-back transaction via pgtest.BeginRollback, so
// writes from one test are never visible to another, and tests may run
// with t.Parallel().
package gateway_test

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"jww/internal/tests/pgtest"
)

var pool *pgxpool.Pool

func TestMain(m *testing.M) {
	ctx := context.Background()
	p, err := pgtest.Start(ctx)
	if err != nil {
		panic(err)
	}
	pool = p.Pool
	code := m.Run()
	p.Close(ctx)
	os.Exit(code)
}
