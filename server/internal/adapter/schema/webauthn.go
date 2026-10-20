package schema

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"jww/internal/domain"
)

// WebAuthnRegisterBeginResponse carries the creation options verbatim — the
// adapter already produces standard WebAuthn JSON, forwarded here as an
// opaque blob for navigator.credentials.create() — plus the challenge token
// the client must echo back to register/finish unchanged.
//
// Options is not a cryptographic key despite the name resemblance to
// domain.WebAuthnCredential.PublicKey — it's the ceremony's options object
// (challenge/rp/user/pubKeyCredParams/...), named to match the browser API
// shape it feeds: navigator.credentials.create({ publicKey: options }).
type WebAuthnRegisterBeginResponse struct {
	Options        json.RawMessage `json:"options"`
	ChallengeToken string          `json:"challenge_token"`
}

// WebAuthnRegisterFinishRequest's Credential is the browser's
// navigator.credentials.create() response, JSON-encoded — opaque here,
// parsed by the WebAuthn ceremony adapter.
type WebAuthnRegisterFinishRequest struct {
	ChallengeToken string          `json:"challenge_token"`
	Name           string          `json:"name"`
	Credential     json.RawMessage `json:"credential"`
}

func (r WebAuthnRegisterFinishRequest) Validate() error {
	if err := firstError(
		minLength("challenge_token", r.ChallengeToken, 1),
		minLength("name", r.Name, 1),
	); err != nil {
		return err
	}
	if len(r.Credential) == 0 {
		return fmt.Errorf("credential must not be blank")
	}
	return nil
}

// WebAuthnLoginBeginResponse mirrors WebAuthnRegisterBeginResponse for the
// discoverable (usernameless) login ceremony.
type WebAuthnLoginBeginResponse struct {
	Options        json.RawMessage `json:"options"`
	ChallengeToken string          `json:"challenge_token"`
}

// WebAuthnLoginFinishRequest's Credential is the browser's
// navigator.credentials.get() response, JSON-encoded.
type WebAuthnLoginFinishRequest struct {
	ChallengeToken string          `json:"challenge_token"`
	Credential     json.RawMessage `json:"credential"`
}

func (r WebAuthnLoginFinishRequest) Validate() error {
	if err := minLength("challenge_token", r.ChallengeToken, 1); err != nil {
		return err
	}
	if len(r.Credential) == 0 {
		return fmt.Errorf("credential must not be blank")
	}
	return nil
}

// WebAuthnCredentialResponse deliberately omits CredentialID/PublicKey —
// this app's own ID is what the delete endpoint takes, same as every other
// resource in this API.
type WebAuthnCredentialResponse struct {
	ID         uuid.UUID  `json:"id"`
	Name       string     `json:"name"`
	CreatedAt  time.Time  `json:"created_at"`
	LastUsedAt *time.Time `json:"last_used_at"`
}

func WebAuthnCredentialToResponse(c domain.WebAuthnCredential) WebAuthnCredentialResponse {
	return WebAuthnCredentialResponse{ID: c.ID, Name: c.Name, CreatedAt: c.CreatedAt, LastUsedAt: c.LastUsedAt}
}
