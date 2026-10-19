package transaction_attachment

import (
	"github.com/go-chi/chi/v5"

	"jww/internal/infrastructure/http_server/kernel"
	"jww/utils/logger"
)

// log is this router package's own logger. It exists for extraction, whose
// outcome summary (attachment id, duration, failure kind) is something the
// kernel's RequestLogger middleware cannot see. It is not used for the plain
// draft CRUD routes — the middleware already covers those.
var log = logger.New().WithComponent("http_server.transaction_attachment")

// Mount registers the standalone draft routes (protected only). A draft is an
// attachment with transaction_id IS NULL; every ID operation here 404s once the
// attachment is linked.
func Mount(public, protected chi.Router, deps *kernel.Deps) {
	public.Get("/v1/attachment-content", grantedContentHandler(deps))
	protected.Post("/v1/transaction-attachments", uploadDraftHandler(deps))
	protected.Get("/v1/transaction-attachments/{attachmentID}", getDraftHandler(deps))
	protected.Get("/v1/transaction-attachments/{attachmentID}/download", downloadDraftHandler(deps))
	protected.Post("/v1/transaction-attachments/{attachmentID}/grants", issueDraftGrantHandler(deps))
	protected.Post("/v1/transaction-attachments/{attachmentID}/extraction", extractionHandler(deps))
	protected.Post("/v1/transaction-attachments/extraction", batchExtractionHandler(deps))
	protected.Delete("/v1/transaction-attachments/{attachmentID}", deleteDraftHandler(deps))
}
