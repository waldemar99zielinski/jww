package webauthn_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	webauthnadapter "jww/internal/adapter/security/webauthn"
)

func TestLoadConfig_RequiredAndDefault(t *testing.T) {
	t.Setenv("WEBAUTHN_RP_ID", "example.com")
	t.Setenv("WEBAUTHN_RP_ORIGINS", "https://example.com,https://www.example.com")
	// WEBAUTHN_RP_DISPLAY_NAME is deliberately left unset so LoadConfig falls
	// back to its `default` tag.

	cfg, err := webauthnadapter.LoadConfig()
	require.NoError(t, err)
	assert.Equal(t, "example.com", cfg.RPID)
	assert.Equal(t, "Joint Wallet Watch", cfg.RPDisplayName, "unset env falls back to the default tag")
	assert.Equal(t, []string{"https://example.com", "https://www.example.com"}, cfg.RPOriginsList())
}

func TestLoadConfig_MissingRequiredFails(t *testing.T) {
	// os.LookupEnv treats a present-but-empty value as "set", so a
	// bind:"required" field only fails when the variable is genuinely
	// absent — unset explicitly rather than Setenv("", "").
	os.Unsetenv("WEBAUTHN_RP_ID")
	os.Unsetenv("WEBAUTHN_RP_ORIGINS")

	_, err := webauthnadapter.LoadConfig()
	assert.Error(t, err)
}
