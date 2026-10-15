package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTransactionCategoryType_Valid(t *testing.T) {
	assert.True(t, TransactionCategoryTypeIncome.Valid())
	assert.True(t, TransactionCategoryTypeExpense.Valid())
	assert.False(t, TransactionCategoryType("bogus").Valid())
	assert.False(t, TransactionCategoryType("").Valid())
}

func TestTransactionType_Valid(t *testing.T) {
	assert.True(t, TransactionTypeIncome.Valid())
	assert.True(t, TransactionTypeExpense.Valid())
	assert.False(t, TransactionType("bogus").Valid())
	assert.False(t, TransactionType("").Valid())
}

func TestTransactionSource_Valid(t *testing.T) {
	assert.True(t, TransactionSourceManual.Valid())
	assert.True(t, TransactionSourceRecurring.Valid())
	assert.False(t, TransactionSource("bogus").Valid())
	assert.False(t, TransactionSource("").Valid())
}

func TestSummaryPeriod_Valid(t *testing.T) {
	assert.True(t, SummaryPeriodDay.Valid())
	assert.True(t, SummaryPeriodWeek.Valid())
	assert.True(t, SummaryPeriodMonth.Valid())
	assert.True(t, SummaryPeriodYear.Valid())
	assert.False(t, SummaryPeriod("bogus").Valid())
	assert.False(t, SummaryPeriod("").Valid())
}
