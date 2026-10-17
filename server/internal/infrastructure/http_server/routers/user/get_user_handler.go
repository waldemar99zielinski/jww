package user

import (
	"net/http"

	"jww/internal/adapter/schema"
	"jww/internal/infrastructure/http_server/kernel"
	httputils "jww/internal/infrastructure/http_server/utils"
)

func getUserHandler(deps *kernel.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		appCtx := kernel.RequestContext(r)

		userID, err := httputils.ParseUUIDPathParam(r, "userID")
		if err != nil {
			kernel.WriteBadRequest(w, err)
			return
		}

		u, err := deps.UserService.GetByID(appCtx, userID)
		if err != nil {
			kernel.WriteError(w, err)
			return
		}
		kernel.WriteJSON(w, http.StatusOK, schema.UserToResponse(*u))
	}
}
