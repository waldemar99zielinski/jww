package webauthn

import (
	"net/http"

	"jww/internal/adapter/schema"
	"jww/internal/domain"
	"jww/internal/infrastructure/http_server/kernel"
)

func loginFinishHandler(deps *kernel.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		appCtx := kernel.RequestContext(r)

		req, err := kernel.DecodeJSON[schema.WebAuthnLoginFinishRequest](r)
		if err != nil {
			kernel.WriteBadRequest(w, err)
			return
		}
		if err := req.Validate(); err != nil {
			kernel.WriteBadRequest(w, err)
			return
		}

		credentialResponse := domain.WebAuthnCeremonyResponse(req.Credential)
		token, err := deps.WebAuthnWorkflow.FinishLogin(appCtx, req.ChallengeToken, credentialResponse)
		if err != nil {
			kernel.WriteError(w, r, err)
			return
		}
		kernel.WriteJSON(w, http.StatusOK, schema.NewTokenResponse(token))
	}
}
