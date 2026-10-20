package jwt_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	jwtadapter "jww/internal/adapter/security/jwt"
)

func TestWebAuthnChallengeService_Registration_RoundTripAndAudienceSeparation(t *testing.T) {
	cfg := jwtadapter.Config{SecretKey: "shared-secret", Algorithm: "HS256", AccessTokenExpireMinutes: 60}
	accessTokens := jwtadapter.NewAccessTokenService(cfg)
	challenges := jwtadapter.NewWebAuthnChallengeService(cfg, 5*time.Minute)
	userID := uuid.New()
	session := json.RawMessage(`{"challenge":"abc123"}`)

	token, err := challenges.IssueRegistration(userID, session)
	require.NoError(t, err)
	assert.Equal(t, "jww", registeredClaims(t, token).Issuer)

	gotUserID, gotSession, err := challenges.VerifyRegistration(token)
	require.NoError(t, err)
	assert.Equal(t, userID, gotUserID)
	assert.JSONEq(t, string(session), string(gotSession))

	// A registration challenge must never be accepted as a login challenge,
	// nor as a user access token, and vice versa.
	_, err = challenges.VerifyLogin(token)
	assert.Error(t, err)
	_, err = accessTokens.Verify(token)
	assert.Error(t, err)

	accessToken, err := accessTokens.Issue(uuid.New())
	require.NoError(t, err)
	_, _, err = challenges.VerifyRegistration(accessToken)
	assert.Error(t, err)
}

func TestWebAuthnChallengeService_Login_RoundTripCarriesNoUserID(t *testing.T) {
	cfg := jwtadapter.Config{SecretKey: "shared-secret", Algorithm: "HS256", AccessTokenExpireMinutes: 60}
	challenges := jwtadapter.NewWebAuthnChallengeService(cfg, 5*time.Minute)
	session := json.RawMessage(`{"challenge":"xyz789"}`)

	token, err := challenges.IssueLogin(session)
	require.NoError(t, err)

	gotSession, err := challenges.VerifyLogin(token)
	require.NoError(t, err)
	assert.JSONEq(t, string(session), string(gotSession))

	_, _, err = challenges.VerifyRegistration(token)
	assert.Error(t, err)
}

func TestWebAuthnChallengeService_ExpiredChallengeFails(t *testing.T) {
	cfg := jwtadapter.Config{SecretKey: "test-secret", Algorithm: "HS256", AccessTokenExpireMinutes: 60}
	challenges := jwtadapter.NewWebAuthnChallengeService(cfg, -time.Minute)

	token, err := challenges.IssueLogin(json.RawMessage(`{"challenge":"abc"}`))
	require.NoError(t, err)
	_, err = challenges.VerifyLogin(token)
	assert.Error(t, err)
}

func TestWebAuthnChallengeService_WrongSecretFails(t *testing.T) {
	issuer := jwtadapter.NewWebAuthnChallengeService(jwtadapter.Config{SecretKey: "secret-one", Algorithm: "HS256"}, 5*time.Minute)
	verifier := jwtadapter.NewWebAuthnChallengeService(jwtadapter.Config{SecretKey: "secret-two", Algorithm: "HS256"}, 5*time.Minute)

	token, err := issuer.IssueLogin(json.RawMessage(`{"challenge":"abc"}`))
	require.NoError(t, err)
	_, err = verifier.VerifyLogin(token)
	assert.Error(t, err)
}
