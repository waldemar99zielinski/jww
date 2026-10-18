package transaction_attachment

import (
	"net/http"
	"time"

	"jww/internal/adapter/schema"
	"jww/internal/infrastructure/http_server/kernel"
)

func uploadDraftHandler(deps *kernel.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		appCtx := kernel.RequestContext(r)

		file, err := kernel.ReadSingleMultipartFile(r)
		if err != nil {
			kernel.WriteBadRequest(w, err)
			return
		}
		// Uploads are deliberately not size-limited; clear the inherited read
		// deadline now that authentication and header validation have passed.
		_ = http.NewResponseController(w).SetReadDeadline(time.Time{})

		attachment, err := deps.TransactionAttachmentService.UploadDraft(appCtx, file.Filename, file.MediaType, file)
		if err != nil {
			kernel.WriteError(w, r, err)
			return
		}
		if err := file.Finish(); err != nil {
			// A trailing part makes the request invalid, but the bytes are
			// already stored; delete the just-created draft so nothing lingers.
			_ = deps.TransactionAttachmentService.DeleteDraft(appCtx, attachment.ID)
			kernel.WriteBadRequest(w, err)
			return
		}
		kernel.WriteCreated(w, schema.TransactionAttachmentToResponse(*attachment))
	}
}
