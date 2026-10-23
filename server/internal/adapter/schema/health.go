package schema

import healthservice "jww/internal/service/health"

const (
	readinessStatusOK    = "ok"
	readinessStatusNotOK = "not_ok"
)

// HealthResponse is the public process-liveness body. It deliberately carries
// no dependency details.
type HealthResponse struct {
	Status string `json:"status"`
}

// ReadinessResponse is the public readiness body. Dependency status, latency,
// and errors stay server-side and are emitted as warning logs instead.
type ReadinessResponse struct {
	Status string `json:"status"`
}

func HealthToResponse(report healthservice.HealthReport) HealthResponse {
	return HealthResponse{Status: report.Status}
}

func ReadinessToResponse(report healthservice.ReadinessReport) ReadinessResponse {
	status := readinessStatusOK
	if !report.Ready {
		status = readinessStatusNotOK
	}
	return ReadinessResponse{Status: status}
}
