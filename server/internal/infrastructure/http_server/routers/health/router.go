// Package health exposes the server's unauthenticated health and readiness
// routes.
package health

import (
	"github.com/go-chi/chi/v5"

	"jww/internal/infrastructure/http_server/kernel"
)

func Mount(public, _ chi.Router, deps *kernel.Deps) {
	public.Get("/v1/health", getHealthHandler(deps))
	public.Get("/v1/readiness", getReadinessHandler(deps))
}
