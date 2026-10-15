package security

import (
	"fmt"

	"github.com/golang-jwt/jwt/v5"

	"jww/utils/configuration"
)

// Config is read from JWT_SECRET_KEY / JWT_ALGORITHM /
// JWT_ACCESS_TOKEN_EXPIRE_MINUTES. No default secret key — a deployment
// that forgets to set this should fail fast, not silently sign tokens with
// a guessable key.
type Config struct {
	SecretKey                string `env:"JWT_SECRET_KEY" bind:"required"`
	Algorithm                string `env:"JWT_ALGORITHM" bind:"optional" default:"HS256"`
	AccessTokenExpireMinutes int    `env:"JWT_ACCESS_TOKEN_EXPIRE_MINUTES" bind:"optional" default:"60"`
}

// LoadConfig also rejects an unsupported JWT_ALGORITHM up front — the same
// fail-fast reasoning as the missing-secret case: NewJwtTokenService would
// otherwise only discover the bad value when it calls
// jwt.GetSigningMethod(...).Sign on the first Issue, i.e. on a user's first
// login, not at startup.
func LoadConfig() (Config, error) {
	cfg, err := configuration.Load[Config]()
	if err != nil {
		return Config{}, err
	}
	if jwt.GetSigningMethod(cfg.Algorithm) == nil {
		return Config{}, fmt.Errorf("security: unsupported JWT_ALGORITHM %q", cfg.Algorithm)
	}
	return cfg, nil
}
