package scheduler

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"jww/internal/domain"
)

func TestNew_RejectsInvalidConfig(t *testing.T) {
	t.Run("invalid timezone", func(t *testing.T) {
		_, err := New(Config{Spec: "1 0 * * *", Timezone: "Not/AZone", TimeoutSeconds: 60}, func(domain.AppContext) error { return nil })
		assert.Error(t, err)
	})

	t.Run("invalid spec", func(t *testing.T) {
		_, err := New(Config{Spec: "not a cron spec", Timezone: "UTC", TimeoutSeconds: 60}, func(domain.AppContext) error { return nil })
		assert.Error(t, err)
	})
}

func TestScheduler_StartStopWithNoJobRunning(t *testing.T) {
	s, err := New(Config{Spec: "1 0 * * *", Timezone: "UTC", TimeoutSeconds: 60}, func(domain.AppContext) error {
		t.Error("job must not run during this test")
		return nil
	})
	require.NoError(t, err)

	s.Start()
	assert.NoError(t, s.Stop(context.Background()))
}

func TestScheduler_RunOnStartInvokesJobWithSchedulerContext(t *testing.T) {
	done := make(chan domain.AppContext, 1)
	s, err := New(Config{Spec: "1 0 * * *", Timezone: "UTC", TimeoutSeconds: 60, RunOnStart: true}, func(ctx domain.AppContext) error {
		done <- ctx
		return nil
	})
	require.NoError(t, err)

	s.Start()
	t.Cleanup(func() { _ = s.Stop(context.Background()) })

	select {
	case appCtx := <-done:
		source := appCtx.RequestSource()
		assert.Equal(t, domain.RequestSourceTypeScheduler, source.Type)
		require.NotNil(t, source.Scheduler)
		assert.Equal(t, "exchange_rate_sync", source.Scheduler.Job)
		assert.NotEmpty(t, appCtx.FlowID())
		assert.NotNil(t, appCtx.Context())
	case <-time.After(2 * time.Second):
		t.Fatal("job was not invoked on start")
	}
}
