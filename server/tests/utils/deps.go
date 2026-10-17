package utils

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"jww/internal/adapter/gateway"
	"jww/internal/adapter/security"
	"jww/internal/adapter/storage"
	"jww/internal/domain"
	"jww/internal/infrastructure/database"
	"jww/internal/infrastructure/http_server/kernel"
	"jww/internal/service/auth"
	"jww/internal/service/currency"
	currency_exchange_rate "jww/internal/service/currency_exchange_rate"
	"jww/internal/service/object_storage"
	"jww/internal/service/transaction"
	"jww/internal/service/transaction_attachment"
	"jww/internal/service/transaction_category"
	"jww/internal/service/user"
	"jww/internal/service/workflow"
)

// stubRateSource satisfies currency_exchange_rate.RateSource for the in-process
// journey tests. They exercise HTTP routes and never call SyncDaily, so it
// returns nothing and never errors.
type stubRateSource struct{}

func (stubRateSource) Rates(domain.AppContext, []string, time.Time) (currency_exchange_rate.RateFetch, error) {
	return currency_exchange_rate.RateFetch{}, nil
}

// BuildTestDeps wires every repository, service, and workflow against pool —
// the same construction cmd/server/deps.go's buildDeps does. Duplicated here
// (rather than imported) because buildDeps is unexported in package main;
// keep this in sync if that wiring changes — see cmd/server/README.md.
//
// storageConfig points the attachment storage adapter at the test MinIO
// instance (see internal/tests/s3test).
func BuildTestDeps(ctx context.Context, pool *pgxpool.Pool, jwtSecret string, storageConfig storage.Config) (*kernel.Deps, error) {
	uowManager := database.NewPostgresUnitOfWorkManager(pool)
	passwordHasher := security.NewArgon2PasswordHasher()
	tokenService := security.NewJwtTokenService(security.Config{
		SecretKey:                jwtSecret,
		Algorithm:                "HS256",
		AccessTokenExpireMinutes: 60,
	})

	userRepo := gateway.NewPostgresUserRepository(pool)
	currencyRepo := gateway.NewPostgresCurrencyRepository(pool)
	exchangeRateRepo := gateway.NewPostgresCurrencyExchangeRateRepository(pool)
	categoryRepo := gateway.NewPostgresTransactionCategoryRepository(pool)
	transactionRepo := gateway.NewPostgresTransactionRepository(pool)
	attachmentRepo := gateway.NewPostgresTransactionAttachmentRepository(pool)

	objectStore, err := storage.NewS3Store(ctx, storageConfig)
	if err != nil {
		return nil, err
	}

	authService := auth.NewService(userRepo, passwordHasher, tokenService)
	userService := user.NewService(userRepo, passwordHasher, uowManager)
	currencyService := currency.NewService(currencyRepo, uowManager)
	exchangeRateService := currency_exchange_rate.NewService(exchangeRateRepo)
	categoryService := transaction_category.NewService(categoryRepo, uowManager)
	transactionService := transaction.NewService(transactionRepo, uowManager)
	attachmentStore := object_storage.Namespaced(objectStore, object_storage.StorageTransactionAttachmentsNamespace)
	attachmentService := transaction_attachment.NewService(attachmentRepo, attachmentStore, reportOrphanedObject)

	return &kernel.Deps{
		AuthService:                  authService,
		UserService:                  userService,
		SignupWorkflow:               workflow.NewSignupWorkflow(userService),
		CurrencyService:              currencyService,
		CurrencyExchangeRateWorkflow: workflow.NewCurrencyExchangeRateWorkflow(currencyService, exchangeRateService, stubRateSource{}, uowManager),
		TransactionCategoryService:   categoryService,
		TransactionService:           transactionService,
		TransactionAttachmentService: attachmentService,
		TransactionWorkflow:          workflow.NewTransactionWorkflow(currencyService, categoryService, userService, transactionService, attachmentService, uowManager),
		SummaryWorkflow:              workflow.NewSummaryWorkflow(currencyService, userService, transactionService, exchangeRateService, uowManager),
		CorsOrigins:                  []string{"*"},
		MaxBodyBytes:                 1 << 20,
	}, nil
}

// reportOrphanedObject mirrors cmd/server's reporter. Tests don't capture logs;
// the reporter exists so the wiring matches production.
func reportOrphanedObject(_ domain.AppContext, _ uuid.UUID, _ error) {}
