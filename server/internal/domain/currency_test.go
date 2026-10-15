package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func validCurrency() Currency {
	return Currency{Code: "USD", Name: "US Dollar"}
}

func TestCurrency_Validate(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(c Currency) Currency
		wantErr string
	}{
		{"valid", func(c Currency) Currency { return c }, ""},
		{"lowercase code rejected", func(c Currency) Currency { c.Code = "usd"; return c }, "code"},
		{"wrong length code rejected", func(c Currency) Currency { c.Code = "US"; return c }, "code"},
		{"non-alpha code rejected", func(c Currency) Currency { c.Code = "US1"; return c }, "code"},
		{"blank name", func(c Currency) Currency { c.Name = "  "; return c }, "name"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.mutate(validCurrency()).Validate()
			if tt.wantErr == "" {
				assert.NoError(t, err)
				return
			}
			assert.ErrorContains(t, err, tt.wantErr)
		})
	}
}

func TestIsUpperAlpha(t *testing.T) {
	assert.True(t, isUpperAlpha("USD"))
	assert.False(t, isUpperAlpha("usd"))
	assert.False(t, isUpperAlpha("US1"))
	assert.True(t, isUpperAlpha(""))
}
