package signup

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGeneratePassword_MeetsLengthAndCharacterRequirements(t *testing.T) {
	for i := 0; i < 50; i++ {
		password, err := GeneratePassword()
		require.NoError(t, err)

		assert.Len(t, password, GeneratedPasswordLength)
		assert.True(t, strings.ContainsAny(password, passwordLower), "missing a lower-case letter: %q", password)
		assert.True(t, strings.ContainsAny(password, passwordUpper), "missing an upper-case letter: %q", password)
		assert.True(t, strings.ContainsAny(password, passwordDigits), "missing a digit: %q", password)
		assert.True(t, strings.ContainsAny(password, passwordSymbols), "missing a symbol: %q", password)
		assert.False(t, strings.ContainsAny(password, " ="), "contains a character that log handlers quote: %q", password)
	}
}

func TestGeneratePassword_ReturnsDistinctPasswords(t *testing.T) {
	seen := make(map[string]struct{}, 100)
	for i := 0; i < 100; i++ {
		password, err := GeneratePassword()
		require.NoError(t, err)

		_, duplicate := seen[password]
		assert.False(t, duplicate, "generated the same password twice: %q", password)
		seen[password] = struct{}{}
	}
}
