package schema_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"jww/internal/adapter/schema"
)

func TestLoginRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     schema.LoginRequest
		wantErr bool
	}{
		{"valid", schema.LoginRequest{Username: "alice", Password: "secret"}, false},
		{"blank username", schema.LoginRequest{Username: "", Password: "secret"}, true},
		{"blank password", schema.LoginRequest{Username: "alice", Password: ""}, true},
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

func TestNewTokenResponse(t *testing.T) {
	resp := schema.NewTokenResponse("the-token")
	assert.Equal(t, "the-token", resp.AccessToken)
	assert.Equal(t, "bearer", resp.TokenType)
}
