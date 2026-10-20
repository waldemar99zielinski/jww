package schema_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"jww/internal/adapter/schema"
	"jww/internal/domain"
)

func TestWebAuthnRegisterFinishRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     schema.WebAuthnRegisterFinishRequest
		wantErr bool
	}{
		{"valid", schema.WebAuthnRegisterFinishRequest{ChallengeToken: "tok", Name: "iPhone", Credential: json.RawMessage(`{"id":"..."}`)}, false},
		{"blank challenge_token", schema.WebAuthnRegisterFinishRequest{ChallengeToken: "", Name: "iPhone", Credential: json.RawMessage(`{}`)}, true},
		{"blank name", schema.WebAuthnRegisterFinishRequest{ChallengeToken: "tok", Name: "", Credential: json.RawMessage(`{}`)}, true},
		{"blank credential", schema.WebAuthnRegisterFinishRequest{ChallengeToken: "tok", Name: "iPhone", Credential: nil}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestWebAuthnLoginFinishRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     schema.WebAuthnLoginFinishRequest
		wantErr bool
	}{
		{"valid", schema.WebAuthnLoginFinishRequest{ChallengeToken: "tok", Credential: json.RawMessage(`{"id":"..."}`)}, false},
		{"blank challenge_token", schema.WebAuthnLoginFinishRequest{ChallengeToken: "", Credential: json.RawMessage(`{}`)}, true},
		{"blank credential", schema.WebAuthnLoginFinishRequest{ChallengeToken: "tok", Credential: nil}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestWebAuthnCredentialToResponse(t *testing.T) {
	id := uuid.New()
	userID := uuid.New()
	createdAt := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	lastUsedAt := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	c := domain.WebAuthnCredential{
		ID:              id,
		UserID:          userID,
		Name:            "iPhone",
		CredentialID:    []byte("cred-1"),
		PublicKey:       []byte("should-never-appear-in-response"),
		AttestationType: "none",
		CreatedAt:       createdAt,
		LastUsedAt:      &lastUsedAt,
	}

	resp := schema.WebAuthnCredentialToResponse(c)

	assert.Equal(t, id, resp.ID)
	assert.Equal(t, "iPhone", resp.Name)
	assert.True(t, createdAt.Equal(resp.CreatedAt))
	require.NotNil(t, resp.LastUsedAt)
	assert.True(t, lastUsedAt.Equal(*resp.LastUsedAt))
}

func TestWebAuthnCredentialResponse_JSONNeverContainsKeyMaterial(t *testing.T) {
	c := domain.WebAuthnCredential{
		ID:           uuid.New(),
		UserID:       uuid.New(),
		Name:         "iPhone",
		CredentialID: []byte("super-secret-credential-id"),
		PublicKey:    []byte("super-secret-public-key"),
	}

	data, err := json.Marshal(schema.WebAuthnCredentialToResponse(c))
	require.NoError(t, err)

	var fields map[string]any
	require.NoError(t, json.Unmarshal(data, &fields))
	_, hasCredentialID := fields["credential_id"]
	_, hasPublicKey := fields["public_key"]
	assert.False(t, hasCredentialID, "WebAuthnCredentialResponse JSON must never include credential_id")
	assert.False(t, hasPublicKey, "WebAuthnCredentialResponse JSON must never include public_key")
	assert.NotContains(t, string(data), "super-secret")
}
