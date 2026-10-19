import { apiFetch, toQueryString } from '@/api/client'
import { mockApi, mockDelay } from '@/api/mock/data'
import type { TransactionDetailsDto, TransactionDto, TransactionSource, TransactionType } from '@/api/types'
import { isDesignMode } from '@/lib/designMode'

export type TransactionSortKey = 'date'
export type TransactionSortOrder = 'asc' | 'desc'

export interface TransactionFilters {
  user_id?: string
  category_id?: string
  type?: TransactionType
  currency_code?: string
  source?: TransactionSource
  created_from?: string
  created_to?: string
  /** Column to order by. Only `date` is supported; omitted means `date`. */
  sort_key?: TransactionSortKey
  /** Sort direction. Omitted means the server's default, `desc`. */
  sort_order?: TransactionSortOrder
}

export function listTransactions(filters: TransactionFilters): Promise<TransactionDto[]> {
  if (isDesignMode()) return mockDelay(mockApi.listTransactions(filters))
  return apiFetch<TransactionDto[]>(`/transactions${toQueryString(filters)}`)
}

/** Fetches one transaction, with per-attachment preview grants, for the detail screen. */
export function getTransaction(id: number): Promise<TransactionDetailsDto> {
  if (isDesignMode()) return mockDelay(mockApi.getTransaction(id))
  return apiFetch<TransactionDetailsDto>(`/transactions/${id}`)
}

export interface CreateTransactionInput {
  type: TransactionType
  /** Whole minor units (cents): 42.50 is sent as 4250. Decimals are display-only. */
  amount: number
  currency_code: string
  category_id: string
  user_id: string
  source: TransactionSource
  name: string
  description?: string
  /** The transaction's date (RFC3339). When omitted the server stamps now(). */
  created_at?: string
  /** Ids of previously uploaded draft attachments to link to the new transaction. */
  attachment_ids?: string[]
  /**
   * Line items to create with the transaction, each amount in whole minor units (cents) like the
   * top-level `amount`. When present and non-empty the server derives the transaction amount as
   * their sum, so `amount` may still be sent but is ignored.
   */
  sub_transactions?: { name: string; label: string; amount: number }[]
}

export function createTransaction(input: CreateTransactionInput): Promise<TransactionDto> {
  if (isDesignMode()) return mockDelay(mockApi.createTransaction(input))
  return apiFetch<TransactionDto>('/transactions', {
    method: 'POST',
    body: JSON.stringify(input),
  })
}

/**
 * Partial update: only the provided fields change. `amount`, like `CreateTransactionInput.amount`,
 * is whole minor units (cents), the canonical Money wire unit.
 * Not patchable here: `id` and attachment membership (use the attachment routes).
 */
export interface UpdateTransactionInput {
  type?: TransactionType
  /** Whole minor units (cents). */
  amount?: number
  currency_code?: string
  category_id?: string
  user_id?: string
  source?: TransactionSource
  name?: string
  description?: string
  /** The transaction's date, as the RFC3339 string `TransactionDto.created_at` carries. */
  created_at?: string
}

export function updateTransaction(id: number, input: UpdateTransactionInput): Promise<TransactionDetailsDto> {
  if (isDesignMode()) return mockDelay(mockApi.updateTransaction(id, input))
  return apiFetch<TransactionDetailsDto>(`/transactions/${id}`, {
    method: 'PATCH',
    body: JSON.stringify(input),
  })
}

/**
 * Deletes a transaction and everything it owns. The server removes the linked attachments
 * (including their stored objects) and the database cascades the line items.
 */
export function deleteTransaction(id: number): Promise<void> {
  if (isDesignMode()) return mockDelay(mockApi.deleteTransaction(id))
  return apiFetch<void>(`/transactions/${id}`, { method: 'DELETE' })
}
