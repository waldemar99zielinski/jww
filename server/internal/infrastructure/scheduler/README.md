# internal/infrastructure/scheduler

**Clean Architecture layer: Frameworks & Drivers.** The **only** package that imports the cron library.

**Role.** Runs a `Job` (a `func(domain.AppContext) error`) on a cron schedule. Wraps [`robfig/cron/v3`](https://github.com/robfig/cron) so the rest of the app sees just a callback plus a `Start`/`Stop` lifecycle. Also the boundary that logs each run's outcome — services and adapters never log. Owns `scheduler.Config` (`EXCHANGE_RATE_SYNC_*`).

## Files

- `scheduler.go` (package doc) — `Config{ Enabled, Spec, Timezone, TimeoutSeconds, RunOnStart }`; `LoadConfig()`; `Job`; `Scheduler` + `New(cfg, job)`; `Start()`; `Stop(ctx)`.
- `New` validates the timezone (`time.LoadLocation`) and the cron spec (`cron.AddFunc`) at construction, so a bad value fails startup instead of silently never running.
- `run()` builds a fresh `domain.AppContext` per run: `source_type=scheduler`, `job=exchange_rate_sync`, a new `flow_id`, and a per-run timeout. It logs the run at **ERROR** (with `error` + `duration`) on failure, or **INFO** (`duration`) on success.
- Cron is configured with `WithLocation` (from `Timezone`), `WithLogger`, and `WithChain(Recover, SkipIfStillRunning)` — a panic becomes a logged error, and a slow run never overlaps the next tick.
- `Start` also kicks off one catch-up run when `RunOnStart` is set; `Stop(ctx)` stops new runs and waits for any in-flight run (cron jobs + the boot run) bounded by `ctx` — `main.go` shares its 10s shutdown budget with `http.Server.Shutdown`.

## Rules

- Keep the cron library confined to this package; nothing else should import it.
- Config fields use `env`/`bind`/`default` tags and `LoadConfig()`; document new vars in `server/.env.example`.
- This is a **boundary** for logging (component `scheduler`); the job closure itself is supplied by `cmd/server`, which logs the per-run detail (warnings + summary).
- Imports: stdlib, `robfig/cron/v3`, `uuid`, `domain`, `utils/{configuration,logger}`.

## Testing requirement for this directory

White-box (`package scheduler`, alongside the source) — no I/O, no real
clock:
`New` rejects an invalid timezone and an invalid spec; `Start`/`Stop` return
cleanly with no job running; a `RunOnStart` scheduler invokes the job with a
`source_type=scheduler` `AppContext` carrying the job name and a flow id.
