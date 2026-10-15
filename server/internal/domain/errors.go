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
