package webauthn_test

import (
	"encoding/json"
	"testing"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	webauthnadapter "jww/internal/adapter/security/webauthn"
	"jww/internal/domain"
)

func validConfig() webauthnadapter.Config {
	return webauthnadapter.Config{
		RPID:          "example.com",
		RPDisplayName: "Joint Wallet Watch",
		RPOrigins:     "https://example.com",
	}
}

func testUser() domain.User {
	return domain.User{ID: uuid.New(), Username: "wozi", DisplayName: "Wozi", PasswordHash: "hash"}
}

func TestNew_ValidConfigSucceeds(t *testing.T) {
	_, err := webauthnadapter.New(validConfig())
	require.NoError(t, err)
}

func TestAdapter_BeginRegistration_ExcludesExistingCredentialsAndSetsPolicy(t *testing.T) {
	adapter, err := webauthnadapter.New(validConfig())
	require.NoError(t, err)

	user := testUser()
	existing := []domain.WebAuthnCredential{
		{ID: uuid.New(), UserID: user.ID, Name: "iPhone", CredentialID: []byte("cred-1"), PublicKey: []byte("pub-1")},
	}

	options, session, err := adapter.BeginRegistration(user, existing)
	require.NoError(t, err)

	var sessionData struct {
		Challenge string `json:"challenge"`
	}
	require.NoError(t, json.Unmarshal(session, &sessionData))
	require.NotEmpty(t, sessionData.Challenge)

	// options must be the bare PublicKeyCredentialCreationOptionsJSON shape
	// (challenge/rp/user/... at the top level) — exactly what the browser's
	// PublicKeyCredential.parseCreationOptionsFromJSON()/navigator.credentials.create()
	// expects — NOT protocol.CredentialCreation's own wrapper (which nests
	// this one level down under its own "publicKey" key). Unmarshaling into
	// the wrapper type here would silently hide a regression back to that
	// double-nesting bug, since every field would just read as its zero value.
	var parsedOptions protocol.PublicKeyCredentialCreationOptions
	require.NoError(t, json.Unmarshal(options, &parsedOptions))

	require.Len(t, parsedOptions.CredentialExcludeList, 1)
	assert.Equal(t, []byte("cred-1"), []byte(parsedOptions.CredentialExcludeList[0].CredentialID))
	assert.NotEmpty(t, parsedOptions.Challenge, "challenge must be present at the top level, not nested under a wrapper")

	assert.Equal(t, protocol.ResidentKeyRequirementRequired, parsedOptions.AuthenticatorSelection.ResidentKey)
	assert.Equal(t, protocol.VerificationRequired, parsedOptions.AuthenticatorSelection.UserVerification)
	assert.Equal(t, protocol.PreferNoAttestation, parsedOptions.Attestation)
}

func TestAdapter_BeginLogin_ProducesChallenge(t *testing.T) {
	adapter, err := webauthnadapter.New(validConfig())
	require.NoError(t, err)

	options, session, err := adapter.BeginLogin()
	require.NoError(t, err)

	var sessionData struct {
		Challenge string `json:"challenge"`
	}
	require.NoError(t, json.Unmarshal(session, &sessionData))
	assert.NotEmpty(t, sessionData.Challenge)

	// Same bare-options shape as BeginRegistration — see its test's comment.
	var parsedOptions protocol.PublicKeyCredentialRequestOptions
	require.NoError(t, json.Unmarshal(options, &parsedOptions))
	assert.NotEmpty(t, parsedOptions.Challenge, "challenge must be present at the top level, not nested under a wrapper")
}

func TestAdapter_FinishRegistration_MalformedCredentialJSONFails(t *testing.T) {
	adapter, err := webauthnadapter.New(validConfig())
	require.NoError(t, err)

	user := testUser()
	_, session, err := adapter.BeginRegistration(user, nil)
	require.NoError(t, err)

	_, _, _, _, _, _, _, err = adapter.FinishRegistration(user, nil, session, []byte("not json"))
	assert.Error(t, err)
}

func TestAdapter_FinishLogin_MalformedCredentialJSONFails(t *testing.T) {
	adapter, err := webauthnadapter.New(validConfig())
	require.NoError(t, err)

	_, session, err := adapter.BeginLogin()
	require.NoError(t, err)

	handler := func(rawID, userHandle []byte) (domain.User, []domain.WebAuthnCredential, error) {
		t.Fatal("handler must not be reached for a malformed body")
		return domain.User{}, nil, nil
	}
	_, _, _, err = adapter.FinishLogin(handler, session, []byte("not json"))
	assert.Error(t, err)
}
