package transaction_attachment

import (
	"net/http"
	"time"

	"jww/internal/adapter/schema"
	"jww/internal/domain"
	"jww/internal/infrastructure/http_server/kernel"
	httputils "jww/internal/infrastructure/http_server/utils"
	"jww/internal/service/transaction_attachment"
)

func issueDraftGrantHandler(deps *kernel.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		appCtx := kernel.RequestContext(r)
		attachmentID, err := httputils.ParseUUIDPathParam(r, "attachmentID")
		if err != nil {
			kernel.WriteBadRequest(w, err)
			return
		}
		req, err := kernel.DecodeJSON[schema.TransactionAttachmentGrantRequest](r)
		if err != nil {
			kernel.WriteBadRequest(w, err)
			return
		}
		purpose, err := transaction_attachment.ParseGrantPurpose(req.Purpose)
		if err != nil {
			kernel.WriteBadRequest(w, err)
			return
		}
		grant, err := deps.TransactionAttachmentService.IssueDraftGrant(appCtx, attachmentID, purpose)
		if err != nil {
			kernel.WriteError(w, r, err)
			return
		}
		w.Header().Set("Cache-Control", "private, no-store")
		kernel.WriteOK(w, schema.TransactionAttachmentGrantResponse{Grant: grant})
	}
}

func grantedContentHandler(deps *kernel.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		appCtx := kernel.RequestContext(r)
		token := r.URL.Query().Get("grant")
		if token == "" {
			kernel.WriteError(w, r, domain.NewAuthenticationError())
			return
		}
		granted, err := deps.TransactionAttachmentService.ResolveGrant(appCtx, token)
		if err != nil {
			kernel.WriteError(w, r, err)
			return
		}
		attachment := granted.Attachment()
		byteRange, err := kernel.ParseSingleByteRange(r.Header.Get("Range"), attachment.ByteSize)
		if err != nil {
			kernel.WriteRangeNotSatisfiable(w, attachment.ByteSize)
			return
		}
		stream, err := deps.TransactionAttachmentService.OpenGranted(appCtx, *granted, byteRange)
		if err != nil {
			kernel.WriteError(w, r, err)
			return
		}
		preview := granted.Purpose() == transaction_attachment.GrantPurposePreview
		_ = http.NewResponseController(w).SetWriteDeadline(time.Time{})
		kernel.WriteGrantedContent(w, attachment.OriginalFilename, attachment.MediaType, attachment.ByteSize, preview, byteRange, stream.Body)
	}
}
