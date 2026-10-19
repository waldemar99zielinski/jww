import type { TransactionDto } from '@/api/types'

export interface CategoryBreakdownItem {
  categoryName: string
  amount: number
  percent: number
}

export interface HomeViewProps {
  displayName: string
  monthLabel: string
  /** Currency the summary totals are reported in (the configured base currency). */
  currencyCode: string
  balance: number
  totalIncome: number
  totalExpense: number
  breakdown: CategoryBreakdownItem[]
  recent: TransactionDto[]
}
