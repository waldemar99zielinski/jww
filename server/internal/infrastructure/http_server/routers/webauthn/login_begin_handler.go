package webauthn

import (
	"encoding/json"
	"net/http"

	"jww/internal/adapter/schema"
	"jww/internal/infrastructure/http_server/kernel"
)

func loginBeginHandler(deps *kernel.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		appCtx := kernel.RequestContext(r)

		options, challengeToken, err := deps.WebAuthnWorkflow.BeginLogin(appCtx)
		if err != nil {
			kernel.WriteError(w, r, err)
			return
		}
		kernel.WriteJSON(w, http.StatusOK, schema.WebAuthnLoginBeginResponse{
			Options:        json.RawMessage(options),
			ChallengeToken: challengeToken,
		})
	}
}
