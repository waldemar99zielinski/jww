package webauthn

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-webauthn/webauthn/protocol"
	libwebauthn "github.com/go-webauthn/webauthn/webauthn"

	"jww/internal/domain"
	"jww/internal/service/webauthn_credential"
)

// ErrCloneWarning is returned by FinishLogin when the authenticator's sign
// counter didn't advance past its stored value — the spec's signal that
// its private key may have been cloned onto a second authenticator.
//
// The signature itself is still valid, so the library only flags this
// (libwebauthn.Authenticator.CloneWarning), it doesn't error. This adapter
// turns that flag into an error, since FinishLogin's callers must reject it.
var ErrCloneWarning = errors.New("authenticator clone warning: sign counter did not advance")

// Adapter wraps the library's ceremony calls behind plain JSON bytes and
// domain types, so nothing outside this package ever imports go-webauthn
// directly — in particular, it implements
// service/webauthn_credential.Ceremony without that package needing to know
// this library exists.
type Adapter struct {
	lib *libwebauthn.WebAuthn
}

func New(cfg Config) (*Adapter, error) {
	lib, err := libwebauthn.New(&libwebauthn.Config{
		RPID:          cfg.RPID,
		RPDisplayName: cfg.RPDisplayName,
		RPOrigins:     cfg.RPOriginsList(),
		// Resident keys + required user verification is what makes this a
		// true "usernameless" biometric flow: the credential is discoverable
		// on the device itself and every assertion re-proves Face ID/Touch
		// ID/etc, not just possession of the device.
		AuthenticatorSelection: protocol.AuthenticatorSelection{
			ResidentKey:      protocol.ResidentKeyRequirementRequired,
			UserVerification: protocol.VerificationRequired,
		},
		// "none" keeps this to a standard consumer passkey flow with no FIDO
		// Metadata Service dependency — this app trusts the platform's own
		// user-verification gate, not a specific authenticator's attestation
		// chain.
		AttestationPreference: protocol.PreferNoAttestation,
	})
	if err != nil {
		return nil, err
	}
	return &Adapter{lib: lib}, nil
}

// BeginRegistration starts enrolling a new credential for user, explicitly
// excluding their existing ones (the library doesn't infer the exclusion
// list from WebAuthnCredentials() on its own).
//
// options: the publicKey creation options, forwarded verbatim to the
// browser's navigator.credentials.create(). This is marshaled from
// creation.Response, not creation itself — the library's
// *protocol.CredentialCreation wraps the real options one level down (under
// its own "publicKey" key), so marshaling creation directly would send the
// browser a doubly-nested object missing challenge/rp/user at the top
// level.
//
// session: opaque ceremony state. Round-trip it to FinishRegistration
// unchanged.
func (a *Adapter) BeginRegistration(user domain.User, existing []domain.WebAuthnCredential) (options domain.WebAuthnCeremonyOptions, session domain.WebAuthnCeremonySession, err error) {
	creation, sessionData, err := a.lib.BeginRegistration(
		webAuthnUser{user: user, credentials: existing},
		libwebauthn.WithExclusions(exclusionList(existing)),
	)
	if err != nil {
		return nil, nil, err
	}
	optionsJSON, err := json.Marshal(creation.Response)
	if err != nil {
		return nil, nil, err
	}
	sessionJSON, err := json.Marshal(sessionData)
	if err != nil {
		return nil, nil, err
	}
	return domain.WebAuthnCeremonyOptions(optionsJSON), domain.WebAuthnCeremonySession(sessionJSON), nil
}

func exclusionList(existing []domain.WebAuthnCredential) []protocol.CredentialDescriptor {
	out := make([]protocol.CredentialDescriptor, len(existing))
	for i, c := range existing {
		out[i] = protocol.CredentialDescriptor{
			Type:         protocol.PublicKeyCredentialType,
			CredentialID: c.CredentialID,
		}
	}
	return out
}

// FinishRegistration verifies response (the browser's
// navigator.credentials.create() result) against session (from
// BeginRegistration).
//
// Returns the fields this app persists as a new domain.WebAuthnCredential:
// credentialID, publicKey, aaguid, attestationType, signCount, transports,
// backupEligible.
func (a *Adapter) FinishRegistration(user domain.User, existing []domain.WebAuthnCredential, session domain.WebAuthnCeremonySession, response domain.WebAuthnCeremonyResponse) (credentialID, publicKey, aaguid []byte, attestationType string, signCount uint32, transports []string, backupEligible bool, err error) {
	var sessionData libwebauthn.SessionData
	if err = json.Unmarshal(session, &sessionData); err != nil {
		return nil, nil, nil, "", 0, nil, false, err
	}
	request, err := newRequest(response)
	if err != nil {
		return nil, nil, nil, "", 0, nil, false, err
	}
	cred, err := a.lib.FinishRegistration(webAuthnUser{user: user, credentials: existing}, sessionData, request)
	if err != nil {
		return nil, nil, nil, "", 0, nil, false, err
	}
	publicKey, aaguid, attestationType, signCount, transports, backupEligible = fromLibraryCredential(cred)
	return cred.ID, publicKey, aaguid, attestationType, signCount, transports, backupEligible, nil
}

// BeginLogin starts a discoverable (usernameless) login ceremony — no user
// is known yet.
//
// options: the publicKey request options, forwarded to the browser's
// navigator.credentials.get(). Same assertion.Response-vs-wrapper reasoning
// as BeginRegistration's options.
//
// session: opaque ceremony state. Round-trip it to FinishLogin unchanged.
func (a *Adapter) BeginLogin() (options domain.WebAuthnCeremonyOptions, session domain.WebAuthnCeremonySession, err error) {
	assertion, sessionData, err := a.lib.BeginDiscoverableLogin()
	if err != nil {
		return nil, nil, err
	}
	optionsJSON, err := json.Marshal(assertion.Response)
	if err != nil {
		return nil, nil, err
	}
	sessionJSON, err := json.Marshal(sessionData)
	if err != nil {
		return nil, nil, err
	}
	return domain.WebAuthnCeremonyOptions(optionsJSON), domain.WebAuthnCeremonySession(sessionJSON), nil
}

// FinishLogin verifies response (the browser's navigator.credentials.get()
// result) against session, resolving the owning user through handler.
//
// Returns the resolved user, the credential's ID, and its authenticator's
// new signature counter — the caller persists the counter so the next
// login can detect a cloned authenticator. Fails with ErrCloneWarning
// instead if the counter didn't advance.
func (a *Adapter) FinishLogin(handler webauthn_credential.DiscoverableUserHandler, session domain.WebAuthnCeremonySession, response domain.WebAuthnCeremonyResponse) (user domain.User, credentialID []byte, signCount uint32, err error) {
	var sessionData libwebauthn.SessionData
	if err = json.Unmarshal(session, &sessionData); err != nil {
		return domain.User{}, nil, 0, err
	}
	request, err := newRequest(response)
	if err != nil {
		return domain.User{}, nil, 0, err
	}
	var resolvedUser domain.User
	libHandler := func(rawID, userHandle []byte) (libwebauthn.User, error) {
		u, credentials, handlerErr := handler(rawID, userHandle)
		if handlerErr != nil {
			return nil, handlerErr
		}
		resolvedUser = u
		return webAuthnUser{user: u, credentials: credentials}, nil
	}
	cred, err := a.lib.FinishDiscoverableLogin(libHandler, sessionData, request)
	if err != nil {
		return domain.User{}, nil, 0, err
	}
	if cred.Authenticator.CloneWarning {
		return domain.User{}, nil, 0, ErrCloneWarning
	}
	return resolvedUser, cred.ID, cred.Authenticator.SignCount, nil
}

// newRequest wraps a raw JSON body in an *http.Request purely to satisfy the
// library's Finish* signatures — no network is involved.
func newRequest(rawJSON []byte) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodPost, "/", bytes.NewReader(rawJSON))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}
