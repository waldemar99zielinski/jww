package transaction

import (
	"net/http"
	"time"

	"jww/internal/adapter/schema"
	"jww/internal/infrastructure/http_server/kernel"
	httputils "jww/internal/infrastructure/http_server/utils"
)

// attachAttachmentHandler links an existing unattached draft to the path
// transaction. Repeating against the same transaction is a 204 no-op; a draft
// already linked elsewhere is a 409.
func attachAttachmentHandler(deps *kernel.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		appCtx := kernel.RequestContext(r)

		transactionID, err := httputils.ParseInt64PathParam(r, "transactionID")
		if err != nil {
			kernel.WriteBadRequest(w, err)
			return
		}
		attachmentID, err := httputils.ParseUUIDPathParam(r, "attachmentID")
		if err != nil {
			kernel.WriteBadRequest(w, err)
			return
		}
		if err := deps.TransactionWorkflow.Attach(appCtx, transactionID, attachmentID); err != nil {
			kernel.WriteError(w, r, err)
			return
		}
		kernel.WriteNoContent(w)
	}
}

func uploadAttachmentHandler(deps *kernel.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		appCtx := kernel.RequestContext(r)

		transactionID, err := httputils.ParseInt64PathParam(r, "transactionID")
		if err != nil {
			kernel.WriteBadRequest(w, err)
			return
		}
		file, err := kernel.ReadSingleMultipartFile(r)
		if err != nil {
			kernel.WriteBadRequest(w, err)
			return
		}
		_ = http.NewResponseController(w).SetReadDeadline(time.Time{})

		attachment, err := deps.TransactionWorkflow.UploadAttachment(appCtx, transactionID, file.Filename, file.MediaType, file)
		if err != nil {
			kernel.WriteError(w, r, err)
			return
		}
		if err := file.Finish(); err != nil {
			_ = deps.TransactionAttachmentService.DeleteForTransaction(appCtx, transactionID, attachment.ID)
			kernel.WriteBadRequest(w, err)
			return
		}
		kernel.WriteCreated(w, schema.TransactionAttachmentToResponse(*attachment))
	}
}

func listAttachmentsHandler(deps *kernel.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		appCtx := kernel.RequestContext(r)

		transactionID, err := httputils.ParseInt64PathParam(r, "transactionID")
		if err != nil {
			kernel.WriteBadRequest(w, err)
			return
		}
		t, err := deps.TransactionWorkflow.Get(appCtx, transactionID)
		if err != nil {
			kernel.WriteError(w, r, err)
			return
		}
		kernel.WriteOK(w, schema.TransactionAttachmentsToResponse(t.Attachments))
	}
}

func downloadAttachmentHandler(deps *kernel.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		appCtx := kernel.RequestContext(r)

		transactionID, err := httputils.ParseInt64PathParam(r, "transactionID")
		if err != nil {
			kernel.WriteBadRequest(w, err)
			return
		}
		attachmentID, err := httputils.ParseUUIDPathParam(r, "attachmentID")
		if err != nil {
			kernel.WriteBadRequest(w, err)
			return
		}
		attachment, stream, err := deps.TransactionAttachmentService.OpenForTransaction(appCtx, transactionID, attachmentID)
		if err != nil {
			kernel.WriteError(w, r, err)
			return
		}
		_ = http.NewResponseController(w).SetWriteDeadline(time.Time{})
		kernel.WriteDownload(w, attachment.OriginalFilename, attachment.MediaType, attachment.ByteSize, stream.Body)
	}
}

func deleteAttachmentHandler(deps *kernel.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		appCtx := kernel.RequestContext(r)

		transactionID, err := httputils.ParseInt64PathParam(r, "transactionID")
		if err != nil {
			kernel.WriteBadRequest(w, err)
			return
		}
		attachmentID, err := httputils.ParseUUIDPathParam(r, "attachmentID")
		if err != nil {
			kernel.WriteBadRequest(w, err)
			return
		}
		if err := deps.TransactionAttachmentService.DeleteForTransaction(appCtx, transactionID, attachmentID); err != nil {
			kernel.WriteError(w, r, err)
			return
		}
		kernel.WriteNoContent(w)
	}
}
