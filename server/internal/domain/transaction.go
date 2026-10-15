package domain

import (
	"fmt"
	"strings"
	"time"
)

// Transaction is a single income/expense record. ID is a plain int64 (a
// Postgres bigint identity column), unlike the UUID-keyed entities.
type Transaction struct {
	ID          int64
	Type        TransactionType
	Amount      Money
	Currency    Currency
	Category    TransactionCategory
	User        User
	Source      TransactionSource
	Name        string
	Description string
	CreatedAt   time.Time
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
