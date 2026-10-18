package user

import (
	"net/http"

	"jww/internal/adapter/schema"
	"jww/internal/infrastructure/http_server/kernel"
)

func listUsersHandler(deps *kernel.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		appCtx := kernel.RequestContext(r)

		users, err := deps.UserService.List(appCtx)
		if err != nil {
			kernel.WriteError(w, r, err)
			return
		}
		out := make([]schema.UserResponse, len(users))
		for i, u := range users {
			out[i] = schema.UserToResponse(u)
		}
		kernel.WriteJSON(w, http.StatusOK, out)
	}
}
