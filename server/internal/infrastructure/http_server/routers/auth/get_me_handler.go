package auth

import (
	"net/http"

	"jww/internal/adapter/schema"
	"jww/internal/infrastructure/http_server/kernel"
)

// getMeHandler runs behind kernel.RequireAuth, so the authenticated user
// is already resolved and stored on the request context.
func getMeHandler(_ *kernel.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		appCtx := kernel.RequestContext(r)

		kernel.WriteJSON(w, http.StatusOK, schema.UserToResponse(*appCtx.User()))
	}
}
