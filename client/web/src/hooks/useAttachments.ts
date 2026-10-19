import { useMutation, useQueryClient } from '@tanstack/react-query'
import {
  deleteDraftAttachment,
  deleteTransactionAttachment,
  extractTransactionPrefill,
  uploadDraftAttachment,
  uploadTransactionAttachment,
} from '@/api/attachments'

/**
 * Draft mutations don't invalidate anything: a draft has no query of its own, and createTransaction
 * (useTransactions) already invalidates the transaction/summary queries when the draft is linked.
 * The read is driven imperatively (with an AbortSignal), not cached.
 */
export function useUploadDraftAttachment() {
  return useMutation({ mutationFn: uploadDraftAttachment })
}

/**
 * Uploads a file directly onto an existing transaction. Refreshes the transaction queries so the
 * detail screen's attachment list picks up the new row after the server responds.
 */
export function useUploadTransactionAttachment() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ transactionId, file }: { transactionId: number; file: File }) =>
      uploadTransactionAttachment(transactionId, file),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['transactions'] })
      queryClient.invalidateQueries({ queryKey: ['summary'] })
    },
  })
}

export function useDeleteDraftAttachment() {
  return useMutation({ mutationFn: deleteDraftAttachment })
}

/** Reads an uploaded draft into proposed field values. Pass a `signal` to cancel. */
export function useExtractTransactionPrefill() {
  return useMutation({
    mutationFn: ({ attachmentIds, signal }: { attachmentIds: string[]; signal?: AbortSignal }) =>
      extractTransactionPrefill(attachmentIds, signal),
  })
}

/**
 * Deletes an attachment already linked to a transaction. Refreshes the transaction queries (list +
 * the `['transactions', 'detail', id]` prefix) and the summary, since the record's attachments — and
 * any totals derived from it — changed.
 */
export function useDeleteTransactionAttachment() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ transactionId, attachmentId }: { transactionId: number; attachmentId: string }) =>
      deleteTransactionAttachment(transactionId, attachmentId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['transactions'] })
      queryClient.invalidateQueries({ queryKey: ['summary'] })
    },
  })
}
