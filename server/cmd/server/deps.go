package main

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"jww/internal/adapter/frankfurter"
	"jww/internal/adapter/gateway"
	"jww/internal/adapter/imageprep"
	"jww/internal/adapter/openai"
	"jww/internal/adapter/security"
	jwtadapter "jww/internal/adapter/security/jwt"
	webauthnadapter "jww/internal/adapter/security/webauthn"
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
	transaction_attachment_extraction "jww/internal/service/transaction_attachment_extraction"
	"jww/internal/service/transaction_category"
	"jww/internal/service/transaction_sub_transaction"
	"jww/internal/service/user"
	"jww/internal/service/webauthn_credential"
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
	jwtConfig jwtadapter.Config,
	webauthnConfig webauthnadapter.Config,
	frankfurterConfig frankfurter.Config,
	storageConfig storage.Config,
	openaiConfig openai.Config,
	extractionConfig transaction_attachment_extraction.Config,
	corsOrigins []string,
	maxBodyBytes int64,
) (*kernel.Deps, error) {
	uowManager := database.NewPostgresUnitOfWorkManager(pool)
	passwordHasher := security.NewArgon2PasswordHasher()
	tokenService := jwtadapter.NewAccessTokenService(jwtConfig)
	attachmentGrantService := jwtadapter.NewAttachmentGrantService(jwtConfig, 5*time.Minute)
	webauthnChallengeService := jwtadapter.NewWebAuthnChallengeService(jwtConfig, 5*time.Minute)
	rateSource := frankfurter.NewClient(frankfurterConfig)

	userRepo := gateway.NewPostgresUserRepository(pool)
	currencyRepo := gateway.NewPostgresCurrencyRepository(pool)
	exchangeRateRepo := gateway.NewPostgresCurrencyExchangeRateRepository(pool)
	categoryRepo := gateway.NewPostgresTransactionCategoryRepository(pool)
	transactionRepo := gateway.NewPostgresTransactionRepository(pool)
	attachmentRepo := gateway.NewPostgresTransactionAttachmentRepository(pool)
	subTransactionRepo := gateway.NewPostgresTransactionSubTransactionRepository(pool)
	webauthnCredentialRepo := gateway.NewPostgresWebAuthnCredentialRepository(pool)

	objectStore, err := storage.NewS3Store(ctx, storageConfig)
	if err != nil {
		return nil, err
	}
	webauthnAdapter, err := webauthnadapter.New(webauthnConfig)
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
	attachmentService := transaction_attachment.NewService(attachmentRepo, attachmentStore, reportOrphanedObject, attachmentGrantService).
		WithContentNormalizer(imageprep.New())
	subTransactionService := transaction_sub_transaction.NewService(subTransactionRepo, uowManager)
	extractionService := transaction_attachment_extraction.NewService(
		openai.NewClient(openaiConfig),
		draftAttachmentSource{attachments: attachmentService},
		extractionCategorySource{categories: categoryService},
		extractionConfig,
	)
	webauthnCredentialService := webauthn_credential.NewService(webauthnCredentialRepo, uowManager)

	return &kernel.Deps{
		AuthService:                            authService,
		UserService:                            userService,
		SignupWorkflow:                         workflow.NewSignupWorkflow(userService),
		CurrencyService:                        currencyService,
		CurrencyExchangeRateWorkflow:           workflow.NewCurrencyExchangeRateWorkflow(currencyService, exchangeRateService, rateSource, uowManager),
		TransactionCategoryService:             categoryService,
		TransactionService:                     transactionService,
		TransactionAttachmentService:           attachmentService,
		TransactionAttachmentExtractionService: extractionService,
		TransactionWorkflow:                    workflow.NewTransactionWorkflow(currencyService, categoryService, userService, transactionService, attachmentService, subTransactionService, uowManager),
		SummaryWorkflow:                        workflow.NewSummaryWorkflow(currencyService, userService, transactionService, exchangeRateService, uowManager),
		WebAuthnWorkflow:                       workflow.NewWebAuthnWorkflow(userService, webauthnCredentialService, webauthnAdapter, webauthnChallengeService, tokenService),
		CorsOrigins:                            corsOrigins,
		MaxBodyBytes:                           maxBodyBytes,
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

// draftAttachmentSource adapts the attachment service to
// transaction_attachment_extraction.AttachmentSource. The port names its
// methods DraftAttachment and ContentGrant, which don't match the service's
// GetDraft / IssueDraftGrant signatures (the latter takes a purpose), so this
// small cmd-local adapter bridges them.
type draftAttachmentSource struct {
	attachments *transaction_attachment.Service
}

func (s draftAttachmentSource) DraftAttachment(ctx domain.AppContext, id uuid.UUID) (*domain.TransactionAttachment, error) {
	return s.attachments.GetDraft(ctx, id)
}

func (s draftAttachmentSource) ContentGrant(ctx domain.AppContext, id uuid.UUID) (string, error) {
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
