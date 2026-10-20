package webauthn

import (
	"github.com/go-webauthn/webauthn/protocol"
	libwebauthn "github.com/go-webauthn/webauthn/webauthn"

	"jww/internal/domain"
)

// webAuthnUser adapts a domain.User and its stored credentials to the
// library's webauthn.User interface. It exists only for the duration of one
// ceremony call — nothing here is persisted directly.
type webAuthnUser struct {
	user        domain.User
	credentials []domain.WebAuthnCredential
}

func (u webAuthnUser) WebAuthnID() []byte          { return u.user.ID[:] }
func (u webAuthnUser) WebAuthnName() string        { return u.user.Username }
func (u webAuthnUser) WebAuthnDisplayName() string { return u.user.DisplayName }
func (u webAuthnUser) WebAuthnCredentials() []libwebauthn.Credential {
	out := make([]libwebauthn.Credential, len(u.credentials))
	for i, c := range u.credentials {
		out[i] = toLibraryCredential(c)
	}
	return out
}

func toLibraryCredential(c domain.WebAuthnCredential) libwebauthn.Credential {
	transports := make([]protocol.AuthenticatorTransport, len(c.Transports))
	for i, t := range c.Transports {
		transports[i] = protocol.AuthenticatorTransport(t)
	}
	return libwebauthn.Credential{
		ID:              c.CredentialID,
		PublicKey:       c.PublicKey,
		AttestationType: c.AttestationType,
		Transport:       transports,
		Flags: libwebauthn.CredentialFlags{
			BackupEligible: c.BackupEligible,
		},
		Authenticator: libwebauthn.Authenticator{
			AAGUID:    c.AAGUID,
			SignCount: c.SignCount,
		},
	}
}

// fromLibraryCredential extracts the fields this app persists from a
// library Credential produced by a successful Begin/Finish ceremony.
func fromLibraryCredential(cred *libwebauthn.Credential) (publicKey, aaguid []byte, attestationType string, signCount uint32, transports []string, backupEligible bool) {
	transports = make([]string, len(cred.Transport))
	for i, t := range cred.Transport {
		transports[i] = string(t)
	}
	return cred.PublicKey, cred.Authenticator.AAGUID, cred.AttestationType, cred.Authenticator.SignCount, transports, cred.Flags.BackupEligible
}
