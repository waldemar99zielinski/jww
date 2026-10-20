package domain

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func validWebAuthnCredential() WebAuthnCredential {
	return WebAuthnCredential{
		ID:              uuid.New(),
		UserID:          uuid.New(),
		Name:            "iPhone",
		CredentialID:    []byte{1, 2, 3},
		PublicKey:       []byte{4, 5, 6},
		AttestationType: "none",
	}
}

func TestWebAuthnCredential_Validate(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(c WebAuthnCredential) WebAuthnCredential
		wantErr string
	}{
		{"valid", func(c WebAuthnCredential) WebAuthnCredential { return c }, ""},
		{"blank user_id", func(c WebAuthnCredential) WebAuthnCredential { c.UserID = uuid.Nil; return c }, "user_id"},
		{"blank name", func(c WebAuthnCredential) WebAuthnCredential { c.Name = "  "; return c }, "name"},
		{"blank credential_id", func(c WebAuthnCredential) WebAuthnCredential { c.CredentialID = nil; return c }, "credential_id"},
		{"blank public_key", func(c WebAuthnCredential) WebAuthnCredential { c.PublicKey = nil; return c }, "public_key"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.mutate(validWebAuthnCredential()).Validate()
			if tt.wantErr == "" {
				assert.NoError(t, err)
				return
			}
			assert.ErrorContains(t, err, tt.wantErr)
		})
	}
}
