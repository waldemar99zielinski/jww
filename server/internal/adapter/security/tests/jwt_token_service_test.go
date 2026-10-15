package security_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"jww/internal/adapter/security"
)

func newTokenService(t *testing.T, secretKey, algorithm string, expireMinutes int) *security.JwtTokenService {
	t.Helper()
	return security.NewJwtTokenService(security.Config{
		SecretKey:                secretKey,
		Algorithm:                algorithm,
		AccessTokenExpireMinutes: expireMinutes,
	})
}

func TestJwtTokenService_IssueVerify_RoundTrip(t *testing.T) {
	svc := newTokenService(t, "test-secret", "HS256", 60)
	userID := uuid.New()

	token, err := svc.Issue(userID)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	gotUserID, err := svc.Verify(token)
	require.NoError(t, err)
	require.NotNil(t, gotUserID)
	assert.Equal(t, userID, *gotUserID)
}

func TestJwtTokenService_Verify_WrongSecretFails(t *testing.T) {
	issuer := newTokenService(t, "secret-one", "HS256", 60)
	verifier := newTokenService(t, "secret-two", "HS256", 60)

	token, err := issuer.Issue(uuid.New())
	require.NoError(t, err)

	_, err = verifier.Verify(token)
	assert.Error(t, err)
}

func TestJwtTokenService_Verify_ExpiredTokenFails(t *testing.T) {
	// A negative expiry puts ExpiresAt in the past the instant it's issued.
	svc := newTokenService(t, "test-secret", "HS256", -1)

	token, err := svc.Issue(uuid.New())
	require.NoError(t, err)

	_, err = svc.Verify(token)
	assert.Error(t, err)
}

func TestJwtTokenService_Verify_UnexpectedAlgorithmRejected(t *testing.T) {
	// Same secret, different algorithms: the verifier must reject a token
	// signed with an algorithm other than the one it's pinned to, even
	// though the secret would otherwise validate it.
	issuer := newTokenService(t, "shared-secret", "HS384", 60)
	verifier := newTokenService(t, "shared-secret", "HS256", 60)

	token, err := issuer.Issue(uuid.New())
	require.NoError(t, err)

	_, err = verifier.Verify(token)
	assert.Error(t, err)
}

func TestJwtTokenService_Verify_MalformedTokenFails(t *testing.T) {
	svc := newTokenService(t, "test-secret", "HS256", 60)

	_, err := svc.Verify("not.a.jwt")
	assert.Error(t, err)
}
