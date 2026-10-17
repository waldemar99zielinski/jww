package domain

import (
	"fmt"

	"github.com/google/uuid"
)

// ErrorKind classifies a DomainError for transport without teaching the domain
// layer about HTTP. The HTTP kernel maps each kind to a status code and a
// machine-readable code, so a new error type only has to pick a kind.
type ErrorKind int

const (
	// ErrorKindNotFound means a requested entity does not exist. It is the
	// zero value, so an unclassified DomainError reports as 404 by default.
	ErrorKindNotFound ErrorKind = iota
	// ErrorKindConflict means the request conflicts with current state,
	// e.g. a duplicate that violates a uniqueness rule.
	ErrorKindConflict
	// ErrorKindValidation means the input violates a business rule.
	ErrorKindValidation
	// ErrorKindUnauthorized means a credential is missing or invalid.
	ErrorKindUnauthorized
	// ErrorKindInternal means an unexpected server-side failure.
	ErrorKindInternal
	// ErrorKindUnavailable means a dependency (e.g. object storage) is
	// temporarily unavailable and the caller may retry. The HTTP kernel maps
	// it to 503.
	ErrorKindUnavailable
)

// DomainError marks every error type below so the HTTP layer can recognize
// them with a single type switch/assertion without knowing each concrete
// type individually.
type DomainError interface {
	error
	isDomainError()
	// Kind classifies the error so the transport layer can map it.
	Kind() ErrorKind
}

type baseDomainError struct {
	kind    ErrorKind
	message string
}

func (e baseDomainError) Error() string   { return e.message }
func (e baseDomainError) isDomainError()  {}
func (e baseDomainError) Kind() ErrorKind { return e.kind }

type TransactionCategoryNotFoundError struct {
	baseDomainError
	CategoryID uuid.UUID
}

func NewTransactionCategoryNotFoundError(categoryID uuid.UUID) *TransactionCategoryNotFoundError {
	return &TransactionCategoryNotFoundError{
		baseDomainError: baseDomainError{kind: ErrorKindNotFound, message: fmt.Sprintf("TransactionCategory %s not found", categoryID)},
		CategoryID:      categoryID,
	}
}

type UserNotFoundError struct {
	baseDomainError
	UserID uuid.UUID
}

func NewUserNotFoundError(userID uuid.UUID) *UserNotFoundError {
	return &UserNotFoundError{
		baseDomainError: baseDomainError{kind: ErrorKindNotFound, message: fmt.Sprintf("User %s not found", userID)},
		UserID:          userID,
	}
}

// AuthenticationError covers a failed login and an invalid/expired/unrecognized
// token alike — deliberately one error, one message, for both: distinguishing
// "wrong password" from "unknown username" (or a bad token from an expired
// one) in the response would let a caller enumerate valid usernames.
type AuthenticationError struct{ baseDomainError }

func NewAuthenticationError() *AuthenticationError {
	return &AuthenticationError{baseDomainError{kind: ErrorKindUnauthorized, message: "Invalid username, password, or token"}}
}

type CurrencyNotFoundError struct {
	baseDomainError
	Code string
}

func NewCurrencyNotFoundError(code string) *CurrencyNotFoundError {
	return &CurrencyNotFoundError{
		baseDomainError: baseDomainError{kind: ErrorKindNotFound, message: fmt.Sprintf("Currency %s not found", code)},
		Code:            code,
	}
}

type CurrencyExchangeRateNotFoundError struct {
	baseDomainError
	FromCurrencyCode string
	ToCurrencyCode   string
}

func NewCurrencyExchangeRateNotFoundError(fromCode, toCode string) *CurrencyExchangeRateNotFoundError {
	return &CurrencyExchangeRateNotFoundError{
		baseDomainError:  baseDomainError{kind: ErrorKindNotFound, message: fmt.Sprintf("CurrencyExchangeRate %s->%s not found", fromCode, toCode)},
		FromCurrencyCode: fromCode,
		ToCurrencyCode:   toCode,
	}
}

type TransactionNotFoundError struct {
	baseDomainError
	TransactionID int64
}

func NewTransactionNotFoundError(transactionID int64) *TransactionNotFoundError {
	return &TransactionNotFoundError{
		baseDomainError: baseDomainError{kind: ErrorKindNotFound, message: fmt.Sprintf("Transaction %d not found", transactionID)},
		TransactionID:   transactionID,
	}
}

type TransactionAttachmentNotFoundError struct {
	baseDomainError
	AttachmentID uuid.UUID
}

func NewTransactionAttachmentNotFoundError(attachmentID uuid.UUID) *TransactionAttachmentNotFoundError {
	return &TransactionAttachmentNotFoundError{
		baseDomainError: baseDomainError{kind: ErrorKindNotFound, message: fmt.Sprintf("TransactionAttachment %s not found", attachmentID)},
		AttachmentID:    attachmentID,
	}
}

// AttachmentAlreadyLinkedError means an attachment that is already attached to
// a transaction was claimed again by a different one. Linking is one-way and
// single-transaction, so this is a conflict, not a validation failure.
type AttachmentAlreadyLinkedError struct {
	baseDomainError
	AttachmentID uuid.UUID
}

func NewAttachmentAlreadyLinkedError(attachmentID uuid.UUID) *AttachmentAlreadyLinkedError {
	return &AttachmentAlreadyLinkedError{
		baseDomainError: baseDomainError{kind: ErrorKindConflict, message: fmt.Sprintf("TransactionAttachment %s is already linked to a transaction", attachmentID)},
		AttachmentID:    attachmentID,
	}
}

// ObjectStorageUnavailableError means the object store could not be reached or
// rejected an operation transiently. Its message is deliberately generic (no
// endpoint, key, or provider detail) because it is returned to clients; the
// underlying cause is available via Unwrap for boundary logging.
type ObjectStorageUnavailableError struct {
	baseDomainError
	cause error
}

func NewObjectStorageUnavailableError(cause error) *ObjectStorageUnavailableError {
	return &ObjectStorageUnavailableError{
		baseDomainError: baseDomainError{kind: ErrorKindUnavailable, message: "attachment storage is temporarily unavailable"},
		cause:           cause,
	}
}

func (e *ObjectStorageUnavailableError) Unwrap() error { return e.cause }

// AttachmentIntegrityError means the metadata row exists but its bytes could
// not be read: a missing object with a live row is an integrity fault, not a
// 404 on the attachment. Kind is internal so the kernel returns 500.
type AttachmentIntegrityError struct {
	baseDomainError
	AttachmentID uuid.UUID
	cause        error
}

func NewAttachmentIntegrityError(attachmentID uuid.UUID, cause error) *AttachmentIntegrityError {
	return &AttachmentIntegrityError{
		baseDomainError: baseDomainError{kind: ErrorKindInternal, message: fmt.Sprintf("TransactionAttachment %s has no readable object", attachmentID)},
		AttachmentID:    attachmentID,
		cause:           cause,
	}
}

func (e *AttachmentIntegrityError) Unwrap() error { return e.cause }

// ValidationError is a service-level business-rule failure that the HTTP kernel
// maps to 400. Domain `Validate()` methods return plain errors; this exists for
// a rule the domain type cannot express on its own (e.g. a cross-item constraint
// over a request list).
type ValidationError struct{ baseDomainError }

func NewValidationError(message string) *ValidationError {
	return &ValidationError{baseDomainError{kind: ErrorKindValidation, message: message}}
}
