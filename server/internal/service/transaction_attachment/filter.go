package transaction_attachment

import "github.com/google/uuid"

// Filter constrains an attachment lookup. ID is always required. DraftOnly
// selects only unattached rows; TransactionID selects only rows belonging to
// that transaction. Set constraints are combined.
type Filter struct {
	ID            uuid.UUID
	TransactionID *int64
	DraftOnly     bool
}
