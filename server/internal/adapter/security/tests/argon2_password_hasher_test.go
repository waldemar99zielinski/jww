package security_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"jww/internal/adapter/security"
)

func TestArgon2PasswordHasher_Hash_ProducesArgon2idPHCFormat(t *testing.T) {
	// The encoded hash must be a standard PHC string: $argon2id$v=<version>
	// $m=<memory>,t=<iterations>,p=<threads>$<salt>$<key>, base64 (no
	// padding) for salt/key — this is the exact shape Verify's
	// strings.Split(hash, "$") parses back apart.
	h := security.NewArgon2PasswordHasher()

	hash, err := h.Hash("correct horse battery staple")
	require.NoError(t, err)

	assert.True(t, strings.HasPrefix(hash, "$argon2id$v=19$m=65536,t=3,p=4$"), "hash %q missing expected prefix", hash)

	parts := strings.Split(hash, "$")
	require.Len(t, parts, 6, "PHC string must have 6 '$'-separated parts (leading empty + 5)")
	assert.Equal(t, "", parts[0])
	assert.Equal(t, "argon2id", parts[1])
	assert.Equal(t, "v=19", parts[2])
	assert.Equal(t, "m=65536,t=3,p=4", parts[3])
	assert.NotEmpty(t, parts[4], "salt segment")
	assert.NotEmpty(t, parts[5], "key segment")
}

func TestArgon2PasswordHasher_HashVerify_RoundTrip(t *testing.T) {
	h := security.NewArgon2PasswordHasher()

	hash, err := h.Hash("correct horse battery staple")
	require.NoError(t, err)
	assert.NotEmpty(t, hash)

	assert.True(t, h.Verify("correct horse battery staple", hash))
}

func TestArgon2PasswordHasher_Verify_WrongPasswordFails(t *testing.T) {
	h := security.NewArgon2PasswordHasher()

	hash, err := h.Hash("correct horse battery staple")
	require.NoError(t, err)

	assert.False(t, h.Verify("wrong password", hash))
}

func TestArgon2PasswordHasher_Hash_DifferentPasswordsProduceDifferentHashes(t *testing.T) {
	h := security.NewArgon2PasswordHasher()

	hash1, err := h.Hash("password-one")
	require.NoError(t, err)
	hash2, err := h.Hash("password-two")
	require.NoError(t, err)

	assert.NotEqual(t, hash1, hash2)
}

func TestArgon2PasswordHasher_Hash_SamePasswordProducesDifferentHashesEachTime(t *testing.T) {
	// Random salt per call means even identical input never collides.
	h := security.NewArgon2PasswordHasher()

	hash1, err := h.Hash("same-password")
	require.NoError(t, err)
	hash2, err := h.Hash("same-password")
	require.NoError(t, err)

	assert.NotEqual(t, hash1, hash2)
	assert.True(t, h.Verify("same-password", hash1))
	assert.True(t, h.Verify("same-password", hash2))
}

func TestArgon2PasswordHasher_Verify_MalformedHashFailsWithoutPanic(t *testing.T) {
	h := security.NewArgon2PasswordHasher()

	tests := []struct {
		name string
		hash string
	}{
		{"empty string", ""},
		{"garbage", "not-a-hash-at-all"},
		{"wrong scheme", "$bcrypt$v=1$m=1,t=1,p=1$c2FsdA$aGFzaA"},
		{"too few parts", "$argon2id$v=19$m=65536,t=3,p=4$c2FsdA"},
		{"non-numeric version", "$argon2id$v=x$m=65536,t=3,p=4$c2FsdA$aGFzaA"},
		{"non-base64 salt", "$argon2id$v=19$m=65536,t=3,p=4$not-base64!$aGFzaA"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotPanics(t, func() {
				assert.False(t, h.Verify("anything", tt.hash))
			})
		})
	}
}
