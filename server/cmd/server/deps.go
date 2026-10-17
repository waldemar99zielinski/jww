package main

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"jww/internal/adapter/frankfurter"
	"jww/internal/adapter/gateway"
	"jww/internal/adapter/security"
	"jww/internal/adapter/storage"
	"jww/internal/domain"
	"jww/internal/infrastructure/database"
	"jww/internal/infrastructure/http_server/kernel"
	"jww/internal/service/auth"
	"jww/internal/service/currency"
	"jww/internal/service/currency_exchange_rate"
	"jww/internal/service/object_storage"
	"jww/internal/service/transaction"
	"jww/internal/service/transaction_attachment"
	"jww/internal/service/transaction_category"
	"jww/internal/service/user"
	"jww/internal/service/workflow"
)

// buildDeps constructs every repository, service, and workflow exactly
// once, at process startup, and wires them into kernel.Deps. This is
// the one place that knows about every concrete adapter (gateway,
// database, security, storage) — resource subpackages under http_server/ never
// see any of this; they only ever call the already-built services/workflows.
func buildDeps(
	ctx context.Context,
	pool *pgxpool.Pool,
	authConfig security.Config,
	frankfurterConfig frankfurter.Config,
	storageConfig storage.Config,
	corsOrigins []string,
	maxBodyBytes int64,
) (*kernel.Deps, error) {
	uowManager := database.NewPostgresUnitOfWorkManager(pool)
	passwordHasher := security.NewArgon2PasswordHasher()
	tokenService := security.NewJwtTokenService(authConfig)
	rateSource := frankfurter.NewClient(frankfurterConfig)

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
		CurrencyExchangeRateWorkflow: workflow.NewCurrencyExchangeRateWorkflow(currencyService, exchangeRateService, rateSource, uowManager),
		TransactionCategoryService:   categoryService,
		TransactionService:           transactionService,
		TransactionAttachmentService: attachmentService,
		TransactionWorkflow:          workflow.NewTransactionWorkflow(currencyService, categoryService, userService, transactionService, attachmentService, uowManager),
		SummaryWorkflow:              workflow.NewSummaryWorkflow(currencyService, userService, transactionService, exchangeRateService, uowManager),
		CorsOrigins:                  corsOrigins,
		MaxBodyBytes:                 maxBodyBytes,
	}, nil
}

// reportOrphanedObject is the boundary's one high-severity log for an object
// that survived a failed metadata write. It logs identifiers only — never the
// original filename, storage key, credentials, or body.
func reportOrphanedObject(ctx domain.AppContext, attachmentID uuid.UUID, cleanupErr error) {
	log.WithContext(ctx).Error(
		"attachment object orphaned after failed metadata write",
		slog.String("attachment_id", attachmentID.String()),
		slog.String("error", cleanupErr.Error()),
	)
}
