package domain

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNotFoundErrors_KindIsNotFoundAndErrorsAsWorks covers every
// New...NotFoundError constructor: kernel.WriteError relies on errors.As
// finding a DomainError and on Kind() == ErrorKindNotFound to map it to 404,
// so both must hold for every constructor added here.
func TestNotFoundErrors_KindIsNotFoundAndErrorsAsWorks(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{"TransactionCategoryNotFoundError", NewTransactionCategoryNotFoundError(uuid.New())},
		{"UserNotFoundError", NewUserNotFoundError(uuid.New())},
		{"CurrencyNotFoundError", NewCurrencyNotFoundError("USD")},
		{"CurrencyExchangeRateNotFoundError", NewCurrencyExchangeRateNotFoundError("USD", "EUR")},
		{"TransactionNotFoundError", NewTransactionNotFoundError(42)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var domainErr DomainError
			require.True(t, errors.As(tt.err, &domainErr), "errors.As must find a DomainError")
			assert.Equal(t, ErrorKindNotFound, domainErr.Kind())
			assert.NotEmpty(t, domainErr.Error())
		})
	}
}

func TestAuthenticationError_KindIsUnauthorized(t *testing.T) {
	err := NewAuthenticationError()

	var domainErr DomainError
	require.True(t, errors.As(error(err), &domainErr))
	assert.Equal(t, ErrorKindUnauthorized, domainErr.Kind())
	assert.NotEmpty(t, domainErr.Error())
}

// TestAuthenticationError_DoesNotEnumerate guards the documented anti-enumeration
// property: the same error/message is used for "bad username" and "bad
// password" alike, so there's only one constructor and one message to check.
func TestAuthenticationError_DoesNotEnumerate(t *testing.T) {
	first := NewAuthenticationError()
	second := NewAuthenticationError()
	assert.Equal(t, first.Error(), second.Error())
}
