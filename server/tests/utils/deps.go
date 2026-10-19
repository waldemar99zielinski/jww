package utils

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"jww/internal/adapter/gateway"
	"jww/internal/adapter/imageprep"
	"jww/internal/adapter/security"
	jwtadapter "jww/internal/adapter/security/jwt"
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
	transaction_attachment_extraction "jww/internal/service/transaction_attachment_extraction"
	"jww/internal/service/transaction_category"
	"jww/internal/service/transaction_sub_transaction"
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

// stubExtractionReader satisfies transaction_attachment_extraction.Extractor for
// the in-process journey tests. It never makes a network call: it returns a
// deterministic, canned result so an extraction request is offline and stable.
type stubExtractionReader struct{}

func (stubExtractionReader) ExtractTransactionPrefill(domain.AppContext, transaction_attachment_extraction.TransactionPrefillInput) (*transaction_attachment_extraction.TransactionPrefillResult, error) {
	return cannedExtractionResult(), nil
}

// cannedExtractionResult is the deterministic, network-free payload every
// extraction request returns in the e2e suite: representative scalar
// suggestions (an expense type, a decimal-string amount, a name, a currency
// hint) plus two line items with names, free-text labels, and decimal-string
// amounts. Two items are the minimum an e2e journey needs to exercise the
// derived-Σ(items) amount rule on create and on later item CRUD.
func cannedExtractionResult() *transaction_attachment_extraction.TransactionPrefillResult {
	typeField := transaction_attachment_extraction.PrefillField[domain.TransactionType]{
		Value:  domain.TransactionTypeExpense,
		Source: "canned stub",
	}
	return &transaction_attachment_extraction.TransactionPrefillResult{
		Type: &typeField,
		Amount: &transaction_attachment_extraction.PrefillField[domain.Money]{
			Value:  domain.Money(1799),
			Source: "canned stub",
		},
		CurrencyCode: &transaction_attachment_extraction.PrefillField[string]{
			Value:  "ZZD",
			Source: "canned stub",
		},
		Name: &transaction_attachment_extraction.PrefillField[string]{
			Value:  "Zzyzx Grocery",
			Source: "canned stub",
		},
		SubTransactions: []transaction_attachment_extraction.PrefillSubTransaction{
			{Name: "Zzyzx Coffee Beans", Label: "pantry", Amount: domain.Money(1234)},
			{Name: "Zzyzx Oat Milk", Label: "dairy-free", Amount: domain.Money(565)},
		},
	}
}

// extractionAttachmentSource adapts the attachment service to
// transaction_attachment_extraction.AttachmentSource for the journey tests,
// mirroring cmd/server's adapter: the port's DraftAttachment / ContentGrant
// method names don't match the service's GetDraft / IssueDraftGrant signatures.
type extractionAttachmentSource struct {
	attachments *transaction_attachment.Service
}

func (s extractionAttachmentSource) DraftAttachment(ctx domain.AppContext, id uuid.UUID) (*domain.TransactionAttachment, error) {
	return s.attachments.GetDraft(ctx, id)
}

func (s extractionAttachmentSource) ContentGrant(ctx domain.AppContext, id uuid.UUID) (string, error) {
	return s.attachments.IssueDraftGrant(ctx, id, transaction_attachment.GrantPurposePreview)
}

type extractionCategorySource struct {
	categories *transaction_category.Service
}

func (s extractionCategorySource) TransactionCategories(ctx domain.AppContext) (expense, income []string, err error) {
	expenseCategories, err := s.categories.ListByType(ctx, domain.TransactionCategoryTypeExpense, true)
	if err != nil {
		return nil, nil, err
	}
	incomeCategories, err := s.categories.ListByType(ctx, domain.TransactionCategoryTypeIncome, true)
	if err != nil {
		return nil, nil, err
	}
	for _, category := range expenseCategories {
		expense = append(expense, category.Name)
	}
	for _, category := range incomeCategories {
		income = append(income, category.Name)
	}
	return expense, income, nil
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
	jwtConfig := jwtadapter.Config{
		SecretKey:                jwtSecret,
		Algorithm:                "HS256",
		AccessTokenExpireMinutes: 60,
	}
	tokenService := jwtadapter.NewAccessTokenService(jwtConfig)

	userRepo := gateway.NewPostgresUserRepository(pool)
	currencyRepo := gateway.NewPostgresCurrencyRepository(pool)
	exchangeRateRepo := gateway.NewPostgresCurrencyExchangeRateRepository(pool)
	categoryRepo := gateway.NewPostgresTransactionCategoryRepository(pool)
	transactionRepo := gateway.NewPostgresTransactionRepository(pool)
	attachmentRepo := gateway.NewPostgresTransactionAttachmentRepository(pool)
	subTransactionRepo := gateway.NewPostgresTransactionSubTransactionRepository(pool)

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
	attachmentGrantService := jwtadapter.NewAttachmentGrantService(jwtConfig, 5*time.Minute)
	attachmentService := transaction_attachment.NewService(attachmentRepo, attachmentStore, reportOrphanedObject, attachmentGrantService).
		WithContentNormalizer(imageprep.New())
	subTransactionService := transaction_sub_transaction.NewService(subTransactionRepo, uowManager)
	extractionService := transaction_attachment_extraction.NewService(
		stubExtractionReader{},
		extractionAttachmentSource{attachments: attachmentService},
		extractionCategorySource{categories: categoryService},
		transaction_attachment_extraction.Config{PublicBaseURL: "http://127.0.0.1"},
	)

	return &kernel.Deps{
		AuthService:                            authService,
		UserService:                            userService,
		SignupWorkflow:                         workflow.NewSignupWorkflow(userService),
		CurrencyService:                        currencyService,
		CurrencyExchangeRateWorkflow:           workflow.NewCurrencyExchangeRateWorkflow(currencyService, exchangeRateService, stubRateSource{}, uowManager),
		TransactionCategoryService:             categoryService,
		TransactionService:                     transactionService,
		TransactionAttachmentService:           attachmentService,
		TransactionAttachmentExtractionService: extractionService,
		TransactionWorkflow:                    workflow.NewTransactionWorkflow(currencyService, categoryService, userService, transactionService, attachmentService, subTransactionService, uowManager),
		SummaryWorkflow:                        workflow.NewSummaryWorkflow(currencyService, userService, transactionService, exchangeRateService, uowManager),
		CorsOrigins:                            []string{"*"},
		MaxBodyBytes:                           1 << 20,
	}, nil
}

// reportOrphanedObject mirrors cmd/server's reporter. Tests don't capture logs;
// the reporter exists so the wiring matches production.
func reportOrphanedObject(_ domain.AppContext, _ uuid.UUID, _ error) {}
