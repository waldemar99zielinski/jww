package security_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"jww/internal/adapter/security"
)

// unsetEnv genuinely removes an environment variable for the test's
// duration and restores it afterward. t.Setenv can't express "unset" (an
// empty string is still "present" to os.LookupEnv, which is what
// utils/configuration.Load checks), so this is needed anywhere a test wants
// to exercise the "var is absent" path.
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

	cfg, err := security.LoadConfig()
	require.NoError(t, err)
	assert.Equal(t, "a-secret", cfg.SecretKey)
	assert.Equal(t, "HS256", cfg.Algorithm, "default algorithm")
	assert.Equal(t, 60, cfg.AccessTokenExpireMinutes, "default expiry")
}

func TestLoadConfig_UnsupportedAlgorithmFailsFast(t *testing.T) {
	t.Setenv("JWT_SECRET_KEY", "a-secret")
	t.Setenv("JWT_ALGORITHM", "not-a-real-algorithm")

	_, err := security.LoadConfig()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "JWT_ALGORITHM")
}

func TestLoadConfig_MissingSecretKeyFails(t *testing.T) {
	// JWT_SECRET_KEY has no default — a deployment that forgets to set it
	// must fail at startup, not silently sign tokens with an empty key.
	t.Setenv("JWT_ALGORITHM", "HS256")
	unsetEnv(t, "JWT_SECRET_KEY")

	_, err := security.LoadConfig()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "JWT_SECRET_KEY")
}
