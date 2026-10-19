import type { TransactionSortOrder } from '@/api/transactions'
import type { TransactionCategoryDto, TransactionDto, TransactionType, UserDto } from '@/api/types'

export type TransactionFilterValue = TransactionType | 'all'

export interface TransactionsViewProps {
  filter: TransactionFilterValue
  onFilterChange: (filter: TransactionFilterValue) => void
  userFilter: string
  onUserFilterChange: (userId: string) => void
  users: UserDto[]
  categoryFilter: string
  onCategoryFilterChange: (categoryId: string) => void
  categories: TransactionCategoryDto[]
  nameFilter: string
  onNameFilterChange: (name: string) => void
  dateFrom: string
  onDateFromChange: (date: string) => void
  dateTo: string
  onDateToChange: (date: string) => void
  sortOrder: TransactionSortOrder
  onSortOrderChange: (sortOrder: TransactionSortOrder) => void
  grouped: [string, TransactionDto[]][]
  isEmpty: boolean
}
