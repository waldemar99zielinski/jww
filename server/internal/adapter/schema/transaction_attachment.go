package schema

import (
	"time"

	"github.com/google/uuid"

	"jww/internal/domain"
)

// TransactionAttachmentResponse is the public view of an attachment. It
// deliberately omits storage_key, storage_version, ETag, bucket/endpoint, any
// ordinal or count, and any download URL. transaction_id is null for an
// unattached draft and populated once linked.
type TransactionAttachmentResponse struct {
	ID               uuid.UUID `json:"id"`
	TransactionID    *int64    `json:"transaction_id"`
	OriginalFilename string    `json:"original_filename"`
	MediaType        string    `json:"media_type"`
	ByteSize         int64     `json:"byte_size"`
	CreatedAt        time.Time `json:"created_at"`
}

func TransactionAttachmentToResponse(a domain.TransactionAttachment) TransactionAttachmentResponse {
	return TransactionAttachmentResponse{
		ID:               a.ID,
		TransactionID:    a.TransactionID,
		OriginalFilename: a.OriginalFilename,
		MediaType:        a.MediaType,
		ByteSize:         a.ByteSize,
		CreatedAt:        a.CreatedAt,
	}
}

// TransactionAttachmentsToResponse converts a slice, always returning a
// non-nil (possibly empty) slice so JSON encodes [] rather than null.
func TransactionAttachmentsToResponse(attachments []domain.TransactionAttachment) []TransactionAttachmentResponse {
	out := make([]TransactionAttachmentResponse, len(attachments))
	for i, a := range attachments {
		out[i] = TransactionAttachmentToResponse(a)
	}
	return out
}
