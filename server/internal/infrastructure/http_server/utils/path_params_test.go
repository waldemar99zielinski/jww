package utils

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func requestWithPathParam(name, value string) *http.Request {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(name, value)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

func TestParseInt64PathParam(t *testing.T) {
	t.Run("parses a named parameter", func(t *testing.T) {
		got, err := ParseInt64PathParam(requestWithPathParam("transactionID", "42"), "transactionID")

		require.NoError(t, err)
		assert.Equal(t, int64(42), got)
	})

	for _, value := range []string{"", "abc", "9223372036854775808"} {
		t.Run("rejects "+value, func(t *testing.T) {
			_, err := ParseInt64PathParam(requestWithPathParam("transactionID", value), "transactionID")
			require.Error(t, err)
		})
	}
}

func TestParseUUIDPathParam(t *testing.T) {
	want := uuid.New()
	got, err := ParseUUIDPathParam(requestWithPathParam("userID", want.String()), "userID")
	require.NoError(t, err)
	assert.Equal(t, want, got)

	for _, value := range []string{"", "not-a-uuid"} {
		t.Run("rejects "+value, func(t *testing.T) {
			_, err := ParseUUIDPathParam(requestWithPathParam("userID", value), "userID")
			require.Error(t, err)
		})
	}
}
