package user

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"jww/internal/infrastructure/http_server/kernel"
)

func deleteUserHandler(deps *kernel.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		appCtx := kernel.RequestContext(r)

		userID, err := uuid.Parse(chi.URLParam(r, "userID"))
		if err != nil {
			kernel.WriteBadRequest(w, err)
			return
		}

		u, err := deps.UserService.GetByID(appCtx, userID)
		if err != nil {
			kernel.WriteError(w, err)
			return
		}
		if err := deps.UserService.Delete(appCtx, *u); err != nil {
			kernel.WriteError(w, err)
			return
		}
		kernel.WriteNoContent(w)
	}
}
