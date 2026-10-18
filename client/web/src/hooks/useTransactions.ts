import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  createTransaction,
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
