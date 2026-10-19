import { useMutation, useQueryClient } from '@tanstack/react-query'
import {
  createSubTransaction,
  deleteSubTransaction,
  updateSubTransaction,
  type SubTransactionInput,
} from '@/api/subTransactions'

/**
 * Line-item mutations refresh the transaction queries (list + the `['transactions', 'detail', id]`
 * prefix) and the summary: the server recomputes the transaction's amount from its items, so both
 * the record's items and the totals derived from it changed.
 */
export function useCreateSubTransaction() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ transactionId, input }: { transactionId: string | number; input: SubTransactionInput }) =>
      createSubTransaction(transactionId, input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['transactions'] })
      queryClient.invalidateQueries({ queryKey: ['summary'] })
    },
  })
}

export function useUpdateSubTransaction() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({
      transactionId,
      subTransactionId,
      input,
    }: {
      transactionId: string | number
      subTransactionId: string
      input: Partial<SubTransactionInput>
    }) => updateSubTransaction(transactionId, subTransactionId, input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['transactions'] })
      queryClient.invalidateQueries({ queryKey: ['summary'] })
    },
  })
}

export function useDeleteSubTransaction() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({
      transactionId,
      subTransactionId,
    }: {
      transactionId: string | number
      subTransactionId: string
    }) => deleteSubTransaction(transactionId, subTransactionId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['transactions'] })
      queryClient.invalidateQueries({ queryKey: ['summary'] })
    },
  })
}
