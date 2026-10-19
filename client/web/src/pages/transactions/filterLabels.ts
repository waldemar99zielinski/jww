import type { TransactionSortOrder } from '@/api/transactions'
import type { TransactionFilterValue } from './types'

/** Shared between TransactionFiltersDropdown and TransactionFiltersSummary. */
export const filterLabels: Record<TransactionFilterValue, string> = {
  all: 'All',
  income: 'Income',
  expense: 'Expenses',
}

export const sortOrderLabels: Record<TransactionSortOrder, string> = {
  asc: 'Ascending',
  desc: 'Descending',
}
