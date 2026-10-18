import { apiFetch, toQueryString } from '@/api/client'
import { mockApi, mockDelay } from '@/api/mock/data'
import type { TransactionDetailsDto, TransactionDto, TransactionSource, TransactionType } from '@/api/types'
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
 * is whole minor units (cents); it is converted to the decimal the server's `Money` parses.
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
  const body: Record<string, unknown> = { ...input }
  if (input.amount !== undefined) body.amount = Number((input.amount / 100).toFixed(2))
  return apiFetch<TransactionDetailsDto>(`/transactions/${id}`, {
    method: 'PATCH',
    body: JSON.stringify(body),
  })
}
