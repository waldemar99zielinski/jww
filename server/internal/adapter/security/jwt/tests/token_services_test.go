package jwt_test

import (
	"testing"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	jwtadapter "jww/internal/adapter/security/jwt"
	"jww/internal/service/transaction_attachment"
)

func newAccessTokenService(secretKey, algorithm string, expireMinutes int) *jwtadapter.AccessTokenService {
	return jwtadapter.NewAccessTokenService(jwtadapter.Config{
		SecretKey: secretKey, Algorithm: algorithm, AccessTokenExpireMinutes: expireMinutes,
	})
}

func registeredClaims(t *testing.T, token string) jwtlib.RegisteredClaims {
	t.Helper()
	claims := jwtlib.RegisteredClaims{}
	_, _, err := new(jwtlib.Parser).ParseUnverified(token, &claims)
	require.NoError(t, err)
	return claims
}

func TestAccessTokenService_IssueVerify_RoundTrip(t *testing.T) {
	svc := newAccessTokenService("test-secret", "HS256", 60)
	userID := uuid.New()

	token, err := svc.Issue(userID)
	require.NoError(t, err)
	assert.Equal(t, "jww", registeredClaims(t, token).Issuer)
	gotUserID, err := svc.Verify(token)
	require.NoError(t, err)
	assert.Equal(t, userID, *gotUserID)
}

func TestAccessTokenService_VerifyFailures(t *testing.T) {
	t.Run("wrong secret", func(t *testing.T) {
		issuer := newAccessTokenService("secret-one", "HS256", 60)
		verifier := newAccessTokenService("secret-two", "HS256", 60)
		token, err := issuer.Issue(uuid.New())
		require.NoError(t, err)
		_, err = verifier.Verify(token)
		assert.Error(t, err)
	})

	t.Run("expired", func(t *testing.T) {
		svc := newAccessTokenService("test-secret", "HS256", -1)
		token, err := svc.Issue(uuid.New())
		require.NoError(t, err)
		_, err = svc.Verify(token)
		assert.Error(t, err)
	})

	t.Run("unexpected algorithm", func(t *testing.T) {
		issuer := newAccessTokenService("shared-secret", "HS384", 60)
		verifier := newAccessTokenService("shared-secret", "HS256", 60)
		token, err := issuer.Issue(uuid.New())
		require.NoError(t, err)
		_, err = verifier.Verify(token)
		assert.Error(t, err)
	})

	t.Run("malformed", func(t *testing.T) {
		_, err := newAccessTokenService("test-secret", "HS256", 60).Verify("not.a.jwt")
		assert.Error(t, err)
	})
}

func TestAttachmentGrantService_RoundTripAndAudienceSeparation(t *testing.T) {
	cfg := jwtadapter.Config{SecretKey: "shared-secret", Algorithm: "HS256", AccessTokenExpireMinutes: 60}
	accessTokens := jwtadapter.NewAccessTokenService(cfg)
	grants := jwtadapter.NewAttachmentGrantService(cfg, 5*time.Minute)
	attachmentID := uuid.New()
	transactionID := int64(42)

	token, err := grants.Issue(transaction_attachment.GrantClaims{
		AttachmentID: attachmentID, TransactionID: &transactionID,
		StorageVersion: "v1", Purpose: transaction_attachment.GrantPurposePreview,
	})
	require.NoError(t, err)
	assert.Equal(t, "jww", registeredClaims(t, token).Issuer)
	claims, err := grants.Verify(token)
	require.NoError(t, err)
	assert.Equal(t, attachmentID, claims.AttachmentID)
	assert.Equal(t, transactionID, *claims.TransactionID)
	assert.Equal(t, "v1", claims.StorageVersion)
	assert.Equal(t, transaction_attachment.GrantPurposePreview, claims.Purpose)

	_, err = accessTokens.Verify(token)
	assert.Error(t, err)
	accessToken, err := accessTokens.Issue(uuid.New())
	require.NoError(t, err)
	_, err = grants.Verify(accessToken)
	assert.Error(t, err)
}

func TestAttachmentGrantService_ExpiredGrantFails(t *testing.T) {
	cfg := jwtadapter.Config{SecretKey: "test-secret", Algorithm: "HS256", AccessTokenExpireMinutes: 60}
	grants := jwtadapter.NewAttachmentGrantService(cfg, -time.Minute)
	token, err := grants.Issue(transaction_attachment.GrantClaims{
		AttachmentID: uuid.New(), Purpose: transaction_attachment.GrantPurposeDownload,
	})
	require.NoError(t, err)
	_, err = grants.Verify(token)
	assert.Error(t, err)
}
