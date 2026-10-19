import { apiDownload, apiFetch, apiUrl } from '@/api/client'
import { mockApi, mockDelay } from '@/api/mock/data'
import type { TransactionAttachmentDto, TransactionPrefillDto } from '@/api/types'
import { isDesignMode } from '@/lib/designMode'

/**
 * Uploads one file as an unattached draft (transaction_id null). The server normalizes supported
 * image formats before storage. The returned id is passed as `attachment_ids` when creating the
 * transaction, which links the draft atomically. The server streams the body and takes the original
 * filename/media type from the multipart part, so the FormData part must be named `file`.
 */
export function uploadDraftAttachment(file: File): Promise<TransactionAttachmentDto> {
  if (isDesignMode()) return mockDelay(mockApi.uploadDraftAttachment(file))
  const body = new FormData()
  body.append('file', file)
  return apiFetch<TransactionAttachmentDto>('/transaction-attachments', { method: 'POST', body })
}

/**
 * Uploads one file directly onto an existing transaction (the transaction-first path). The server
 * stores the object, inserts the linked metadata row, and the GET detail response then includes it.
 */
export function uploadTransactionAttachment(transactionId: number, file: File): Promise<TransactionAttachmentDto> {
  if (isDesignMode()) return mockDelay(mockApi.uploadTransactionAttachment(transactionId, file))
  const body = new FormData()
  body.append('file', file)
  return apiFetch<TransactionAttachmentDto>(`/transactions/${transactionId}/attachments`, { method: 'POST', body })
}

/** Deletes an unattached draft (used when the user removes an attachment before saving). */
export function deleteDraftAttachment(attachmentId: string): Promise<void> {
  if (isDesignMode()) return mockDelay(mockApi.deleteDraftAttachment(attachmentId))
  return apiFetch<void>(`/transaction-attachments/${attachmentId}`, { method: 'DELETE' })
}

/** Downloads one draft attachment's bytes (protected route, forced attachment disposition). */
export function downloadDraftAttachment(attachmentId: string): Promise<Blob> {
  if (isDesignMode()) return mockDelay(mockApi.downloadAttachment())
  return apiDownload(`/transaction-attachments/${attachmentId}/download`)
}

/** Deletes one attachment that is already linked to a transaction (transaction-scoped route). */
export function deleteTransactionAttachment(transactionId: number, attachmentId: string): Promise<void> {
  if (isDesignMode()) return mockDelay(mockApi.deleteTransactionAttachment(transactionId, attachmentId))
  return apiFetch<void>(`/transactions/${transactionId}/attachments/${attachmentId}`, { method: 'DELETE' })
}

/** Tiny inline placeholder so previews still render in design mode, where there's no backend. */
const DESIGN_MODE_PREVIEW_URL = `data:image/svg+xml;utf8,${encodeURIComponent(
  '<svg xmlns="http://www.w3.org/2000/svg" width="96" height="96"><rect width="96" height="96" fill="#e6e8ea"/><path d="M22 64l16-16 12 12 9-9 15 15v10H22z" fill="#9aa3a8"/><circle cx="36" cy="34" r="7" fill="#9aa3a8"/></svg>',
)}`

/**
 * URL for a short-lived inline preview of an attachment, built from the `preview_grant` that
 * `GET /transactions/{id}` mints for previewable media. It points at the public,
 * grant-authenticated content route, so it is safe to use directly as an `<img src>` or a
 * new-tab link — unlike the protected download route, it needs no auth header.
 */
export function attachmentPreviewUrl(previewGrant: string): string {
  if (isDesignMode()) return DESIGN_MODE_PREVIEW_URL
  return apiUrl(`/attachment-content?grant=${encodeURIComponent(previewGrant)}`)
}

/**
 * Downloads one attachment's bytes through the transaction-scoped route (forced attachment
 * disposition). This is the documented bytes-not-JSON exception, so it returns a `Blob` for the
 * caller to save rather than parsed data.
 */
export function downloadTransactionAttachment(transactionId: number, attachmentId: string): Promise<Blob> {
  if (isDesignMode()) return mockDelay(mockApi.downloadAttachment())
  return apiDownload(`/transactions/${transactionId}/attachments/${attachmentId}/download`)
}

/**
 * Reads an already-uploaded draft and proposes transaction field values. Synchronous: it resolves
 * when the read finishes and rejects on failure/timeout, so the caller drives the waiting UI. The
 * file stays attached whatever the outcome — a slow or failed read never costs the user their
 * upload. `signal` is the cancel path; `apiFetch` spreads `init`, so no change to `client.ts` is
 * needed.
 */
export function extractTransactionPrefill(attachmentIds: string[], signal?: AbortSignal): Promise<TransactionPrefillDto> {
  if (isDesignMode()) return mockDelay(mockApi.extractTransactionPrefill(attachmentIds), 1400)
  return apiFetch<TransactionPrefillDto>('/transaction-attachments/extraction', {
    method: 'POST',
    body: JSON.stringify({ attachment_ids: attachmentIds }),
    signal,
  })
}
