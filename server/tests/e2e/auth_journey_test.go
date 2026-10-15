//go:build e2e

package e2e_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"jww/internal/adapter/schema"
	"jww/tests/utils"
)

// TestAuthJourney covers login, the authenticated /me lookup, and the
// unauthenticated-request rejection — the one journey every other journey's
// authenticated requests depend on working correctly.
func TestAuthJourney(t *testing.T) {
	token := apiClient.Login(t, seededUser.Username, seededPassword)

	var me utils.Envelope[schema.UserResponse]
	status := apiClient.Do(t, http.MethodGet, "/api/v1/auth/me", token, nil, &me)
	require.Equal(t, http.StatusOK, status)
	assert.Equal(t, seededUser.Username, me.Data.Username)
	assert.Equal(t, seededUser.ID, me.Data.ID)

	t.Run("no token", func(t *testing.T) {
		t.Parallel()
		req, err := http.NewRequest(http.MethodGet, httpSrv.URL+"/api/v1/auth/me", nil)
		require.NoError(t, err)
		resp, err := apiClient.HTTP.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		var errResp utils.Envelope[struct{}]
		utils.DecodeBody(t, resp, &errResp)

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		assert.Equal(t, "Bearer", resp.Header.Get("WWW-Authenticate"))
		require.NotNil(t, errResp.Error)
		assert.Equal(t, "unauthorized", errResp.Error.Code)
	})

	t.Run("garbage token", func(t *testing.T) {
		t.Parallel()
		var errResp utils.Envelope[struct{}]
		status := apiClient.Do(t, http.MethodGet, "/api/v1/auth/me", "not-a-real-token", nil, &errResp)
		assert.Equal(t, http.StatusUnauthorized, status)
		require.NotNil(t, errResp.Error)
		assert.Equal(t, "unauthorized", errResp.Error.Code)
	})
}
