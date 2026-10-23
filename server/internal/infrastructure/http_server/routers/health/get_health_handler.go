package health

import (
	"net/http"

	"jww/internal/adapter/schema"
	"jww/internal/infrastructure/http_server/kernel"
)

func getHealthHandler(deps *kernel.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		report := deps.HealthService.CheckHealth(kernel.RequestContext(r))

		w.Header().Set("Cache-Control", "no-store")
		kernel.WriteOK(w, schema.HealthToResponse(report))
	}
}
