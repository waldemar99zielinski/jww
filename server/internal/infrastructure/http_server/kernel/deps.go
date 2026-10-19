// Package kernel holds the HTTP-layer kernel every resource subpackage
// (auth, user, currency, currency_exchange_rate, transaction_category,
// transaction, summary) depends on: the Deps singleton and request/response
// helpers. It deliberately knows nothing about any specific resource —
// resource subpackages import this package, never the other way around, so
// there's no import cycle. The http_server package mounts the subpackages.
package kernel

import (
	"jww/internal/service/auth"
	"jww/internal/service/currency"
	"jww/internal/service/transaction"
	"jww/internal/service/transaction_attachment"
	transaction_attachment_extraction "jww/internal/service/transaction_attachment_extraction"
	"jww/internal/service/transaction_category"
	"jww/internal/service/user"
	"jww/internal/service/workflow"
	"jww/utils/logger"
)

// httpLog is the package-wide logger for the HTTP layer. It is built from the
// logger initialized once in cmd/server and every component derives its own
// child logger from it via With / WithComponent.
var httpLog = logger.New().WithComponent("http_server")

// Deps holds every process-lifetime singleton the HTTP layer needs — the
// services/workflows each handler calls, already fully wired (repository,
// unit of work, and any collaborators) at startup. No handler constructs
// anything per request anymore; all business logic, including transaction
// demarcation, lives in the service layer these were built from.
type Deps struct {
	AuthService                            *auth.Service
	UserService                            *user.Service
	SignupWorkflow                         *workflow.SignupWorkflow
	CurrencyService                        *currency.Service
	CurrencyExchangeRateWorkflow           *workflow.CurrencyExchangeRateWorkflow
	TransactionCategoryService             *transaction_category.Service
	TransactionService                     *transaction.Service
	TransactionAttachmentService           *transaction_attachment.Service
	TransactionAttachmentExtractionService *transaction_attachment_extraction.Service
	TransactionWorkflow                    *workflow.TransactionWorkflow
	SummaryWorkflow                        *workflow.SummaryWorkflow
	CorsOrigins                            []string
	// MaxBodyBytes caps every request body; <= 0 falls back to
	// defaultMaxBodyBytes (see LimitRequestBody in middleware.go).
	MaxBodyBytes int64
}
