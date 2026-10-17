package domain

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func validAttachment() TransactionAttachment {
	return TransactionAttachment{
		ID:               uuid.New(),
		OriginalFilename: "receipt.pdf",
		MediaType:        "application/pdf",
		ByteSize:         1024,
		StorageKey:       "transaction-attachments/" + uuid.NewString(),
		CreatedAt:        time.Now().UTC(),
	}
}

func TestTransactionAttachment_Validate(t *testing.T) {
	transactionID := int64(7)

	tests := []struct {
		name    string
		mutate  func(*TransactionAttachment)
		wantErr bool
	}{
		{"valid draft without a transaction", func(a *TransactionAttachment) {}, false},
		{"valid attached row", func(a *TransactionAttachment) { a.TransactionID = &transactionID }, false},
		{"zero byte size is allowed", func(a *TransactionAttachment) { a.ByteSize = 0 }, false},
		{"empty id", func(a *TransactionAttachment) { a.ID = uuid.Nil }, true},
		{"non-positive transaction id", func(a *TransactionAttachment) { zero := int64(0); a.TransactionID = &zero }, true},
		{"blank filename", func(a *TransactionAttachment) { a.OriginalFilename = "   " }, true},
		{"over-long filename", func(a *TransactionAttachment) {
			a.OriginalFilename = strings.Repeat("a", maxAttachmentFilenameLength+1)
		}, true},
		{"blank media type", func(a *TransactionAttachment) { a.MediaType = "" }, true},
		{"negative byte size", func(a *TransactionAttachment) { a.ByteSize = -1 }, true},
		{"blank storage key", func(a *TransactionAttachment) { a.StorageKey = "" }, true},
		{"zero created at", func(a *TransactionAttachment) { a.CreatedAt = time.Time{} }, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := validAttachment()
			tt.mutate(&a)
			if tt.wantErr {
				assert.Error(t, a.Validate())
			} else {
				assert.NoError(t, a.Validate())
			}
		})
	}
}
