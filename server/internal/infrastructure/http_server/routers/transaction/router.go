package transaction

import (
	"github.com/go-chi/chi/v5"

	"jww/internal/infrastructure/http_server/kernel"
)

// Mount registers transaction CRUD plus the transaction-scoped attachment
// routes. The nested attachment routes require the attachment to belong to the
// path transaction.
func Mount(_, protected chi.Router, deps *kernel.Deps) {
	protected.Post("/v1/transactions", createTransactionHandler(deps))
	protected.Get("/v1/transactions", listTransactionsHandler(deps))
	protected.Get("/v1/transactions/{transactionID}", getTransactionHandler(deps))
	protected.Patch("/v1/transactions/{transactionID}", updateTransactionHandler(deps))
	protected.Delete("/v1/transactions/{transactionID}", deleteTransactionHandler(deps))

	protected.Post("/v1/transactions/{transactionID}/attachments", uploadAttachmentHandler(deps))
	protected.Put("/v1/transactions/{transactionID}/attachments/{attachmentID}", attachAttachmentHandler(deps))
	protected.Get("/v1/transactions/{transactionID}/attachments", listAttachmentsHandler(deps))
	protected.Get("/v1/transactions/{transactionID}/attachments/{attachmentID}/download", downloadAttachmentHandler(deps))
	protected.Delete("/v1/transactions/{transactionID}/attachments/{attachmentID}", deleteAttachmentHandler(deps))
}
