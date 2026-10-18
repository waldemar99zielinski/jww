package jwt_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	jwtadapter "jww/internal/adapter/security/jwt"
)

func unsetEnv(t *testing.T, key string) {
	t.Helper()
	original, had := os.LookupEnv(key)
	require.NoError(t, os.Unsetenv(key))
	t.Cleanup(func() {
		if had {
			_ = os.Setenv(key, original)
		}
	})
}

func TestLoadConfig_ValidEnvSucceedsWithDefaults(t *testing.T) {
	t.Setenv("JWT_SECRET_KEY", "a-secret")
	unsetEnv(t, "JWT_ALGORITHM")
	unsetEnv(t, "JWT_ACCESS_TOKEN_EXPIRE_MINUTES")

	cfg, err := jwtadapter.LoadConfig()
	require.NoError(t, err)
	assert.Equal(t, "a-secret", cfg.SecretKey)
	assert.Equal(t, "HS256", cfg.Algorithm)
	assert.Equal(t, 60, cfg.AccessTokenExpireMinutes)
}

func TestLoadConfig_UnsupportedAlgorithmFailsFast(t *testing.T) {
	t.Setenv("JWT_SECRET_KEY", "a-secret")
	t.Setenv("JWT_ALGORITHM", "not-a-real-algorithm")

	_, err := jwtadapter.LoadConfig()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "JWT_ALGORITHM")
}

func TestLoadConfig_MissingSecretKeyFails(t *testing.T) {
	t.Setenv("JWT_ALGORITHM", "HS256")
	unsetEnv(t, "JWT_SECRET_KEY")

	_, err := jwtadapter.LoadConfig()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "JWT_SECRET_KEY")
}
