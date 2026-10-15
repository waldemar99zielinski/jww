package main

import (
	"github.com/jackc/pgx/v5/pgxpool"

	"jww/internal/adapter/frankfurter"
	"jww/internal/adapter/gateway"
	"jww/internal/adapter/security"
	"jww/internal/infrastructure/database"
	"jww/internal/infrastructure/http_server/kernel"
	"jww/internal/service/auth"
	"jww/internal/service/currency"
	"jww/internal/service/currency_exchange_rate"
	"jww/internal/service/transaction"
	"jww/internal/service/transaction_category"
	"jww/internal/service/user"
	"jww/internal/service/workflow"
)

// buildDeps constructs every repository, service, and workflow exactly
// once, at process startup, and wires them into kernel.Deps. This is
// the one place that knows about every concrete adapter (gateway,
// database, security) — resource subpackages under http_server/ never see
// any of this; they only ever call the already-built services/workflows.
func buildDeps(pool *pgxpool.Pool, authConfig security.Config, frankfurterConfig frankfurter.Config, corsOrigins []string, maxBodyBytes int64) *kernel.Deps {
	uowManager := database.NewPostgresUnitOfWorkManager(pool)
	passwordHasher := security.NewArgon2PasswordHasher()
	tokenService := security.NewJwtTokenService(authConfig)
	rateSource := frankfurter.NewClient(frankfurterConfig)

	userRepo := gateway.NewPostgresUserRepository(pool)
	currencyRepo := gateway.NewPostgresCurrencyRepository(pool)
	exchangeRateRepo := gateway.NewPostgresCurrencyExchangeRateRepository(pool)
	categoryRepo := gateway.NewPostgresTransactionCategoryRepository(pool)
	transactionRepo := gateway.NewPostgresTransactionRepository(pool)

	authService := auth.NewService(userRepo, passwordHasher, tokenService)
	userService := user.NewService(userRepo, passwordHasher, uowManager)
	currencyService := currency.NewService(currencyRepo, uowManager)
	exchangeRateService := currency_exchange_rate.NewService(exchangeRateRepo)
	categoryService := transaction_category.NewService(categoryRepo, uowManager)
	transactionService := transaction.NewService(transactionRepo, uowManager)

	return &kernel.Deps{
		AuthService:                  authService,
		UserService:                  userService,
		SignupWorkflow:               workflow.NewSignupWorkflow(userService),
		CurrencyService:              currencyService,
		CurrencyExchangeRateWorkflow: workflow.NewCurrencyExchangeRateWorkflow(currencyService, exchangeRateService, rateSource, uowManager),
		TransactionCategoryService:   categoryService,
		TransactionService:           transactionService,
		TransactionWorkflow:          workflow.NewTransactionWorkflow(currencyService, categoryService, userService, transactionService, uowManager),
		SummaryWorkflow:              workflow.NewSummaryWorkflow(currencyService, userService, transactionService, exchangeRateService, uowManager),
		CorsOrigins:                  corsOrigins,
		MaxBodyBytes:                 maxBodyBytes,
	}
}
