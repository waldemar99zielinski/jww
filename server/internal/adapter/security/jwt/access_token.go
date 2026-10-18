package jwt

import (
	"fmt"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const accessTokenAudience = "jww-access"

// AccessTokenService implements the authentication token port.
type AccessTokenService struct {
	codec       codec
	expireAfter time.Duration
}

func NewAccessTokenService(cfg Config) *AccessTokenService {
	return &AccessTokenService{
		codec:       newCodec(cfg),
		expireAfter: time.Duration(cfg.AccessTokenExpireMinutes) * time.Minute,
	}
}

func (s *AccessTokenService) Issue(userID uuid.UUID) (string, error) {
	now := time.Now().UTC()
	return s.codec.sign(jwtlib.RegisteredClaims{
		Issuer:    tokenIssuer,
		Subject:   userID.String(),
		Audience:  jwtlib.ClaimStrings{accessTokenAudience},
		IssuedAt:  jwtlib.NewNumericDate(now),
		ExpiresAt: jwtlib.NewNumericDate(now.Add(s.expireAfter)),
	})
}

func (s *AccessTokenService) Verify(tokenString string) (*uuid.UUID, error) {
	claims := &jwtlib.RegisteredClaims{}
	token, err := s.codec.parse(tokenString, claims,
		jwtlib.WithIssuer(tokenIssuer),
		jwtlib.WithAudience(accessTokenAudience),
		jwtlib.WithExpirationRequired(),
		jwtlib.WithIssuedAt(),
	)
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, fmt.Errorf("security/jwt: access token is not valid")
	}
	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return nil, fmt.Errorf("security/jwt: parsing access token subject: %w", err)
	}
	return &userID, nil
}
