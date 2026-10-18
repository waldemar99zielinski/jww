package transaction_attachment

import (
	"net/http"
	"time"

	"jww/internal/infrastructure/http_server/kernel"
	httputils "jww/internal/infrastructure/http_server/utils"
)

func downloadDraftHandler(deps *kernel.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		appCtx := kernel.RequestContext(r)

		attachmentID, err := httputils.ParseUUIDPathParam(r, "attachmentID")
		if err != nil {
			kernel.WriteBadRequest(w, err)
			return
		}
		attachment, stream, err := deps.TransactionAttachmentService.OpenDraft(appCtx, attachmentID)
		if err != nil {
			kernel.WriteError(w, r, err)
			return
		}
		// The response can be arbitrarily large; clear the inherited write
		// deadline before streaming (still after auth/validation).
		_ = http.NewResponseController(w).SetWriteDeadline(time.Time{})
		kernel.WriteDownload(w, attachment.OriginalFilename, attachment.MediaType, attachment.ByteSize, stream.Body)
	}
}
