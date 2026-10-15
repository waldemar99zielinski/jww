package domain

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// User is a registered account. PasswordHash is the Argon2id hash, never a
// plaintext password.
type User struct {
	ID           uuid.UUID
	Username     string
	PasswordHash string
	DisplayName  string
	CreatedAt    time.Time
}

func (u User) Validate() error {
	if strings.TrimSpace(u.Username) == "" {
		return fmt.Errorf("user username must not be blank")
	}
	if strings.TrimSpace(u.PasswordHash) == "" {
		return fmt.Errorf("user password_hash must not be blank")
	}
	if strings.TrimSpace(u.DisplayName) == "" {
		return fmt.Errorf("user display_name must not be blank")
	}
	return nil
}
