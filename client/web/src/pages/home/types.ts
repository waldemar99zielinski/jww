import type { TransactionDto } from '@/api/types'

export interface CategoryBreakdownItem {
  categoryName: string
  amount: number
  percent: number
}

export interface HomeViewProps {
  displayName: string
  monthLabel: string
  balance: number
  totalIncome: number
  totalExpense: number
  breakdown: CategoryBreakdownItem[]
  recent: TransactionDto[]
}
