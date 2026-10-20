package jwt

import (
	"encoding/json"
	"fmt"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// webauthnRegistrationAudience and webauthnLoginAudience are deliberately
// distinct from each other and from accessTokenAudience: a registration
// challenge can never be replayed as a login challenge (or vice versa), and
// neither can ever be accepted as a user access token.
const (
	webauthnRegistrationAudience = "jww-webauthn-registration"
	webauthnLoginAudience        = "jww-webauthn-login"
)

// webauthnChallengeClaims carries an opaque, caller-marshaled WebAuthn
// ceremony session (go-webauthn's SessionData) between a Begin and a Finish
// call. This package stays agnostic of that struct's shape — it only signs
// and verifies whatever JSON the caller hands it.
type webauthnChallengeClaims struct {
	UserID      string          `json:"user_id,omitempty"`
	SessionData json.RawMessage `json:"session_data"`
	jwtlib.RegisteredClaims
}

// WebAuthnChallengeService issues short-lived, purpose-bound tokens standing
// in for a server-side session store: the WebAuthn ceremony is a two-step
// Begin/Finish exchange that needs to remember a challenge in between, and
// this app has no session cache, so the challenge travels with the client
// instead, signed so it can't be tampered with.
type WebAuthnChallengeService struct {
	codec codec
	ttl   time.Duration
}

func NewWebAuthnChallengeService(cfg Config, ttl time.Duration) *WebAuthnChallengeService {
	return &WebAuthnChallengeService{codec: newCodec(cfg), ttl: ttl}
}

// IssueRegistration binds the challenge to userID, so FinishRegistration can
// reject a challenge token that doesn't belong to the authenticated caller.
func (s *WebAuthnChallengeService) IssueRegistration(userID uuid.UUID, session json.RawMessage) (string, error) {
	return s.issue(webauthnRegistrationAudience, userID.String(), session)
}

// IssueLogin carries no user ID — a discoverable login doesn't know which
// user is authenticating until the authenticator response arrives.
func (s *WebAuthnChallengeService) IssueLogin(session json.RawMessage) (string, error) {
	return s.issue(webauthnLoginAudience, "", session)
}

func (s *WebAuthnChallengeService) issue(audience, userID string, session json.RawMessage) (string, error) {
	now := time.Now().UTC()
	return s.codec.sign(webauthnChallengeClaims{
		UserID:      userID,
		SessionData: session,
		RegisteredClaims: jwtlib.RegisteredClaims{
			Issuer:    tokenIssuer,
			Audience:  jwtlib.ClaimStrings{audience},
			IssuedAt:  jwtlib.NewNumericDate(now),
			ExpiresAt: jwtlib.NewNumericDate(now.Add(s.ttl)),
		},
	})
}

func (s *WebAuthnChallengeService) VerifyRegistration(tokenString string) (uuid.UUID, json.RawMessage, error) {
	claims, err := s.verify(tokenString, webauthnRegistrationAudience)
	if err != nil {
		return uuid.Nil, nil, err
	}
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return uuid.Nil, nil, fmt.Errorf("security/jwt: invalid webauthn registration challenge subject: %w", err)
	}
	return userID, claims.SessionData, nil
}

func (s *WebAuthnChallengeService) VerifyLogin(tokenString string) (json.RawMessage, error) {
	claims, err := s.verify(tokenString, webauthnLoginAudience)
	if err != nil {
		return nil, err
	}
	return claims.SessionData, nil
}

func (s *WebAuthnChallengeService) verify(tokenString, audience string) (*webauthnChallengeClaims, error) {
	claims := &webauthnChallengeClaims{}
	token, err := s.codec.parse(tokenString, claims,
		jwtlib.WithAudience(audience),
		jwtlib.WithIssuer(tokenIssuer),
		jwtlib.WithExpirationRequired(),
		jwtlib.WithIssuedAt(),
	)
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, fmt.Errorf("security/jwt: webauthn challenge is not valid")
	}
	return claims, nil
}
