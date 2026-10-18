package user

import (
	"net/http"

	"jww/internal/infrastructure/http_server/kernel"
	httputils "jww/internal/infrastructure/http_server/utils"
)

func deleteUserHandler(deps *kernel.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		appCtx := kernel.RequestContext(r)

		userID, err := httputils.ParseUUIDPathParam(r, "userID")
		if err != nil {
			kernel.WriteBadRequest(w, err)
			return
		}

		u, err := deps.UserService.GetByID(appCtx, userID)
		if err != nil {
			kernel.WriteError(w, r, err)
			return
		}
		if err := deps.UserService.Delete(appCtx, *u); err != nil {
			kernel.WriteError(w, r, err)
			return
		}
		kernel.WriteNoContent(w)
	}
}
