package user

import (
	"net/http"

	"jww/internal/adapter/schema"
	"jww/internal/infrastructure/http_server/kernel"
	httputils "jww/internal/infrastructure/http_server/utils"
)

func updateUserHandler(deps *kernel.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		appCtx := kernel.RequestContext(r)

		userID, err := httputils.ParseUUIDPathParam(r, "userID")
		if err != nil {
			kernel.WriteBadRequest(w, err)
			return
		}
		req, err := kernel.DecodeJSON[schema.UserUpdateRequest](r)
		if err != nil {
			kernel.WriteBadRequest(w, err)
			return
		}
		if err := req.Validate(); err != nil {
			kernel.WriteBadRequest(w, err)
			return
		}

		existing, err := deps.UserService.GetByID(appCtx, userID)
		if err != nil {
			kernel.WriteError(w, err)
			return
		}
		updated := *existing
		updated.DisplayName = req.DisplayName
		if _, err := deps.UserService.Update(appCtx, updated); err != nil {
			kernel.WriteError(w, err)
			return
		}
		kernel.WriteJSON(w, http.StatusOK, schema.UserToResponse(updated))
	}
}
