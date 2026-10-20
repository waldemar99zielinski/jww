package domain

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// WebAuthnCredential is one registered platform-authenticator credential
// (Face ID, Touch ID, Windows Hello, ...) attached to an existing User. It
// never creates a user on its own — it's purely an additional login method
// for an account that already has a password.
type WebAuthnCredential struct {
	ID              uuid.UUID
	UserID          uuid.UUID
	Name            string
	CredentialID    []byte
	PublicKey       []byte
	AttestationType string
	AAGUID          []byte
	SignCount       uint32
	Transports      []string
	BackupEligible  bool
	CreatedAt       time.Time
	LastUsedAt      *time.Time
}

func (c WebAuthnCredential) Validate() error {
	if c.UserID == uuid.Nil {
		return fmt.Errorf("webauthn credential user_id must not be blank")
	}
	if strings.TrimSpace(c.Name) == "" {
		return fmt.Errorf("webauthn credential name must not be blank")
	}
	if len(c.CredentialID) == 0 {
		return fmt.Errorf("webauthn credential credential_id must not be blank")
	}
	if len(c.PublicKey) == 0 {
		return fmt.Errorf("webauthn credential public_key must not be blank")
	}
	return nil
}
