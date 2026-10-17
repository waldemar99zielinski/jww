package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMatchPath(t *testing.T) {
	t.Run("matches literals and extracts named parameters", func(t *testing.T) {
		got := MatchPath(
			"/api/v1/transactions/{transactionID}/attachments/{attachmentID}",
			"/api/v1/transactions/42/attachments/receipt",
		)

		assert.True(t, got.Matched)
		assert.Equal(t, map[string]string{
			"transactionID": "42",
			"attachmentID":  "receipt",
		}, got.Params)
	})

	t.Run("matches a route without parameters", func(t *testing.T) {
		got := MatchPath("/api/v1/transaction-attachments", "/api/v1/transaction-attachments")

		assert.True(t, got.Matched)
		assert.Empty(t, got.Params)
	})

	for _, path := range []string{
		"/api/v1/transactions//attachments",
		"/api/v1/transactions/42/attachments/7",
		"/api/v1/transactions/42/other",
		"/api/v1/transactions/42/attachments/",
	} {
		t.Run("rejects "+path, func(t *testing.T) {
			got := MatchPath("/api/v1/transactions/{transactionID}/attachments", path)

			assert.False(t, got.Matched)
			assert.Nil(t, got.Params)
		})
	}
}
