package webauthn

import (
	"net/http"

	"jww/internal/adapter/schema"
	"jww/internal/infrastructure/http_server/kernel"
)

func listCredentialsHandler(deps *kernel.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		appCtx := kernel.RequestContext(r)

		credentials, err := deps.WebAuthnWorkflow.ListCredentials(appCtx, appCtx.User().ID)
		if err != nil {
			kernel.WriteError(w, r, err)
			return
		}
		out := make([]schema.WebAuthnCredentialResponse, len(credentials))
		for i, c := range credentials {
			out[i] = schema.WebAuthnCredentialToResponse(c)
		}
		kernel.WriteJSON(w, http.StatusOK, out)
	}
}
