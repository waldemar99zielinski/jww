// Command bootstrap prepares a complete local database in one run: migrations,
// fixed users/currencies, two months of daily rates, 5,000 transactions, and
// representative PDF/photo attachments in object storage.
package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"jww/internal/adapter/storage"
	"jww/internal/infrastructure/database"
	"jww/internal/service/object_storage"
	"jww/scripts/bootstrap"
	"jww/utils/logger"
)

var log = logger.New().WithComponent("scripts.bootstrap")

type params struct {
	databaseURL  string
	transactions int
	attachments  int
	randomSeed   int64
}

func main() {
	if err := loadDotEnv(".env"); err != nil {
		log.Slog().Error("bootstrap failed", slog.String("error", fmt.Sprintf("loading .env: %v", err)))
		os.Exit(1)
	}
	p, err := parseParams(os.Args[1:], os.Getenv)
	if err != nil {
		log.Slog().Error("bootstrap failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	if err := run(p); err != nil {
		log.Slog().Error("bootstrap failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

// loadDotEnv gives this local-development command the same run-from-server
// convenience as cmd/server. Explicit process environment values always win.
func loadDotEnv(path string) error {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, found := strings.Cut(line, "=")
		if !found {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.Trim(strings.TrimSpace(value), `"'`)
		if _, exists := os.LookupEnv(key); !exists {
			if err := os.Setenv(key, value); err != nil {
				return err
			}
		}
	}
	return scanner.Err()
}

func parseParams(args []string, getenv func(string) string) (params, error) {
	defaults := bootstrap.DefaultOptions()
	transactionCount, err := envInt(getenv, "BOOTSTRAP_TRANSACTIONS", defaults.TransactionCount)
	if err != nil {
		return params{}, err
	}
	attachmentCount, err := envInt(getenv, "BOOTSTRAP_ATTACHMENTS", defaults.AttachmentCount)
	if err != nil {
		return params{}, err
	}
	randomSeed, err := envInt64(getenv, "BOOTSTRAP_RANDOM_SEED", defaults.RandomSeed)
	if err != nil {
		return params{}, err
	}

	p := params{transactions: transactionCount, attachments: attachmentCount, randomSeed: randomSeed}
	fs := flag.NewFlagSet("bootstrap", flag.ContinueOnError)
	fs.StringVar(&p.databaseURL, "database-url", getenv("DATABASE_URL"), "Postgres DSN (or DATABASE_URL)")
	fs.IntVar(&p.transactions, "transactions", transactionCount, "number of generated transactions (or BOOTSTRAP_TRANSACTIONS)")
	fs.IntVar(&p.attachments, "attachments", attachmentCount, "number of generated attachments (or BOOTSTRAP_ATTACHMENTS)")
	fs.Int64Var(&p.randomSeed, "random-seed", randomSeed, "deterministic random seed (or BOOTSTRAP_RANDOM_SEED)")
	if err := fs.Parse(args); err != nil {
		return p, err
	}
	if p.databaseURL == "" {
		return p, fmt.Errorf("missing required value: -database-url or DATABASE_URL")
	}
	if p.transactions < 0 || p.attachments < 0 || p.attachments > p.transactions {
		return p, fmt.Errorf("counts must be non-negative and attachments must not exceed transactions")
	}
	return p, nil
}

func envInt(getenv func(string) string, name string, fallback int) (int, error) {
	raw := getenv(name)
	if raw == "" {
		return fallback, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("invalid %s %q: %w", name, raw, err)
	}
	return value, nil
}

func envInt64(getenv func(string) string, name string, fallback int64) (int64, error) {
	raw := getenv(name)
	if raw == "" {
		return fallback, nil
	}
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid %s %q: %w", name, raw, err)
	}
	return value, nil
}

func run(p params) error {
	loggerConfig, err := logger.LoadConfig()
	if err != nil {
		return err
	}
	if err := logger.Configure(loggerConfig); err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	pool, err := database.NewPool(ctx, database.Config{URL: p.databaseURL})
	if err != nil {
		return err
	}
	defer pool.Close()
	if err := database.Migrate(ctx, pool); err != nil {
		return err
	}

	var attachmentStore object_storage.Store
	if p.attachments > 0 {
		storageConfig, err := storage.LoadConfig()
		if err != nil {
			return err
		}
		baseStore, err := storage.NewS3Store(ctx, storageConfig)
		if err != nil {
			return err
		}
		attachmentStore = object_storage.Namespaced(baseStore, object_storage.StorageTransactionAttachmentsNamespace)
	}

	opts := bootstrap.DefaultOptions()
	opts.TransactionCount = p.transactions
	opts.AttachmentCount = p.attachments
	opts.RandomSeed = p.randomSeed
	result, err := bootstrap.Seed(ctx, pool, attachmentStore, opts)
	if err != nil {
		return err
	}
	log.Slog().Info("bootstrap complete",
		slog.Int("rate_days", result.RateDays),
		slog.Int("rates_inserted", result.RatesInserted),
		slog.Int("rates_updated", result.RatesUpdated),
		slog.Int("transactions_found", result.TransactionsFound),
		slog.Int("transactions_added", result.TransactionsAdded),
		slog.Int("attachments_added", result.AttachmentsAdded),
	)
	return nil
}
