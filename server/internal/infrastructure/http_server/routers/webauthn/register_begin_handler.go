package webauthn

import (
	"encoding/json"
	"net/http"

	"jww/internal/adapter/schema"
	"jww/internal/infrastructure/http_server/kernel"
)

// registerBeginHandler runs behind kernel.RequireAuth, so the authenticated
// user is already resolved — enrollment always attaches to the caller's own
// account, never a request body field.
func registerBeginHandler(deps *kernel.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		appCtx := kernel.RequestContext(r)

		options, challengeToken, err := deps.WebAuthnWorkflow.BeginRegistration(appCtx, appCtx.User().ID)
		if err != nil {
			kernel.WriteError(w, r, err)
			return
		}
		kernel.WriteJSON(w, http.StatusOK, schema.WebAuthnRegisterBeginResponse{
			Options:        json.RawMessage(options),
			ChallengeToken: challengeToken,
		})
	}
}
