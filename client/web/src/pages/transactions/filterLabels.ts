import type { TransactionFilterValue } from './types'

/** Shared between TransactionFiltersDropdown and TransactionFiltersSummary. */
export const filterLabels: Record<TransactionFilterValue, string> = {
  all: 'All',
  income: 'Income',
  expense: 'Expenses',
}
