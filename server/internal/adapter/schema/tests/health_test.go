package schema_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"jww/internal/adapter/schema"
	healthservice "jww/internal/service/health"
)

func TestHealthToResponse(t *testing.T) {
	response := schema.HealthToResponse(healthservice.HealthReport{Status: healthservice.HealthStatusOK})

	assert.Equal(t, schema.HealthResponse{Status: healthservice.HealthStatusOK}, response)
}

func TestReadinessToResponse(t *testing.T) {
	tests := []struct {
		name   string
		report healthservice.ReadinessReport
		want   string
	}{
		{name: "ready", report: healthservice.ReadinessReport{Ready: true}, want: "ok"},
		{name: "not ready", report: healthservice.ReadinessReport{Ready: false}, want: "not_ok"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := schema.ReadinessToResponse(tt.report)
			assert.Equal(t, schema.ReadinessResponse{Status: tt.want}, response)
		})
	}
}

func TestHealthResponses_JSONShape(t *testing.T) {
	healthBody, err := json.Marshal(schema.HealthToResponse(healthservice.HealthReport{Status: healthservice.HealthStatusOK}))
	require.NoError(t, err)
	assert.JSONEq(t, `{"status":"ok"}`, string(healthBody))

	readinessBody, err := json.Marshal(schema.ReadinessToResponse(healthservice.ReadinessReport{Ready: true}))
	require.NoError(t, err)
	assert.JSONEq(t, `{"status":"ok"}`, string(readinessBody))
}
