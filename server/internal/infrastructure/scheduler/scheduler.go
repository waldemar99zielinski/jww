// Package scheduler runs jobs on a cron schedule. It is the only package that
// imports the cron library, so the rest of the app sees just the Job callback
// and a lifecycle of Start/Stop. It is also the boundary that logs a run's
// outcome: jobs return an error and services/adapters never log.
package scheduler

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/robfig/cron/v3"

	"jww/internal/domain"
	"jww/utils/configuration"
	"jww/utils/logger"
)

// jobName labels every run's log lines with a stable "job" attribute.
const jobName = "exchange_rate_sync"

var log = logger.New().WithComponent("scheduler")

// Job is the work a schedule runs. It receives a fresh AppContext carrying a
// source_type=scheduler, a new flow id, and a per-run timeout.
type Job func(ctx domain.AppContext) error

// Config is read from the EXCHANGE_RATE_SYNC_* env vars. Disabled by default:
// a background feature that makes outbound calls should be opt-in.
type Config struct {
	Enabled        bool   `env:"EXCHANGE_RATE_SYNC_ENABLED" bind:"optional" default:"false"`
	Spec           string `env:"EXCHANGE_RATE_SYNC_CRON" bind:"optional" default:"1 0 * * *"`
	Timezone       string `env:"EXCHANGE_RATE_SYNC_TZ" bind:"optional" default:"UTC"`
	TimeoutSeconds int64  `env:"EXCHANGE_RATE_SYNC_TIMEOUT_SECONDS" bind:"optional" default:"60"`
	RunOnStart     bool   `env:"EXCHANGE_RATE_SYNC_RUN_ON_START" bind:"optional" default:"false"`
}

func LoadConfig() (Config, error) {
	return configuration.Load[Config]()
}

// Scheduler wraps a cron schedule around one Job. Both the spec and the
// timezone are validated at construction, so a bad value fails startup rather
// than silently never running.
type Scheduler struct {
	cron       *cron.Cron
	job        Job
	timeout    time.Duration
	runOnStart bool
	wg         sync.WaitGroup
}

func New(cfg Config, job Job) (*Scheduler, error) {
	loc, err := time.LoadLocation(cfg.Timezone)
	if err != nil {
		return nil, fmt.Errorf("scheduler: invalid timezone %q: %w", cfg.Timezone, err)
	}
	s := &Scheduler{
		job:        job,
		timeout:    time.Duration(cfg.TimeoutSeconds) * time.Second,
		runOnStart: cfg.RunOnStart,
	}
	s.cron = cron.New(
		cron.WithLocation(loc),
		cron.WithLogger(cronLogger{}),
		// Recover turns a panic in the job into a logged error instead of a
		// crashed process; SkipIfStillRunning guarantees a slow run never
		// overlaps the next tick.
		cron.WithChain(cron.Recover(cronLogger{}), cron.SkipIfStillRunning(cronLogger{})),
	)
	if _, err := s.cron.AddFunc(cfg.Spec, s.run); err != nil {
		return nil, fmt.Errorf("scheduler: invalid spec %q: %w", cfg.Spec, err)
	}
	return s, nil
}

// Start begins scheduling and, when configured, kicks off one catch-up run
// immediately (safe because every run upserts on the business date).
func (s *Scheduler) Start() {
	if s.runOnStart {
		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			s.run()
		}()
	}
	s.cron.Start()
}

// Stop stops scheduling new runs and waits for any in-flight run — both the
// cron jobs and a boot catch-up — bounded by ctx (main.go's shutdown budget).
func (s *Scheduler) Stop(ctx context.Context) error {
	done := make(chan struct{})
	go func() {
		<-s.cron.Stop().Done()
		s.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *Scheduler) run() {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	var appCtx domain.AppContext
	appCtx.SetContext(ctx)
	appCtx.SetRequestSource(domain.RequestSource{
		Type:      domain.RequestSourceTypeScheduler,
		Scheduler: &domain.SchedulerRequestSource{Job: jobName},
	})
	appCtx.SetFlowID(uuid.NewString())

	started := time.Now()
	if err := s.job(appCtx); err != nil {
		log.WithContext(appCtx).Error("scheduled job failed",
			slog.String("error", err.Error()), slog.Duration("duration", time.Since(started)))
		return
	}
	log.WithContext(appCtx).Info("scheduled job completed", slog.Duration("duration", time.Since(started)))
}

// cronLogger adapts the shared slog logger to cron's Logger interface.
type cronLogger struct{}

func (cronLogger) Info(msg string, kv ...any) { log.Slog().Info(msg, kv...) }

func (cronLogger) Error(err error, msg string, kv ...any) {
	log.Slog().Error(msg, append(kv, slog.String("error", err.Error()))...)
}
