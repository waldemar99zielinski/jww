package user

import (
	"github.com/go-chi/chi/v5"

	"jww/internal/infrastructure/http_server/kernel"
)

func Mount(_, protected chi.Router, deps *kernel.Deps) {
	protected.Get("/v1/users", listUsersHandler(deps))
	protected.Get("/v1/users/{userID}", getUserHandler(deps))
	protected.Put("/v1/users/{userID}", updateUserHandler(deps))
	protected.Delete("/v1/users/{userID}", deleteUserHandler(deps))
}
