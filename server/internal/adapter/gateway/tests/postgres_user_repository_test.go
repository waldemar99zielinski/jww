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
	"jww/internal/tests/pgtest"
)

func newTestUser(username string) domain.User {
	return domain.User{
		ID:           uuid.New(),
		Username:     username,
		PasswordHash: "argon2id$fake-hash-for-tests",
		DisplayName:  "Display " + username,
		CreatedAt:    time.Now().UTC().Truncate(time.Microsecond),
	}
}

func TestPostgresUserRepository_CreateGetByID_RoundTrip(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	appCtx := pgtest.AppContext(ctx)
	uow := pgtest.BeginRollback(t, pool, appCtx)
	repo := gateway.NewPostgresUserRepository(pool)

	u := newTestUser("alice")
	require.NoError(t, repo.Create(appCtx, u, uow))

	got, err := repo.GetByID(appCtx, u.ID, uow)
	require.NoError(t, err)
	require.NotNil(t, got)
	// time.Time round-trips through Postgres as an equal instant but not
	// necessarily an identical *time.Location value (pgx may attach zone
	// data assert.Equal's struct comparison would flag) — compare it
	// separately via Equal(), then the rest of the struct by value.
	assert.True(t, u.CreatedAt.Equal(got.CreatedAt), "CreatedAt = %v, want %v", got.CreatedAt, u.CreatedAt)
	gotCopy := *got
	gotCopy.CreatedAt = u.CreatedAt
	assert.Equal(t, u, gotCopy)
}

func TestPostgresUserRepository_GetByID_NotFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	appCtx := pgtest.AppContext(ctx)
	uow := pgtest.BeginRollback(t, pool, appCtx)
	repo := gateway.NewPostgresUserRepository(pool)

	got, err := repo.GetByID(appCtx, uuid.New(), uow)
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestPostgresUserRepository_GetByUsername(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	appCtx := pgtest.AppContext(ctx)
	uow := pgtest.BeginRollback(t, pool, appCtx)
	repo := gateway.NewPostgresUserRepository(pool)

	u := newTestUser("bob")
	require.NoError(t, repo.Create(appCtx, u, uow))

	got, err := repo.GetByUsername(appCtx, "bob", uow)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, u.ID, got.ID)

	miss, err := repo.GetByUsername(appCtx, "nobody", uow)
	require.NoError(t, err)
	assert.Nil(t, miss)
}

func TestPostgresUserRepository_List(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	appCtx := pgtest.AppContext(ctx)
	uow := pgtest.BeginRollback(t, pool, appCtx)
	repo := gateway.NewPostgresUserRepository(pool)

	require.NoError(t, repo.Create(appCtx, newTestUser("carol"), uow))
	require.NoError(t, repo.Create(appCtx, newTestUser("dave"), uow))

	list, err := repo.List(appCtx, uow)
	require.NoError(t, err)
	require.Len(t, list, 2, "isolated to this test's own rolled-back transaction")
}

func TestPostgresUserRepository_Update(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	appCtx := pgtest.AppContext(ctx)
	uow := pgtest.BeginRollback(t, pool, appCtx)
	repo := gateway.NewPostgresUserRepository(pool)

	u := newTestUser("erin")
	require.NoError(t, repo.Create(appCtx, u, uow))

	u.DisplayName = "Erin Renamed"
	require.NoError(t, repo.Update(appCtx, u, uow))

	got, err := repo.GetByID(appCtx, u.ID, uow)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "Erin Renamed", got.DisplayName)
}

func TestPostgresUserRepository_Delete(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	appCtx := pgtest.AppContext(ctx)
	uow := pgtest.BeginRollback(t, pool, appCtx)
	repo := gateway.NewPostgresUserRepository(pool)

	u := newTestUser("frank")
	require.NoError(t, repo.Create(appCtx, u, uow))
	require.NoError(t, repo.Delete(appCtx, u, uow))

	got, err := repo.GetByID(appCtx, u.ID, uow)
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestPostgresUserRepository_Create_DuplicateUsernameViolatesUniqueConstraint(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	appCtx := pgtest.AppContext(ctx)
	uow := pgtest.BeginRollback(t, pool, appCtx)
	repo := gateway.NewPostgresUserRepository(pool)

	require.NoError(t, repo.Create(appCtx, newTestUser("grace"), uow))
	err := repo.Create(appCtx, newTestUser("grace"), uow)
	assert.Error(t, err, "users_username_uq should reject a second row with the same username")
}
