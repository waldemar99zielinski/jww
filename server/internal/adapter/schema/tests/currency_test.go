package schema_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"jww/internal/adapter/schema"
	"jww/internal/domain"
)

func TestCurrencyCreateRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     schema.CurrencyCreateRequest
		wantErr bool
	}{
		{"valid", schema.CurrencyCreateRequest{Code: "USD", Name: "US Dollar"}, false},
		{"code too short", schema.CurrencyCreateRequest{Code: "US", Name: "US Dollar"}, true},
		{"code too long", schema.CurrencyCreateRequest{Code: "USDX", Name: "US Dollar"}, true},
		{"blank name", schema.CurrencyCreateRequest{Code: "USD", Name: ""}, true},
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

func TestCurrencyUpdateRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     schema.CurrencyUpdateRequest
		wantErr bool
	}{
		{"valid", schema.CurrencyUpdateRequest{Name: "US Dollar", IsBase: true}, false},
		{"blank name", schema.CurrencyUpdateRequest{Name: ""}, true},
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

func TestCurrencyToResponse(t *testing.T) {
	c := domain.Currency{Code: "EUR", Name: "Euro", IsBase: true}
	resp := schema.CurrencyToResponse(c)
	assert.Equal(t, schema.CurrencyResponse{Code: "EUR", Name: "Euro", IsBase: true}, resp)
}
