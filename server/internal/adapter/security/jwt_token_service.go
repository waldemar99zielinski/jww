package security

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// JwtTokenService satisfies service/auth.TokenService using golang-jwt/jwt —
// the only place that knows tokens are JWTs rather than some other scheme.
type JwtTokenService struct {
	secretKey            string
	algorithm            string
	signingMethod        jwt.SigningMethod
	accessTokenExpireMin int
}

func NewJwtTokenService(cfg Config) *JwtTokenService {
	return &JwtTokenService{
		secretKey:            cfg.SecretKey,
		algorithm:            cfg.Algorithm,
		signingMethod:        jwt.GetSigningMethod(cfg.Algorithm),
		accessTokenExpireMin: cfg.AccessTokenExpireMinutes,
	}
}

func (s *JwtTokenService) Issue(userID uuid.UUID) (string, error) {
	now := time.Now().UTC()
	claims := jwt.RegisteredClaims{
		Subject:   userID.String(),
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(s.accessTokenExpireMin) * time.Minute)),
	}
	token := jwt.NewWithClaims(s.signingMethod, claims)
	return token.SignedString([]byte(s.secretKey))
}

func (s *JwtTokenService) Verify(tokenString string) (*uuid.UUID, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(t *jwt.Token) (any, error) {
		return []byte(s.secretKey), nil
	}, jwt.WithValidMethods([]string{s.algorithm}))
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, fmt.Errorf("security: token is not valid")
	}
	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok {
		return nil, fmt.Errorf("security: unexpected claims type %T", token.Claims)
	}
	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return nil, fmt.Errorf("security: parsing token subject as uuid: %w", err)
	}
	return &userID, nil
}
