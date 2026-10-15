package kernel

import "errors"

// Shared query-parameter validation errors, reused by more than one
// resource subpackage (transaction_category validates a
// TransactionCategoryType "type" param; transaction validates both a
// TransactionType "type" param and a TransactionSource "source" param —
// TransactionCategoryType and TransactionType share the same two string
// values, so one error fits both).
var (
	ErrInvalidType   = errors.New("type must be 'income' or 'expense'")
	ErrInvalidSource = errors.New("source must be 'manual' or 'recurring'")
)
