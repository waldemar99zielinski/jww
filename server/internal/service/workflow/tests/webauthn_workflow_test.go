package workflow_test

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"jww/internal/domain"
	"jww/internal/service/user"
	"jww/internal/service/webauthn_credential"
	"jww/internal/service/workflow"
)

func newWebAuthnWorkflow(
	userRepo *fakeUserRepo,
	credRepo *fakeWebAuthnCredentialRepo,
	ceremony *fakeCeremony,
	challenges *fakeChallengeTokens,
	tokens *fakeTokenService,
) *workflow.WebAuthnWorkflow {
	userService := user.NewService(userRepo, fakePasswordHasher{}, &fakeUoWManager{})
	credService := webauthn_credential.NewService(credRepo, &fakeUoWManager{})
	return workflow.NewWebAuthnWorkflow(userService, credService, ceremony, challenges, tokens)
}

func TestWebAuthnWorkflow_BeginRegistration(t *testing.T) {
	u := domain.User{ID: uuid.New(), Username: "wozi", DisplayName: "Wozi", PasswordHash: "hash"}
	userRepo := newFakeUserRepo(u)
	credRepo := newFakeWebAuthnCredentialRepo()
	ceremony := &fakeCeremony{beginRegOptions: []byte(`{"publicKey":{}}`), beginRegSession: []byte(`{"challenge":"abc"}`)}
	challenges := &fakeChallengeTokens{issueRegToken: "reg-token"}
	w := newWebAuthnWorkflow(userRepo, credRepo, ceremony, challenges, &fakeTokenService{})

	options, token, err := w.BeginRegistration(newAppContext(), u.ID)
	require.NoError(t, err)
	assert.Equal(t, ceremony.beginRegOptions, options)
	assert.Equal(t, "reg-token", token)
}

func TestWebAuthnWorkflow_BeginRegistration_UnknownUserPropagatesError(t *testing.T) {
	userRepo := newFakeUserRepo()
	credRepo := newFakeWebAuthnCredentialRepo()
	w := newWebAuthnWorkflow(userRepo, credRepo, &fakeCeremony{}, &fakeChallengeTokens{}, &fakeTokenService{})

	_, _, err := w.BeginRegistration(newAppContext(), uuid.New())
	require.Error(t, err)
	var notFound domain.DomainError
	require.True(t, errors.As(err, &notFound))
	assert.Equal(t, domain.ErrorKindNotFound, notFound.Kind())
}

func TestWebAuthnWorkflow_FinishRegistration_HappyPathPersistsCredential(t *testing.T) {
	u := domain.User{ID: uuid.New(), Username: "wozi", DisplayName: "Wozi", PasswordHash: "hash"}
	userRepo := newFakeUserRepo(u)
	credRepo := newFakeWebAuthnCredentialRepo()
	ceremony := &fakeCeremony{
		finishRegCredentialID:    []byte("cred-1"),
		finishRegPublicKey:       []byte("pub-1"),
		finishRegAAGUID:          []byte{0},
		finishRegAttestationType: "none",
		finishRegSignCount:       0,
		finishRegTransports:      []string{"internal"},
		finishRegBackupEligible:  true,
	}
	challenges := &fakeChallengeTokens{verifyRegUserID: u.ID, verifyRegSession: []byte(`{"challenge":"abc"}`)}
	w := newWebAuthnWorkflow(userRepo, credRepo, ceremony, challenges, &fakeTokenService{})

	got, err := w.FinishRegistration(newAppContext(), u.ID, "reg-token", "iPhone", []byte(`{"id":"..."}`))
	require.NoError(t, err)
	assert.Equal(t, u.ID, got.UserID)
	assert.Equal(t, "iPhone", got.Name)
	assert.Equal(t, []byte("cred-1"), got.CredentialID)
	// Regression: an earlier version never passed this through at all, so
	// every stored credential silently read back BackupEligible=false,
	// which go-webauthn then rejects on every subsequent login whose
	// assertion reports BE=true (true for any synced/platform passkey,
	// including Face ID) — see migration 00011's comment.
	assert.True(t, got.BackupEligible)
	_, stored := credRepo.byID[got.ID]
	assert.True(t, stored)
}

func TestWebAuthnWorkflow_FinishRegistration_ChallengeForAnotherUserFails(t *testing.T) {
	u := domain.User{ID: uuid.New(), Username: "wozi", DisplayName: "Wozi", PasswordHash: "hash"}
	attacker := domain.User{ID: uuid.New(), Username: "attacker", DisplayName: "Attacker", PasswordHash: "hash"}
	userRepo := newFakeUserRepo(u, attacker)
	credRepo := newFakeWebAuthnCredentialRepo()
	ceremony := &fakeCeremony{}
	// Challenge token was issued for u, but attacker tries to redeem it.
	challenges := &fakeChallengeTokens{verifyRegUserID: u.ID, verifyRegSession: []byte(`{"challenge":"abc"}`)}
	w := newWebAuthnWorkflow(userRepo, credRepo, ceremony, challenges, &fakeTokenService{})

	_, err := w.FinishRegistration(newAppContext(), attacker.ID, "reg-token", "iPhone", []byte(`{}`))
	require.Error(t, err)
	var domainErr domain.DomainError
	require.True(t, errors.As(err, &domainErr))
	assert.Equal(t, domain.ErrorKindValidation, domainErr.Kind())
	assert.Empty(t, credRepo.byID, "no credential must be persisted")
}

func TestWebAuthnWorkflow_FinishRegistration_InvalidChallengeTokenFails(t *testing.T) {
	u := domain.User{ID: uuid.New(), Username: "wozi", DisplayName: "Wozi", PasswordHash: "hash"}
	userRepo := newFakeUserRepo(u)
	credRepo := newFakeWebAuthnCredentialRepo()
	challenges := &fakeChallengeTokens{verifyRegErr: errors.New("expired")}
	w := newWebAuthnWorkflow(userRepo, credRepo, &fakeCeremony{}, challenges, &fakeTokenService{})

	_, err := w.FinishRegistration(newAppContext(), u.ID, "reg-token", "iPhone", []byte(`{}`))
	require.Error(t, err)
	var domainErr domain.DomainError
	require.True(t, errors.As(err, &domainErr))
	assert.Equal(t, domain.ErrorKindValidation, domainErr.Kind())
}

func TestWebAuthnWorkflow_FinishRegistration_CeremonyFailureDoesNotPersist(t *testing.T) {
	u := domain.User{ID: uuid.New(), Username: "wozi", DisplayName: "Wozi", PasswordHash: "hash"}
	userRepo := newFakeUserRepo(u)
	credRepo := newFakeWebAuthnCredentialRepo()
	ceremony := &fakeCeremony{finishRegErr: errors.New("bad signature")}
	challenges := &fakeChallengeTokens{verifyRegUserID: u.ID, verifyRegSession: []byte(`{"challenge":"abc"}`)}
	w := newWebAuthnWorkflow(userRepo, credRepo, ceremony, challenges, &fakeTokenService{})

	_, err := w.FinishRegistration(newAppContext(), u.ID, "reg-token", "iPhone", []byte(`{}`))
	require.Error(t, err)
	var domainErr domain.DomainError
	require.True(t, errors.As(err, &domainErr))
	assert.Equal(t, domain.ErrorKindValidation, domainErr.Kind())
	assert.Empty(t, credRepo.byID)
}

func TestWebAuthnWorkflow_FinishRegistration_DuplicateCredentialFailsGenericNotInternal(t *testing.T) {
	u := domain.User{ID: uuid.New(), Username: "wozi", DisplayName: "Wozi", PasswordHash: "hash"}
	userRepo := newFakeUserRepo(u)
	credRepo := newFakeWebAuthnCredentialRepo()
	credRepo.createErr = errors.New("duplicate key value violates unique constraint \"webauthn_credentials_credential_id_key\"")
	ceremony := &fakeCeremony{
		finishRegCredentialID: []byte("cred-1"),
		finishRegPublicKey:    []byte("pub-1"),
	}
	challenges := &fakeChallengeTokens{verifyRegUserID: u.ID, verifyRegSession: []byte(`{"challenge":"abc"}`)}
	w := newWebAuthnWorkflow(userRepo, credRepo, ceremony, challenges, &fakeTokenService{})

	// A retried/double-submitted finish (e.g. a network retry replaying the
	// same signed credential) must surface as the same client-facing
	// WebAuthnRegistrationFailedError every other failure on this path
	// uses, not an unwrapped repository error that kernel.WriteError can't
	// recognize and falls back to a generic 500 for.
	_, err := w.FinishRegistration(newAppContext(), u.ID, "reg-token", "iPhone", []byte(`{}`))
	require.Error(t, err)
	var domainErr domain.DomainError
	require.True(t, errors.As(err, &domainErr))
	assert.Equal(t, domain.ErrorKindValidation, domainErr.Kind())
}

func TestWebAuthnWorkflow_BeginLogin(t *testing.T) {
	ceremony := &fakeCeremony{beginLoginOptions: []byte(`{"publicKey":{}}`), beginLoginSession: []byte(`{"challenge":"xyz"}`)}
	challenges := &fakeChallengeTokens{issueLoginToken: "login-token"}
	w := newWebAuthnWorkflow(newFakeUserRepo(), newFakeWebAuthnCredentialRepo(), ceremony, challenges, &fakeTokenService{})

	options, token, err := w.BeginLogin(newAppContext())
	require.NoError(t, err)
	assert.Equal(t, ceremony.beginLoginOptions, options)
	assert.Equal(t, "login-token", token)
}

func TestWebAuthnWorkflow_FinishLogin_HappyPathIssuesAccessTokenAndTouchesSignCount(t *testing.T) {
	u := domain.User{ID: uuid.New(), Username: "wozi", DisplayName: "Wozi", PasswordHash: "hash"}
	cred := domain.WebAuthnCredential{ID: uuid.New(), UserID: u.ID, Name: "iPhone", CredentialID: []byte("cred-1"), PublicKey: []byte("pub-1")}
	userRepo := newFakeUserRepo(u)
	credRepo := newFakeWebAuthnCredentialRepo(cred)
	ceremony := &fakeCeremony{
		loginRawID:              cred.CredentialID,
		finishLoginCredentialID: cred.CredentialID,
		finishLoginSignCount:    7,
	}
	challenges := &fakeChallengeTokens{verifyLoginSession: []byte(`{"challenge":"xyz"}`)}
	tokens := &fakeTokenService{issueToken: "access-token"}
	w := newWebAuthnWorkflow(userRepo, credRepo, ceremony, challenges, tokens)

	got, err := w.FinishLogin(newAppContext(), "login-token", []byte(`{}`))
	require.NoError(t, err)
	assert.Equal(t, "access-token", got)
	assert.Equal(t, uint32(7), credRepo.byID[cred.ID].SignCount)
	require.NotNil(t, credRepo.byID[cred.ID].LastUsedAt)
}

func TestWebAuthnWorkflow_FinishLogin_InvalidChallengeTokenFailsGeneric(t *testing.T) {
	challenges := &fakeChallengeTokens{verifyLoginErr: errors.New("expired")}
	w := newWebAuthnWorkflow(newFakeUserRepo(), newFakeWebAuthnCredentialRepo(), &fakeCeremony{}, challenges, &fakeTokenService{})

	_, err := w.FinishLogin(newAppContext(), "login-token", []byte(`{}`))
	assertGenericAuthenticationError(t, err)
}

func TestWebAuthnWorkflow_FinishLogin_UnknownCredentialFailsGeneric(t *testing.T) {
	credRepo := newFakeWebAuthnCredentialRepo() // empty: no credential matches
	ceremony := &fakeCeremony{loginRawID: []byte("does-not-exist")}
	challenges := &fakeChallengeTokens{verifyLoginSession: []byte(`{"challenge":"xyz"}`)}
	w := newWebAuthnWorkflow(newFakeUserRepo(), credRepo, ceremony, challenges, &fakeTokenService{})

	_, err := w.FinishLogin(newAppContext(), "login-token", []byte(`{}`))
	assertGenericAuthenticationError(t, err)
}

func TestWebAuthnWorkflow_FinishLogin_CeremonyVerificationFailureFailsGeneric(t *testing.T) {
	u := domain.User{ID: uuid.New(), Username: "wozi", DisplayName: "Wozi", PasswordHash: "hash"}
	cred := domain.WebAuthnCredential{ID: uuid.New(), UserID: u.ID, Name: "iPhone", CredentialID: []byte("cred-1"), PublicKey: []byte("pub-1")}
	userRepo := newFakeUserRepo(u)
	credRepo := newFakeWebAuthnCredentialRepo(cred)
	ceremony := &fakeCeremony{loginRawID: cred.CredentialID, postHandlerErr: errors.New("bad signature")}
	challenges := &fakeChallengeTokens{verifyLoginSession: []byte(`{"challenge":"xyz"}`)}
	w := newWebAuthnWorkflow(userRepo, credRepo, ceremony, challenges, &fakeTokenService{})

	_, err := w.FinishLogin(newAppContext(), "login-token", []byte(`{}`))
	assertGenericAuthenticationError(t, err)
	assert.Equal(t, uint32(0), credRepo.byID[cred.ID].SignCount, "sign count must not update on a verification failure")
}

func TestWebAuthnWorkflow_FinishLogin_RepoErrorDuringLookupFailsGeneric(t *testing.T) {
	credRepo := newFakeWebAuthnCredentialRepo()
	credRepo.getByCredErr = errors.New("connection refused")
	ceremony := &fakeCeremony{loginRawID: []byte("cred-1")}
	challenges := &fakeChallengeTokens{verifyLoginSession: []byte(`{"challenge":"xyz"}`)}
	w := newWebAuthnWorkflow(newFakeUserRepo(), credRepo, ceremony, challenges, &fakeTokenService{})

	_, err := w.FinishLogin(newAppContext(), "login-token", []byte(`{}`))
	// Even a genuine infra error on this path collapses to the same generic
	// AuthenticationError — see WebAuthnWorkflow.FinishLogin's doc comment
	// for why (anti-enumeration takes priority on the login path).
	assertGenericAuthenticationError(t, err)
}

func assertGenericAuthenticationError(t *testing.T, err error) {
	t.Helper()
	require.Error(t, err)
	var domainErr domain.DomainError
	require.True(t, errors.As(err, &domainErr))
	assert.Equal(t, domain.ErrorKindUnauthorized, domainErr.Kind())
	assert.Equal(t, domain.NewAuthenticationError().Error(), err.Error())
}

func TestWebAuthnWorkflow_ListCredentials(t *testing.T) {
	userID := uuid.New()
	cred := domain.WebAuthnCredential{ID: uuid.New(), UserID: userID, Name: "iPhone", CredentialID: []byte("c"), PublicKey: []byte("p")}
	credRepo := newFakeWebAuthnCredentialRepo(cred)
	w := newWebAuthnWorkflow(newFakeUserRepo(), credRepo, &fakeCeremony{}, &fakeChallengeTokens{}, &fakeTokenService{})

	got, err := w.ListCredentials(newAppContext(), userID)
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, cred.ID, got[0].ID)
}

func TestWebAuthnWorkflow_DeleteCredential_BelongsToAnotherUserFails(t *testing.T) {
	owner := uuid.New()
	attacker := uuid.New()
	cred := domain.WebAuthnCredential{ID: uuid.New(), UserID: owner, Name: "iPhone", CredentialID: []byte("c"), PublicKey: []byte("p")}
	credRepo := newFakeWebAuthnCredentialRepo(cred)
	w := newWebAuthnWorkflow(newFakeUserRepo(), credRepo, &fakeCeremony{}, &fakeChallengeTokens{}, &fakeTokenService{})

	err := w.DeleteCredential(newAppContext(), attacker, cred.ID)
	require.Error(t, err)
	var notFound domain.DomainError
	require.True(t, errors.As(err, &notFound))
	assert.Equal(t, domain.ErrorKindNotFound, notFound.Kind())
	_, stillThere := credRepo.byID[cred.ID]
	assert.True(t, stillThere)
}
