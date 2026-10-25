package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseParamsUsesDefaultsAndEnvironment(t *testing.T) {
	env := map[string]string{
		"DATABASE_URL":          "postgres://development",
		"BOOTSTRAP_ATTACHMENTS": "12",
	}

	p, err := parseParams(nil, func(key string) string { return env[key] })

	require.NoError(t, err)
	assert.Equal(t, "postgres://development", p.databaseURL)
	assert.Equal(t, 5_000, p.transactions)
	assert.Equal(t, 12, p.attachments)
}

func TestParseParamsFlagsOverrideEnvironment(t *testing.T) {
	env := map[string]string{
		"DATABASE_URL":           "postgres://environment",
		"BOOTSTRAP_TRANSACTIONS": "100",
		"BOOTSTRAP_ATTACHMENTS":  "10",
	}

	p, err := parseParams([]string{
		"-database-url", "postgres://flag",
		"-transactions", "20",
		"-attachments", "2",
	}, func(key string) string { return env[key] })

	require.NoError(t, err)
	assert.Equal(t, "postgres://flag", p.databaseURL)
	assert.Equal(t, 20, p.transactions)
	assert.Equal(t, 2, p.attachments)
}

func TestParseParamsRejectsMoreAttachmentsThanTransactions(t *testing.T) {
	_, err := parseParams([]string{"-transactions", "2", "-attachments", "3"}, func(key string) string {
		if key == "DATABASE_URL" {
			return "postgres://development"
		}
		return ""
	})

	assert.Error(t, err)
}
