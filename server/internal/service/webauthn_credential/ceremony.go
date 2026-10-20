package webauthn_credential

import (
	"encoding/json"

	"github.com/google/uuid"

	"jww/internal/domain"
)

// DiscoverableUserHandler resolves which user+credentials a login assertion
// claims to belong to, given the credential's raw ID and the authenticator's
// user handle (our domain.User.ID bytes).
type DiscoverableUserHandler func(rawID, userHandle []byte) (domain.User, []domain.WebAuthnCredential, error)

// Ceremony is the outbound port for the WebAuthn Begin/Finish ceremony
// calls. Every type here is a domain type or a stdlib primitive, so this
// package never needs to import the WebAuthn library itself. Implemented by
// internal/adapter/security/webauthn.Adapter.
type Ceremony interface {
	// BeginRegistration returns options (forwarded verbatim to the
	// browser's navigator.credentials.create()) and session (opaque
	// ceremony state the caller must round-trip to FinishRegistration
	// unchanged).
	BeginRegistration(user domain.User, existing []domain.WebAuthnCredential) (options domain.WebAuthnCeremonyOptions, session domain.WebAuthnCeremonySession, err error)
	// FinishRegistration verifies response (the browser's
	// navigator.credentials.create() result) against session and returns
	// the fields to persist as a new domain.WebAuthnCredential.
	// backupEligible is the authenticator's "BE" flag — per the WebAuthn
	// spec it never changes after registration, and every later login's
	// assertion must report the same value or the ceremony is rejected, so
	// it must be captured here and persisted, not left at its zero value.
	FinishRegistration(user domain.User, existing []domain.WebAuthnCredential, session domain.WebAuthnCeremonySession, response domain.WebAuthnCeremonyResponse) (credentialID, publicKey, aaguid []byte, attestationType string, signCount uint32, transports []string, backupEligible bool, err error)
	// BeginLogin starts a discoverable (usernameless) login ceremony — no
	// user is known yet.
	BeginLogin() (options domain.WebAuthnCeremonyOptions, session domain.WebAuthnCeremonySession, err error)
	// FinishLogin verifies response against session, resolving the owning
	// user via handler, and returns the user plus the fields needed to
	// persist the authenticator's updated signature counter. It also fails
	// if the assertion carries a clone warning (the authenticator's
	// signature counter did not advance past the stored value) — see
	// webauthn.ErrCloneWarning.
	FinishLogin(handler DiscoverableUserHandler, session domain.WebAuthnCeremonySession, response domain.WebAuthnCeremonyResponse) (user domain.User, credentialID []byte, signCount uint32, err error)
}

// ChallengeTokens is the outbound port for signing/verifying the short-lived
// token that carries a Ceremony's session between a Begin and a Finish
// call (this app has no session store — see
// internal/adapter/security/jwt/webauthn_challenge.go for why). Implemented
// by internal/adapter/security/jwt.WebAuthnChallengeService.
type ChallengeTokens interface {
	IssueRegistration(userID uuid.UUID, session json.RawMessage) (string, error)
	VerifyRegistration(token string) (uuid.UUID, json.RawMessage, error)
	IssueLogin(session json.RawMessage) (string, error)
	VerifyLogin(token string) (json.RawMessage, error)
}
