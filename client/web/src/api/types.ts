/** DTOs mirroring server/internal/adapter/schema/*.go — the HTTP wire shapes. */

export type TransactionType = 'income' | 'expense'
export type TransactionSource = 'manual' | 'recurring'
export type SummaryPeriod = 'day' | 'week' | 'month' | 'year'

export interface UserDto {
  id: string
  username: string
  display_name: string
  created_at: string
}

export interface CurrencyDto {
  code: string
  name: string
  is_base: boolean
}

export interface TransactionCategoryDto {
  id: string
  type: TransactionType
  name: string
  active: boolean
  sort_order: number
}

export interface TransactionDto {
  id: number
  type: TransactionType
  amount: string
  currency_code: string
  category_id: string
  category_name: string
  user_id: string
  source: TransactionSource
  name: string
  created_at: string
  description: string
}

export interface CurrencyStatsDto {
  currency_code: string
  total_income: string
  total_expense: string
  net: string
}

export interface SummaryDto {
  period: SummaryPeriod | null
  period_start: string | null
  period_end: string
  currency_code: string
  total_income: string
  total_expense: string
  net: string
  by_currency: Record<string, CurrencyStatsDto>
  transaction_count: number
}
