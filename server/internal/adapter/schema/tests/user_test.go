package schema_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"jww/internal/adapter/schema"
	"jww/internal/domain"
)

func TestUserCreateRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     schema.UserCreateRequest
		wantErr bool
	}{
		{"valid", schema.UserCreateRequest{Username: "alice", Password: "longenough", DisplayName: "Alice"}, false},
		{"blank username", schema.UserCreateRequest{Username: "", Password: "longenough", DisplayName: "Alice"}, true},
		{"password too short", schema.UserCreateRequest{Username: "alice", Password: "short", DisplayName: "Alice"}, true},
		{"blank display name", schema.UserCreateRequest{Username: "alice", Password: "longenough", DisplayName: ""}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUserUpdateRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     schema.UserUpdateRequest
		wantErr bool
	}{
		{"valid", schema.UserUpdateRequest{DisplayName: "Alice"}, false},
		{"blank display name", schema.UserUpdateRequest{DisplayName: ""}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUserToResponse(t *testing.T) {
	id := uuid.New()
	createdAt := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	u := domain.User{
		ID:           id,
		Username:     "alice",
		PasswordHash: "$argon2id$...should-never-appear-in-response",
		DisplayName:  "Alice",
		CreatedAt:    createdAt,
	}

	resp := schema.UserToResponse(u)

	assert.Equal(t, id, resp.ID)
	assert.Equal(t, "alice", resp.Username)
	assert.Equal(t, "Alice", resp.DisplayName)
	assert.True(t, createdAt.Equal(resp.CreatedAt))
}

func TestUserResponse_JSONNeverContainsPasswordHash(t *testing.T) {
	u := domain.User{ID: uuid.New(), Username: "alice", PasswordHash: "super-secret-hash", DisplayName: "Alice"}

	data, err := json.Marshal(schema.UserToResponse(u))
	require.NoError(t, err)

	var fields map[string]any
	require.NoError(t, json.Unmarshal(data, &fields))
	_, hasPasswordHash := fields["password_hash"]
	assert.False(t, hasPasswordHash, "UserResponse JSON must never include password_hash")
	assert.NotContains(t, string(data), "super-secret-hash")
}
