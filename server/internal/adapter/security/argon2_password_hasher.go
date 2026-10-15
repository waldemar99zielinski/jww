// Package security implements the auth/user PasswordHasher and auth
// TokenService ports — the only place that knows passwords are hashed with
// Argon2id and tokens are signed JWTs.
package security

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// Argon2PasswordHasher satisfies both service/auth.PasswordHasher and
// service/user.PasswordHasher — identical shapes, one concrete
// implementation, same as the Postgres gateways below.
type Argon2PasswordHasher struct {
	memoryKiB  uint32
	iterations uint32
	threads    uint8
	saltLen    uint32
	keyLen     uint32
}

// NewArgon2PasswordHasher uses the common Argon2id defaults (memory=64MiB,
// time=3, parallelism=4) so existing password hashes keep verifying across
// deployments that carried data forward from an earlier hasher config.
func NewArgon2PasswordHasher() *Argon2PasswordHasher {
	return &Argon2PasswordHasher{memoryKiB: 65536, iterations: 3, threads: 4, saltLen: 16, keyLen: 32}
}

func (h *Argon2PasswordHasher) Hash(password string) (string, error) {
	salt := make([]byte, h.saltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("argon2: generating salt: %w", err)
	}
	key := argon2.IDKey([]byte(password), salt, h.iterations, h.memoryKiB, h.threads, h.keyLen)
	encoded := fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, h.memoryKiB, h.iterations, h.threads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(key),
	)
	return encoded, nil
}

func (h *Argon2PasswordHasher) Verify(password, passwordHash string) bool {
	parts := strings.Split(passwordHash, "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		return false
	}
	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil {
		return false
	}
	var memoryKiB, iterations uint32
	var threads uint8
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memoryKiB, &iterations, &threads); err != nil {
		return false
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false
	}
	want, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false
	}
	got := argon2.IDKey([]byte(password), salt, iterations, memoryKiB, threads, uint32(len(want)))
	return subtle.ConstantTimeCompare(got, want) == 1
}
