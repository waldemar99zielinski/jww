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
		{"TransactionAttachmentNotFoundError", NewTransactionAttachmentNotFoundError(uuid.New())},
		{"TransactionSubTransactionNotFoundError", NewTransactionSubTransactionNotFoundError(uuid.New())},
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

// TestTransactionSubTransactionNotFoundError_CarriesIdentity mirrors the
// TransactionAttachmentNotFoundError precedent: the error carries only the
// sub-transaction identity (no parent transaction id), so errors.As can pull
// the offending id back out and the message names it.
func TestTransactionSubTransactionNotFoundError_CarriesIdentity(t *testing.T) {
	id := uuid.New()
	err := NewTransactionSubTransactionNotFoundError(id)

	var domainErr DomainError
	require.True(t, errors.As(error(err), &domainErr))
	assert.Equal(t, ErrorKindNotFound, domainErr.Kind())

	var notFound *TransactionSubTransactionNotFoundError
	require.True(t, errors.As(error(err), &notFound))
	assert.Equal(t, id, notFound.SubTransactionID)
	assert.ErrorContains(t, err, id.String())
}

func TestAuthenticationError_KindIsUnauthorized(t *testing.T) {
	err := NewAuthenticationError()

	var domainErr DomainError
	require.True(t, errors.As(error(err), &domainErr))
	assert.Equal(t, ErrorKindUnauthorized, domainErr.Kind())
	assert.NotEmpty(t, domainErr.Error())
}

func TestAttachmentErrors_KindMapping(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want ErrorKind
	}{
		{"already linked is conflict", NewAttachmentAlreadyLinkedError(uuid.New()), ErrorKindConflict},
		{"validation", NewValidationError("bad input"), ErrorKindValidation},
		{"storage unavailable", NewObjectStorageUnavailableError(errors.New("dial tcp")), ErrorKindUnavailable},
		{"integrity", NewAttachmentIntegrityError(uuid.New(), errors.New("no such key")), ErrorKindInternal},
		// A non-image attachment is a validation failure (400), checked before
		// any grant is issued.
		{"unsupported media type", NewUnsupportedAttachmentMediaTypeError("application/pdf"), ErrorKindValidation},
		// Provider refusal and provider outage share one kind (503).
		{"extraction failed", NewAttachmentExtractionFailedError(errors.New("provider 500")), ErrorKindUnavailable},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var domainErr DomainError
			require.True(t, errors.As(tt.err, &domainErr))
			assert.Equal(t, tt.want, domainErr.Kind())
			assert.NotEmpty(t, domainErr.Error())
		})
	}
}

// TestStorageUnavailableError_DoesNotLeakCause guards the client-facing message:
// provider/endpoint detail must not appear in Error(), only via Unwrap.
func TestStorageUnavailableError_DoesNotLeakCause(t *testing.T) {
	err := NewObjectStorageUnavailableError(errors.New("https://minio.internal:9000 bucket=secret"))
	assert.NotContains(t, err.Error(), "minio.internal")
	assert.NotContains(t, err.Error(), "secret")
	assert.ErrorContains(t, errors.Unwrap(err), "minio.internal")
}

// TestAuthenticationError_DoesNotEnumerate guards the documented anti-enumeration
// property: the same error/message is used for "bad username" and "bad
// password" alike, so there's only one constructor and one message to check.
func TestAuthenticationError_DoesNotEnumerate(t *testing.T) {
	first := NewAuthenticationError()
	second := NewAuthenticationError()
	assert.Equal(t, first.Error(), second.Error())
}

// TestExtractionFailedError_WrapsCause keeps the same posture as the storage
// error: the client-facing message stays generic (never the provider's raw
// response), while the underlying cause stays inspectable for boundary logging.
func TestExtractionFailedError_WrapsCause(t *testing.T) {
	cause := errors.New("openai: unexpected status 429 body={...raw provider payload...}")
	err := NewAttachmentExtractionFailedError(cause)

	var domainErr DomainError
	require.True(t, errors.As(error(err), &domainErr))
	assert.Equal(t, ErrorKindUnavailable, domainErr.Kind())
	assert.NotContains(t, err.Error(), "raw provider payload")
	assert.ErrorContains(t, errors.Unwrap(err), "429")
}
