package workflow

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"jww/internal/domain"
	"jww/internal/service/auth"
	"jww/internal/service/user"
	"jww/internal/service/webauthn_credential"
)

// WebAuthnWorkflow orchestrates the WebAuthn enrollment and login ceremonies
// across user, webauthn_credential, and the outbound Ceremony/ChallengeTokens
// ports. It never creates a user: enrollment (BeginRegistration/
// FinishRegistration) only ever attaches a credential to the caller's own,
// already-authenticated account, and login (BeginLogin/FinishLogin) only
// ever resolves an existing credential to an existing user — this workflow
// has no path that calls user.Service.Create.
type WebAuthnWorkflow struct {
	userService       *user.Service
	credentialService *webauthn_credential.Service
	ceremony          webauthn_credential.Ceremony
	challengeTokens   webauthn_credential.ChallengeTokens
	tokenService      auth.TokenService
}

func NewWebAuthnWorkflow(
	userService *user.Service,
	credentialService *webauthn_credential.Service,
	ceremony webauthn_credential.Ceremony,
	challengeTokens webauthn_credential.ChallengeTokens,
	tokenService auth.TokenService,
) *WebAuthnWorkflow {
	return &WebAuthnWorkflow{
		userService:       userService,
		credentialService: credentialService,
		ceremony:          ceremony,
		challengeTokens:   challengeTokens,
		tokenService:      tokenService,
	}
}

// BeginRegistration starts enrolling a new credential for userID — always
// the caller's own ID, resolved from their access token by the HTTP layer,
// never a request body field.
func (w *WebAuthnWorkflow) BeginRegistration(ctx domain.AppContext, userID uuid.UUID) (options domain.WebAuthnCeremonyOptions, challengeToken string, err error) {
	u, err := w.userService.GetByID(ctx, userID)
	if err != nil {
		return nil, "", err
	}
	existing, err := w.credentialService.List(ctx, userID)
	if err != nil {
		return nil, "", err
	}
	options, session, err := w.ceremony.BeginRegistration(*u, existing)
	if err != nil {
		return nil, "", err
	}
	challengeToken, err = w.challengeTokens.IssueRegistration(userID, json.RawMessage(session))
	if err != nil {
		return nil, "", err
	}
	return options, challengeToken, nil
}

// FinishRegistration verifies credentialResponse against challengeToken and
// persists a new domain.WebAuthnCredential for userID — the caller's own
// ID, same as BeginRegistration. A challengeToken whose embedded user
// doesn't match userID is rejected the same as any other verification
// failure, so one user can never attach a credential using a challenge
// issued for someone else.
func (w *WebAuthnWorkflow) FinishRegistration(ctx domain.AppContext, userID uuid.UUID, challengeToken, name string, credentialResponse domain.WebAuthnCeremonyResponse) (*domain.WebAuthnCredential, error) {
	tokenUserID, session, err := w.challengeTokens.VerifyRegistration(challengeToken)
	if err != nil {
		return nil, domain.NewWebAuthnRegistrationFailedError(err)
	}
	if tokenUserID != userID {
		return nil, domain.NewWebAuthnRegistrationFailedError(nil)
	}

	u, err := w.userService.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	existing, err := w.credentialService.List(ctx, userID)
	if err != nil {
		return nil, err
	}

	credentialID, publicKey, aaguid, attestationType, signCount, transports, backupEligible, err := w.ceremony.FinishRegistration(*u, existing, domain.WebAuthnCeremonySession(session), credentialResponse)
	if err != nil {
		return nil, domain.NewWebAuthnRegistrationFailedError(err)
	}

	created, err := w.credentialService.Create(ctx, userID, name, credentialID, publicKey, attestationType, aaguid, signCount, transports, backupEligible)
	if err != nil {
		return nil, domain.NewWebAuthnRegistrationFailedError(err)
	}
	return created, nil
}

// BeginLogin starts a discoverable (usernameless) login ceremony — no user
// is known yet.
func (w *WebAuthnWorkflow) BeginLogin(ctx domain.AppContext) (options domain.WebAuthnCeremonyOptions, challengeToken string, err error) {
	options, session, err := w.ceremony.BeginLogin()
	if err != nil {
		return nil, "", err
	}
	challengeToken, err = w.challengeTokens.IssueLogin(json.RawMessage(session))
	if err != nil {
		return nil, "", err
	}
	return options, challengeToken, nil
}

// FinishLogin verifies credentialResponse against challengeToken, resolves
// which existing user the assertion belongs to, and — on success — issues a
// normal access token, exactly like a password login. Every failure on this
// path (expired/tampered challenge, unknown credential, bad signature,
// clone-counter mismatch, or any lookup error along the way) gets the same
// generic client-facing message and 401 AuthenticationError already uses —
// same anti-enumeration rule, same reasoning: distinguishing any of these
// for the caller would let them probe which credentials or accounts exist.
// Unlike AuthenticationError, WebAuthnLoginFailedError preserves the real
// cause via Unwrap so boundary logging can still diagnose it.
func (w *WebAuthnWorkflow) FinishLogin(ctx domain.AppContext, challengeToken string, credentialResponse domain.WebAuthnCeremonyResponse) (string, error) {
	session, err := w.challengeTokens.VerifyLogin(challengeToken)
	if err != nil {
		return "", domain.NewWebAuthnLoginFailedError(fmt.Errorf("verifying challenge token: %w", err))
	}

	var storedCredentialID uuid.UUID
	handler := func(rawID, _ []byte) (domain.User, []domain.WebAuthnCredential, error) {
		stored, lookupErr := w.credentialService.GetByCredentialID(ctx, rawID)
		if lookupErr != nil {
			return domain.User{}, nil, fmt.Errorf("looking up credential by id: %w", lookupErr)
		}
		if stored == nil {
			return domain.User{}, nil, fmt.Errorf("no stored credential matches the asserted id")
		}
		storedCredentialID = stored.ID

		u, lookupErr := w.userService.GetByID(ctx, stored.UserID)
		if lookupErr != nil {
			return domain.User{}, nil, fmt.Errorf("looking up credential owner: %w", lookupErr)
		}
		existing, lookupErr := w.credentialService.List(ctx, stored.UserID)
		if lookupErr != nil {
			return domain.User{}, nil, fmt.Errorf("listing credential owner's credentials: %w", lookupErr)
		}
		return *u, existing, nil
	}

	resolvedUser, _, signCount, err := w.ceremony.FinishLogin(handler, domain.WebAuthnCeremonySession(session), credentialResponse)
	if err != nil {
		return "", domain.NewWebAuthnLoginFailedError(fmt.Errorf("ceremony: %w", err))
	}

	if err := w.credentialService.TouchSignCount(ctx, storedCredentialID, signCount, time.Now().UTC()); err != nil {
		return "", err
	}

	return w.tokenService.Issue(resolvedUser.ID)
}

// ListCredentials lists every credential enrolled for userID, for the
// Settings management UI.
func (w *WebAuthnWorkflow) ListCredentials(ctx domain.AppContext, userID uuid.UUID) ([]domain.WebAuthnCredential, error) {
	return w.credentialService.List(ctx, userID)
}

// DeleteCredential removes a credential, but only if it belongs to userID —
// see webauthn_credential.Service.Delete for the anti-enumeration rule.
func (w *WebAuthnWorkflow) DeleteCredential(ctx domain.AppContext, userID, credentialID uuid.UUID) error {
	return w.credentialService.Delete(ctx, userID, credentialID)
}
