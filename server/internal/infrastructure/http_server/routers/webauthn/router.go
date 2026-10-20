package webauthn

import (
	"github.com/go-chi/chi/v5"

	"jww/internal/infrastructure/http_server/kernel"
)

// Mount registers every WebAuthn route. Enrollment (register/*) is always
// protected — only an already-logged-in (password-verified) user can attach
// a credential to their own account. Only the login ceremony is public,
// since the whole point is signing in without a password.
func Mount(public, protected chi.Router, deps *kernel.Deps) {
	protected.Post("/v1/webauthn/credentials/register/begin", registerBeginHandler(deps))
	protected.Post("/v1/webauthn/credentials/register/finish", registerFinishHandler(deps))
	protected.Get("/v1/webauthn/credentials", listCredentialsHandler(deps))
	protected.Delete("/v1/webauthn/credentials/{credentialID}", deleteCredentialHandler(deps))

	public.Post("/v1/webauthn/login/begin", loginBeginHandler(deps))
	public.Post("/v1/webauthn/login/finish", loginFinishHandler(deps))
}
