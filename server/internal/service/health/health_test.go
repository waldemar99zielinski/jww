package health_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"jww/internal/domain"
	"jww/internal/service/health"
)

func appContext() domain.AppContext {
	var appCtx domain.AppContext
	appCtx.SetContext(context.Background())
	return appCtx
}

func newService(databaseErr, optionalErr error) *health.Service {
	return health.NewService(
		health.Config{LatencyWarnMS: 1000},
		health.Dependency{Name: "database", Required: true, Checker: health.CheckFunc(func(domain.AppContext) error { return databaseErr })},
		health.Dependency{Name: "openai", Checker: health.CheckFunc(func(domain.AppContext) error { return optionalErr })},
	)
}

func TestService_CheckHealth_DoesNotCheckDependencies(t *testing.T) {
	checked := false
	service := health.NewService(health.Config{}, health.Dependency{
		Name: "database",
		Checker: health.CheckFunc(func(domain.AppContext) error {
			checked = true
			return errors.New("offline")
		}),
	})

	report := service.CheckHealth(appContext())

	assert.Equal(t, health.HealthStatusOK, report.Status)
	assert.False(t, checked, "liveness must not perform dependency I/O")
}

func TestService_CheckReadiness(t *testing.T) {
	tests := []struct {
		name        string
		databaseErr error
		optionalErr error
		wantStatus  string
		wantReady   bool
	}{
		{name: "all available", wantStatus: health.ReadinessStatusReady, wantReady: true},
		{name: "optional dependency unavailable", optionalErr: errors.New("offline"), wantStatus: health.ReadinessStatusDegraded, wantReady: true},
		{name: "required dependency unavailable", databaseErr: errors.New("offline"), wantStatus: health.ReadinessStatusNotReady, wantReady: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report := newService(tt.databaseErr, tt.optionalErr).CheckReadiness(appContext())

			assert.Equal(t, tt.wantStatus, report.Status)
			assert.Equal(t, tt.wantReady, report.Ready)
			require.Len(t, report.Dependencies, 2)
			assert.Equal(t, "database", report.Dependencies[0].Name)
			assert.Equal(t, "openai", report.Dependencies[1].Name)
		})
	}
}

func TestService_CheckReadiness_RecordsDependencyLatency(t *testing.T) {
	service := health.NewService(health.Config{}, health.Dependency{
		Name: "slow",
		Checker: health.CheckFunc(func(domain.AppContext) error {
			time.Sleep(5 * time.Millisecond)
			return nil
		}),
	})

	report := service.CheckReadiness(appContext())

	require.Len(t, report.Dependencies, 1)
	assert.GreaterOrEqual(t, report.Dependencies[0].Latency, 5*time.Millisecond)
}
