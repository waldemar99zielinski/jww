package domain

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// TransactionWithPreviewGrants is a transaction paired with short-lived inline
// preview grants for its safely previewable attachments, keyed by attachment
// ID. It is the detail-read result; create and list paths do not mint grants.
type TransactionWithPreviewGrants struct {
	Transaction   Transaction
	PreviewGrants map[uuid.UUID]string
}

// Transaction is a single income/expense record. ID is a plain int64 (a
// Postgres bigint identity column), unlike the UUID-keyed entities.
//
// Attachments holds this transaction's linked files, ordered by creation time
// then id. SubTransactions holds this transaction's line items, ordered by
// creation time then id. Both are populated for reads and are not part of
// Validate; a bare transaction (e.g. one just built for a write) leaves them
// nil.
type Transaction struct {
	ID              int64
	Type            TransactionType
	Amount          Money
	Currency        Currency
	Category        TransactionCategory
	User            User
	Source          TransactionSource
	Name            string
	Description     string
	CreatedAt       time.Time
	Attachments     []TransactionAttachment
	SubTransactions []TransactionSubTransaction
}

func (t Transaction) Validate() error {
	if t.Amount <= 0 {
		return fmt.Errorf("amount must be positive")
	}
	if string(t.Category.Type) != string(t.Type) {
		return fmt.Errorf("category type must match transaction type")
	}
	if strings.TrimSpace(t.Name) == "" {
		return fmt.Errorf("name must not be blank")
	}
	return nil
}
