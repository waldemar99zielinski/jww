import { apiFetch } from '@/api/client'
import { mockApi, mockDelay } from '@/api/mock/data'
import type { TransactionDto } from '@/api/types'
import { isDesignMode } from '@/lib/designMode'

/**
 * One line item's editable fields. `amount` is whole minor units (cents), the
 * canonical wire unit for every Money value.
 */
export type SubTransactionInput = { name: string; label: string; amount: number }

/**
 * Adds one line item to an existing transaction. The server recomputes the transaction's amount as
 * the sum of its items and returns the updated transaction.
 */
export function createSubTransaction(transactionId: string | number, input: SubTransactionInput): Promise<TransactionDto> {
  if (isDesignMode()) return mockDelay(mockApi.createSubTransaction(transactionId, input))
  return apiFetch<TransactionDto>(`/transactions/${transactionId}/sub-transactions`, {
    method: 'POST',
    body: JSON.stringify(input),
  })
}

/** Partial update: only the provided fields change. Returns the updated transaction. */
export function updateSubTransaction(
  transactionId: string | number,
  subTransactionId: string,
  input: Partial<SubTransactionInput>,
): Promise<TransactionDto> {
  if (isDesignMode()) return mockDelay(mockApi.updateSubTransaction(transactionId, subTransactionId, input))
  return apiFetch<TransactionDto>(`/transactions/${transactionId}/sub-transactions/${subTransactionId}`, {
    method: 'PATCH',
    body: JSON.stringify(input),
  })
}

/**
 * Removes one line item. The server recomputes the transaction's amount (or releases it back to
 * being independently editable when the last item goes), and returns 204 with no body.
 */
export function deleteSubTransaction(transactionId: string | number, subTransactionId: string): Promise<void> {
  if (isDesignMode()) return mockDelay(mockApi.deleteSubTransaction(transactionId, subTransactionId))
  return apiFetch<void>(`/transactions/${transactionId}/sub-transactions/${subTransactionId}`, {
    method: 'DELETE',
  })
}
