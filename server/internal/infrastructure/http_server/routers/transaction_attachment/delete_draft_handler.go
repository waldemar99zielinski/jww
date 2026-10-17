package transaction_attachment

import (
	"net/http"

	"jww/internal/infrastructure/http_server/kernel"
	httputils "jww/internal/infrastructure/http_server/utils"
)

func deleteDraftHandler(deps *kernel.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		appCtx := kernel.RequestContext(r)

		attachmentID, err := httputils.ParseUUIDPathParam(r, "attachmentID")
		if err != nil {
			kernel.WriteBadRequest(w, err)
			return
		}
		if err := deps.TransactionAttachmentService.DeleteDraft(appCtx, attachmentID); err != nil {
			kernel.WriteError(w, err)
			return
		}
		kernel.WriteNoContent(w)
	}
}
