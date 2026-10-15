package domain

// TransactionCategoryType and TransactionType are kept as two distinct
// named types with the same two values, mirroring the original two
// independent enums — a transaction's Type must match its Category.Type,
// which Validate on Transaction checks explicitly rather than relying on
// the two types being identical.
type TransactionCategoryType string

const (
	TransactionCategoryTypeIncome  TransactionCategoryType = "income"
	TransactionCategoryTypeExpense TransactionCategoryType = "expense"
)

func (t TransactionCategoryType) Valid() bool {
	return t == TransactionCategoryTypeIncome || t == TransactionCategoryTypeExpense
}

type TransactionType string

const (
	TransactionTypeIncome  TransactionType = "income"
	TransactionTypeExpense TransactionType = "expense"
)

func (t TransactionType) Valid() bool {
	return t == TransactionTypeIncome || t == TransactionTypeExpense
}

type TransactionSource string

const (
	TransactionSourceManual    TransactionSource = "manual"
	TransactionSourceRecurring TransactionSource = "recurring"
)

func (s TransactionSource) Valid() bool {
	switch s {
	case TransactionSourceManual, TransactionSourceRecurring:
		return true
	default:
		return false
	}
}

type SummaryPeriod string

const (
	SummaryPeriodDay   SummaryPeriod = "day"
	SummaryPeriodWeek  SummaryPeriod = "week"
	SummaryPeriodMonth SummaryPeriod = "month"
	SummaryPeriodYear  SummaryPeriod = "year"
)

func (p SummaryPeriod) Valid() bool {
	switch p {
	case SummaryPeriodDay, SummaryPeriodWeek, SummaryPeriodMonth, SummaryPeriodYear:
		return true
	default:
		return false
	}
}
