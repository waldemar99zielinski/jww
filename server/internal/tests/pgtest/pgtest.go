// Package pgtest starts a disposable Postgres container for tests that need
// a real database — internal/adapter/gateway's integration tests and
// server/tests/e2e's full-stack tests. It is test-only: nothing outside a
// _test.go file imports it, and it never ships in a production binary.
// See server/TESTING.md for the strategy this implements.
package pgtest

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"jww/internal/domain"
	"jww/internal/infrastructure/database"
	"jww/internal/service/unit_of_work"
)

// image matches development/docker-compose.yml so test and dev behavior stay
// in sync.
const image = "postgres:17-alpine"

// Pool is a running, fully-migrated Postgres instance for tests.
type Pool struct {
	*pgxpool.Pool
	container *postgres.PostgresContainer
	// DSN is the connection string this pool was opened with — needed by
	// anything that can't share the *pgxpool.Pool directly, e.g. e2e's boot
	// smoke test, which points a spawned cmd/server subprocess at the same
	// database via DATABASE_URL.
	DSN string
}

// Start launches a fresh container, opens a pool against it, and applies
// every embedded migration. Call once per test binary (from TestMain) — one
// container is shared by every test in the calling package.
func Start(ctx context.Context) (*Pool, error) {
	container, err := postgres.Run(ctx, image,
		postgres.WithDatabase("jww_test"),
		postgres.WithUsername("jww_test"),
		postgres.WithPassword("jww_test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("pgtest: starting container: %w", err)
	}

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return nil, fmt.Errorf("pgtest: connection string: %w", err)
	}

	pool, err := database.NewPool(ctx, database.Config{URL: dsn})
	if err != nil {
		return nil, fmt.Errorf("pgtest: opening pool: %w", err)
	}

	if err := database.Migrate(ctx, pool); err != nil {
		return nil, fmt.Errorf("pgtest: migrating: %w", err)
	}

	return &Pool{Pool: pool, container: container, DSN: dsn}, nil
}

// Close shuts down the pool and terminates the container. Call from TestMain
// after m.Run().
func (p *Pool) Close(ctx context.Context) {
	p.Pool.Close()
	_ = p.container.Terminate(ctx)
}

// AppContext builds a minimal domain.AppContext for tests that don't go
// through an HTTP request — a context and a fresh flow id, same shape
// scripts/bootstrap uses for its own non-HTTP calls.
func AppContext(ctx context.Context) domain.AppContext {
	var appCtx domain.AppContext
	appCtx.SetContext(ctx)
	appCtx.SetFlowID(uuid.NewString())
	return appCtx
}

// BeginRollback starts a real Postgres transaction via the Unit of Work port
// and registers a cleanup that always rolls it back, pass or fail. Pass the
// returned UnitOfWork to every gateway call in the test so its writes are
// invisible to every other test sharing the same container.
func BeginRollback(tb testing.TB, pool *pgxpool.Pool, appCtx domain.AppContext) unit_of_work.UnitOfWork {
	tb.Helper()
	manager := database.NewPostgresUnitOfWorkManager(pool)
	uow, err := manager.Create(appCtx)
	if err != nil {
		tb.Fatalf("pgtest: beginning transaction: %v", err)
	}
	tb.Cleanup(func() {
		if err := uow.Restore(appCtx); err != nil {
			tb.Logf("pgtest: rollback: %v", err)
		}
	})
	return uow
}
