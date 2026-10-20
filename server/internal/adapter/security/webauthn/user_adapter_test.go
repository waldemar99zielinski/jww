package webauthn

import (
	"testing"

	libwebauthn "github.com/go-webauthn/webauthn/webauthn"
	"github.com/stretchr/testify/assert"

	"jww/internal/domain"
)

// TestToLibraryCredential_RoundTripsBackupEligible is a regression test: an
// earlier version never set Flags at all, so every stored credential's
// BackupEligible silently read back as false. go-webauthn hard-rejects a
// login whose assertion reports BE=true (true for any iCloud Keychain /
// platform-synced passkey, e.g. Face ID) against a stored record saying
// BE=false — "Backup Eligible flag inconsistency detected during login
// validation" — so this must round-trip exactly, not just compile.
func TestToLibraryCredential_RoundTripsBackupEligible(t *testing.T) {
	for _, backupEligible := range []bool{true, false} {
		got := toLibraryCredential(domain.WebAuthnCredential{
			CredentialID:   []byte("cred-1"),
			PublicKey:      []byte("pub-1"),
			BackupEligible: backupEligible,
		})
		assert.Equal(t, backupEligible, got.Flags.BackupEligible)
	}
}

func TestFromLibraryCredential_ExtractsBackupEligible(t *testing.T) {
	for _, backupEligible := range []bool{true, false} {
		cred := &libwebauthn.Credential{
			ID:        []byte("cred-1"),
			PublicKey: []byte("pub-1"),
			Flags:     libwebauthn.CredentialFlags{BackupEligible: backupEligible},
		}
		_, _, _, _, _, gotBackupEligible := fromLibraryCredential(cred)
		assert.Equal(t, backupEligible, gotBackupEligible)
	}
}
