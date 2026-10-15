package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func validUser() User {
	return User{
		Username:     "wozi",
		PasswordHash: "$argon2id$v=19$m=65536,t=3,p=2$...",
		DisplayName:  "Wozi",
	}
}

func TestUser_Validate(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(u User) User
		wantErr string
	}{
		{"valid", func(u User) User { return u }, ""},
		{"blank username", func(u User) User { u.Username = "  "; return u }, "username"},
		{"blank password hash", func(u User) User { u.PasswordHash = ""; return u }, "password_hash"},
		{"blank display name", func(u User) User { u.DisplayName = "   "; return u }, "display_name"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.mutate(validUser()).Validate()
			if tt.wantErr == "" {
				assert.NoError(t, err)
				return
			}
			assert.ErrorContains(t, err, tt.wantErr)
		})
	}
}
