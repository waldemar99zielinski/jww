package health

import (
	"net/http"

	"jww/internal/adapter/schema"
	"jww/internal/infrastructure/http_server/kernel"
)

func getReadinessHandler(deps *kernel.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		report := deps.HealthService.CheckReadiness(kernel.RequestContext(r))

		w.Header().Set("Cache-Control", "no-store")
		if !report.Ready {
			kernel.WriteServiceUnavailable(w)
			return
		}
		kernel.WriteOK(w, schema.ReadinessToResponse(report))
	}
}
