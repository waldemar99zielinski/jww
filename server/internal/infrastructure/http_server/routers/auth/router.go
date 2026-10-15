package auth

import (
	"github.com/go-chi/chi/v5"

	"jww/internal/infrastructure/http_server/kernel"
)

func Mount(public, protected chi.Router, deps *kernel.Deps) {
	public.Post("/v1/auth/login", loginHandler(deps))

	protected.Get("/v1/auth/me", getMeHandler(deps))
}
