// Command server is the process entry point. It loads configuration, builds
// the dependency graph, and asks the http_server/server package for a
// fully-mounted HTTP server to run.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"jww/internal/infrastructure/database"
	httpserver "jww/internal/infrastructure/http_server"
	"jww/utils/logger"
)

var log = logger.New().WithComponent("cmd.server")

func main() {
	if err := run(); err != nil {
		log.Slog().Error("server stopped", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func run() error {
	if err := loadDotEnv(".env"); err != nil {
		return fmt.Errorf("loading .env: %w", err)
	}

	cfg, err := LoadConfig()
	if err != nil {
		return err
	}

	if err := logger.Configure(cfg.Logger); err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := database.NewPool(ctx, cfg.Database)
	if err != nil {
		return err
	}
	defer pool.Close()

	if err := database.Migrate(ctx, pool); err != nil {
		return err
	}

	deps, err := buildDeps(ctx, pool, cfg.JWT, cfg.WebAuthn, cfg.Frankfurter, cfg.Storage, cfg.OpenAI, cfg.Extraction, cfg.Http.CorsOriginsList(), cfg.Http.MaxBodyBytes, cfg.Health)
	if err != nil {
		return err
	}

	stopScheduler, err := startScheduler(cfg.Scheduler, deps.CurrencyExchangeRateWorkflow)
	if err != nil {
		return err
	}

	srv, err := httpserver.New(deps, cfg.Http, cfg.Web)
	if err != nil {
		return fmt.Errorf("building HTTP server: %w", err)
	}

	errCh := make(chan error, 1)
	go func() {
		log.Slog().Info("listening", slog.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	stopScheduler(shutdownCtx)
	return srv.Shutdown(shutdownCtx)
}
