package webauthn

import (
	"net/http"

	"jww/internal/infrastructure/http_server/kernel"
	httputils "jww/internal/infrastructure/http_server/utils"
)

func deleteCredentialHandler(deps *kernel.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		appCtx := kernel.RequestContext(r)

		credentialID, err := httputils.ParseUUIDPathParam(r, "credentialID")
		if err != nil {
			kernel.WriteBadRequest(w, err)
			return
		}

		if err := deps.WebAuthnWorkflow.DeleteCredential(appCtx, appCtx.User().ID, credentialID); err != nil {
			kernel.WriteError(w, r, err)
			return
		}
		kernel.WriteNoContent(w)
	}
}
