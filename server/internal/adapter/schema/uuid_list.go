package schema

import (
	"fmt"

	"github.com/google/uuid"
)

// noDuplicateUUIDs rejects a list carrying the same id twice; a repeated claim
// is a client error, and silently de-duplicating it would hide the mistake.
func noDuplicateUUIDs(field string, ids []uuid.UUID) error {
	seen := make(map[uuid.UUID]struct{}, len(ids))
	for _, id := range ids {
		if _, ok := seen[id]; ok {
			return fmt.Errorf("%s must not contain duplicate ids", field)
		}
		seen[id] = struct{}{}
	}
	return nil
}

// noNilUUIDs rejects an all-zero UUID in a list of ids, which can only be a
// client mistake and would never match a row.
func noNilUUIDs(field string, ids []uuid.UUID) error {
	for _, id := range ids {
		if id == uuid.Nil {
			return fmt.Errorf("%s must not contain an empty id", field)
		}
	}
	return nil
}
