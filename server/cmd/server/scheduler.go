package main

import (
	"context"
	"log/slog"

	"jww/internal/domain"
	"jww/internal/infrastructure/scheduler"
	"jww/internal/service/workflow"
)

// startScheduler builds the daily exchange-rate sync scheduler and starts it
// when enabled. It returns a stop function for graceful shutdown, which is a
// no-op when the scheduler was never started.
//
// This is the boundary that logs scheduled runs: the job closure runs the
// workflow's SyncDaily and logs every warning at WARN and the run summary at
// INFO. The workflow and the Frankfurter adapter return their warnings/errors;
// they never log.
func startScheduler(cfg scheduler.Config, rateWorkflow *workflow.CurrencyExchangeRateWorkflow) (stop func(context.Context), err error) {
	sched, err := scheduler.New(cfg, func(appCtx domain.AppContext) error {
		result, err := rateWorkflow.SyncDaily(appCtx)
		if err != nil {
			return err // the scheduler logs it at ERROR
		}
		// Never fail silently: log every skip/drop at WARN, then the summary.
		for _, warning := range result.Warnings {
			log.WithContext(appCtx).Warn("exchange rate sync warning", slog.String("detail", warning))
		}
		log.WithContext(appCtx).Info("exchange rates synced",
			slog.Int("currencies", result.Currencies),
			slog.Int("fetched", result.Fetched),
			slog.Int("inserted", result.Inserted),
			slog.Int("updated", result.Updated),
			slog.Int("skipped", result.Skipped))
		return nil
	})
	if err != nil {
		return nil, err
	}

	if cfg.Enabled {
		sched.Start()
		log.Slog().Info("exchange rate sync scheduled",
			slog.String("spec", cfg.Spec), slog.String("timezone", cfg.Timezone))
	}

	return func(ctx context.Context) {
		if err := sched.Stop(ctx); err != nil {
			log.Slog().Warn("scheduler shutdown", slog.String("error", err.Error()))
		}
	}, nil
}
