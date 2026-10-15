import { apiFetch, toQueryString } from '@/api/client'
import { mockApi, mockDelay } from '@/api/mock/data'
import type { TransactionDto, TransactionSource, TransactionType } from '@/api/types'
import { isDesignMode } from '@/lib/designMode'

export interface TransactionFilters {
  user_id?: string
  category_id?: string
  type?: TransactionType
  currency_code?: string
  source?: TransactionSource
  created_from?: string
  created_to?: string
}

export function listTransactions(filters: TransactionFilters): Promise<TransactionDto[]> {
  if (isDesignMode()) return mockDelay(mockApi.listTransactions(filters))
  return apiFetch<TransactionDto[]>(`/transactions${toQueryString(filters)}`)
}

export interface CreateTransactionInput {
  type: TransactionType
  amount: number
  currency_code: string
  category_id: string
  user_id: string
  source: TransactionSource
  name: string
  description?: string
}

export function createTransaction(input: CreateTransactionInput): Promise<TransactionDto> {
  if (isDesignMode()) return mockDelay(mockApi.createTransaction(input))
  return apiFetch<TransactionDto>('/transactions', {
    method: 'POST',
    body: JSON.stringify(input),
  })
}
