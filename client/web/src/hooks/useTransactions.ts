import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  createTransaction,
  deleteTransaction,
  getTransaction,
  listTransactions,
  updateTransaction,
  type TransactionFilters,
  type UpdateTransactionInput,
} from '@/api/transactions'

export function useTransactions(filters: TransactionFilters) {
  return useQuery({
    queryKey: ['transactions', filters],
    queryFn: () => listTransactions(filters),
  })
}

/** One transaction by id, for the detail screen. Disabled until `id` is a whole number. */
export function useTransaction(id: number) {
  return useQuery({
    queryKey: ['transactions', 'detail', id],
    queryFn: () => getTransaction(id),
    enabled: Number.isInteger(id),
  })
}

export function useCreateTransaction() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: createTransaction,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['transactions'] })
      queryClient.invalidateQueries({ queryKey: ['summary'] })
    },
  })
}

/**
 * Patches an existing transaction. Refreshes the transaction queries (list + the
 * `['transactions', 'detail', id]` prefix) and the summary, since a changed amount, type,
 * or currency moves totals.
 */
export function useUpdateTransaction() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, input }: { id: number; input: UpdateTransactionInput }) => updateTransaction(id, input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['transactions'] })
      queryClient.invalidateQueries({ queryKey: ['summary'] })
    },
  })
}

/**
 * Deletes a transaction along with its line items and attachments. The detail query is removed
 * (not invalidated, which would immediately refetch the now-404 record); the list and summary are
 * refreshed so the deleted row and its totals disappear.
 */
export function useDeleteTransaction() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: deleteTransaction,
    onSuccess: (_data, id) => {
      queryClient.removeQueries({ queryKey: ['transactions', 'detail', id] })
      queryClient.invalidateQueries({ queryKey: ['transactions'] })
      queryClient.invalidateQueries({ queryKey: ['summary'] })
    },
  })
}
