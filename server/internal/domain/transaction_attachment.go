package domain

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

// Bounds mirrored by the migration's CHECK constraints. They bound the
// metadata text we accept; they do not bound the file's byte size, which is
// deliberately unlimited (see the design's scope note).
const (
	maxAttachmentFilenameLength  = 255
	maxAttachmentMediaTypeLength = 255
	maxAttachmentStorageKeyLen   = 512
)

// TransactionAttachment is one file's metadata. The bytes live in private
// object storage under StorageKey; this record is the only authoritative
// mapping from an attachment (and, once linked, its transaction) to that
// object.
//
// TransactionID is nil while the attachment is an unattached draft and is set
// exactly once, in a single NULL -> transaction transition, when the
// attachment is linked. There is no detach or reassign.
type TransactionAttachment struct {
	ID               uuid.UUID
	TransactionID    *int64
	OriginalFilename string
	MediaType        string
	ByteSize         int64
	// StorageKey and StorageVersion are internal only: they are never
	// serialized to a client and never logged.
	StorageKey     string
	StorageVersion string
	ETag           string
	CreatedAt      time.Time
}

// Validate enforces the attachment's own invariants. It deliberately does not
// allowlist media types, impose a byte maximum, or model an ordinal/count.
func (a TransactionAttachment) Validate() error {
	if a.ID == uuid.Nil {
		return fmt.Errorf("id must not be empty")
	}
	if a.TransactionID != nil && *a.TransactionID <= 0 {
		return fmt.Errorf("transaction id must be positive")
	}
	if strings.TrimSpace(a.OriginalFilename) == "" {
		return fmt.Errorf("original filename must not be blank")
	}
	if utf8.RuneCountInString(a.OriginalFilename) > maxAttachmentFilenameLength {
		return fmt.Errorf("original filename must be at most %d characters", maxAttachmentFilenameLength)
	}
	if strings.TrimSpace(a.MediaType) == "" {
		return fmt.Errorf("media type must not be blank")
	}
	if len(a.MediaType) > maxAttachmentMediaTypeLength {
		return fmt.Errorf("media type must be at most %d characters", maxAttachmentMediaTypeLength)
	}
	if a.ByteSize < 0 {
		return fmt.Errorf("byte size must not be negative")
	}
	if strings.TrimSpace(a.StorageKey) == "" {
		return fmt.Errorf("storage key must not be blank")
	}
	if len(a.StorageKey) > maxAttachmentStorageKeyLen {
		return fmt.Errorf("storage key must be at most %d characters", maxAttachmentStorageKeyLen)
	}
	if a.CreatedAt.IsZero() {
		return fmt.Errorf("created at must not be zero")
	}
	return nil
}
