package transaction_attachment

import (
	"net/http"

	"jww/internal/adapter/schema"
	"jww/internal/infrastructure/http_server/kernel"
	httputils "jww/internal/infrastructure/http_server/utils"
)

func getDraftHandler(deps *kernel.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		appCtx := kernel.RequestContext(r)

		attachmentID, err := httputils.ParseUUIDPathParam(r, "attachmentID")
		if err != nil {
			kernel.WriteBadRequest(w, err)
			return
		}
		attachment, err := deps.TransactionAttachmentService.GetDraft(appCtx, attachmentID)
		if err != nil {
			kernel.WriteError(w, err)
			return
		}
		kernel.WriteOK(w, schema.TransactionAttachmentToResponse(*attachment))
	}
}
