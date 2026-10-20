package webauthn

import (
	"net/http"

	"jww/internal/adapter/schema"
	"jww/internal/domain"
	"jww/internal/infrastructure/http_server/kernel"
)

func registerFinishHandler(deps *kernel.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		appCtx := kernel.RequestContext(r)

		req, err := kernel.DecodeJSON[schema.WebAuthnRegisterFinishRequest](r)
		if err != nil {
			kernel.WriteBadRequest(w, err)
			return
		}
		if err := req.Validate(); err != nil {
			kernel.WriteBadRequest(w, err)
			return
		}

		credentialResponse := domain.WebAuthnCeremonyResponse(req.Credential)
		credential, err := deps.WebAuthnWorkflow.FinishRegistration(appCtx, appCtx.User().ID, req.ChallengeToken, req.Name, credentialResponse)
		if err != nil {
			kernel.WriteError(w, r, err)
			return
		}
		kernel.WriteCreated(w, schema.WebAuthnCredentialToResponse(*credential))
	}
}
