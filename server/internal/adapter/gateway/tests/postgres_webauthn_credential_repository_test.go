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

func newTestWebAuthnCredential(userID uuid.UUID, name string) domain.WebAuthnCredential {
	return domain.WebAuthnCredential{
		ID:              uuid.New(),
		UserID:          userID,
		Name:            name,
		CredentialID:    []byte("cred-" + name),
		PublicKey:       []byte("pubkey-" + name),
		AttestationType: "none",
		AAGUID:          []byte{0, 0, 0, 0},
		SignCount:       0,
		Transports:      []string{"internal", "hybrid"},
		// true (not the zero value) so the round-trip test actually proves the
		// column is read back, not just defaulted — see migration 00011's
		// comment for why this one field previously broke every login.
		BackupEligible: true,
		CreatedAt:      time.Now().UTC().Truncate(time.Microsecond),
	}
}

func TestPostgresWebAuthnCredentialRepository_CreateGetByID_RoundTrip(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	appCtx := pgtest.AppContext(ctx)
	uow := pgtest.BeginRollback(t, pool, appCtx)
	userRepo := gateway.NewPostgresUserRepository(pool)
	user := newTestUser("webauthn-user-1")
	require.NoError(t, userRepo.Create(appCtx, user, uow))
	repo := gateway.NewPostgresWebAuthnCredentialRepository(pool)

	c := newTestWebAuthnCredential(user.ID, "iPhone")
	require.NoError(t, repo.Create(appCtx, c, uow))

	got, err := repo.GetByID(appCtx, c.ID, uow)
	require.NoError(t, err)
	require.NotNil(t, got)
	// time.Time round-trips through Postgres as an equal instant but not
	// necessarily an identical *time.Location value (see
	// postgres_user_repository_test.go's CreateGetByID_RoundTrip for the
	// same pattern) — compare it separately via Equal(), then the rest of
	// the struct by value.
	assert.True(t, c.CreatedAt.Equal(got.CreatedAt), "CreatedAt = %v, want %v", got.CreatedAt, c.CreatedAt)
	gotCopy := *got
	gotCopy.CreatedAt = c.CreatedAt
	assert.Equal(t, c, gotCopy)
}

func TestPostgresWebAuthnCredentialRepository_GetByID_NotFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	appCtx := pgtest.AppContext(ctx)
	uow := pgtest.BeginRollback(t, pool, appCtx)
	repo := gateway.NewPostgresWebAuthnCredentialRepository(pool)

	got, err := repo.GetByID(appCtx, uuid.New(), uow)
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestPostgresWebAuthnCredentialRepository_GetByCredentialID(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	appCtx := pgtest.AppContext(ctx)
	uow := pgtest.BeginRollback(t, pool, appCtx)
	userRepo := gateway.NewPostgresUserRepository(pool)
	user := newTestUser("webauthn-user-2")
	require.NoError(t, userRepo.Create(appCtx, user, uow))
	repo := gateway.NewPostgresWebAuthnCredentialRepository(pool)

	c := newTestWebAuthnCredential(user.ID, "iPad")
	require.NoError(t, repo.Create(appCtx, c, uow))

	got, err := repo.GetByCredentialID(appCtx, c.CredentialID, uow)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, c.ID, got.ID)

	none, err := repo.GetByCredentialID(appCtx, []byte("does-not-exist"), uow)
	require.NoError(t, err)
	assert.Nil(t, none)
}

func TestPostgresWebAuthnCredentialRepository_GetByUserID_ListsOnlyThatUsersCredentials(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	appCtx := pgtest.AppContext(ctx)
	uow := pgtest.BeginRollback(t, pool, appCtx)
	userRepo := gateway.NewPostgresUserRepository(pool)
	userA := newTestUser("webauthn-user-a")
	userB := newTestUser("webauthn-user-b")
	require.NoError(t, userRepo.Create(appCtx, userA, uow))
	require.NoError(t, userRepo.Create(appCtx, userB, uow))
	repo := gateway.NewPostgresWebAuthnCredentialRepository(pool)

	require.NoError(t, repo.Create(appCtx, newTestWebAuthnCredential(userA.ID, "A-iPhone"), uow))
	require.NoError(t, repo.Create(appCtx, newTestWebAuthnCredential(userA.ID, "A-iPad"), uow))
	require.NoError(t, repo.Create(appCtx, newTestWebAuthnCredential(userB.ID, "B-iPhone"), uow))

	got, err := repo.GetByUserID(appCtx, userA.ID, uow)
	require.NoError(t, err)
	assert.Len(t, got, 2)
	for _, c := range got {
		assert.Equal(t, userA.ID, c.UserID)
	}
}

func TestPostgresWebAuthnCredentialRepository_UpdateSignCount(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	appCtx := pgtest.AppContext(ctx)
	uow := pgtest.BeginRollback(t, pool, appCtx)
	userRepo := gateway.NewPostgresUserRepository(pool)
	user := newTestUser("webauthn-user-3")
	require.NoError(t, userRepo.Create(appCtx, user, uow))
	repo := gateway.NewPostgresWebAuthnCredentialRepository(pool)

	c := newTestWebAuthnCredential(user.ID, "iPhone")
	require.NoError(t, repo.Create(appCtx, c, uow))

	lastUsedAt := time.Now().UTC().Truncate(time.Microsecond)
	require.NoError(t, repo.UpdateSignCount(appCtx, c.ID, 7, lastUsedAt, uow))

	got, err := repo.GetByID(appCtx, c.ID, uow)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, uint32(7), got.SignCount)
	require.NotNil(t, got.LastUsedAt)
	assert.Equal(t, lastUsedAt, got.LastUsedAt.UTC())
}

func TestPostgresWebAuthnCredentialRepository_Delete(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	appCtx := pgtest.AppContext(ctx)
	uow := pgtest.BeginRollback(t, pool, appCtx)
	userRepo := gateway.NewPostgresUserRepository(pool)
	user := newTestUser("webauthn-user-4")
	require.NoError(t, userRepo.Create(appCtx, user, uow))
	repo := gateway.NewPostgresWebAuthnCredentialRepository(pool)

	c := newTestWebAuthnCredential(user.ID, "iPhone")
	require.NoError(t, repo.Create(appCtx, c, uow))
	require.NoError(t, repo.Delete(appCtx, c.ID, uow))

	got, err := repo.GetByID(appCtx, c.ID, uow)
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestPostgresWebAuthnCredentialRepository_Create_DuplicateCredentialIDViolatesUniqueIndex(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	appCtx := pgtest.AppContext(ctx)
	uow := pgtest.BeginRollback(t, pool, appCtx)
	userRepo := gateway.NewPostgresUserRepository(pool)
	user := newTestUser("webauthn-user-5")
	require.NoError(t, userRepo.Create(appCtx, user, uow))
	repo := gateway.NewPostgresWebAuthnCredentialRepository(pool)

	c1 := newTestWebAuthnCredential(user.ID, "iPhone")
	c2 := newTestWebAuthnCredential(user.ID, "iPhone-again")
	c2.CredentialID = c1.CredentialID
	require.NoError(t, repo.Create(appCtx, c1, uow))

	err := repo.Create(appCtx, c2, uow)
	assert.Error(t, err, "webauthn_credentials_credential_id_uq should reject a duplicate credential_id")
}
