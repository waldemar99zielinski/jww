// Package jwt implements the application's signed token adapters. Shared JWT
// configuration and cryptographic mechanics live here; each token purpose has
// its own service and claims in a separate file.
package jwt

import (
	"fmt"

	jwtlib "github.com/golang-jwt/jwt/v5"

	"jww/utils/configuration"
)

// Config is shared by every JWT purpose. Audiences and claim shapes remain
// purpose-specific and are not configurable.
type Config struct {
	SecretKey                string `env:"JWT_SECRET_KEY" bind:"required"`
	Algorithm                string `env:"JWT_ALGORITHM" bind:"optional" default:"HS256"`
	AccessTokenExpireMinutes int    `env:"JWT_ACCESS_TOKEN_EXPIRE_MINUTES" bind:"optional" default:"60"`
}

func LoadConfig() (Config, error) {
	cfg, err := configuration.Load[Config]()
	if err != nil {
		return Config{}, err
	}
	if jwtlib.GetSigningMethod(cfg.Algorithm) == nil {
		return Config{}, fmt.Errorf("security/jwt: unsupported JWT_ALGORITHM %q", cfg.Algorithm)
	}
	return cfg, nil
}
