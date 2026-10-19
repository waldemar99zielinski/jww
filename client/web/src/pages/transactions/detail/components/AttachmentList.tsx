import { useState } from 'react'
import { IconDots, IconPaperclip, IconSpinner } from '@/components/icons'
import { IconButton } from '@/components/ui/IconButton'
import { formatBytes } from '@/lib/formatBytes'
import { previewKindFor } from '@/lib/previewKind'
import type { TransactionAttachmentItem } from '../types'

export interface AttachmentListProps {
  attachments: TransactionAttachmentItem[]
  downloadError: boolean
  onPreview: (attachment: TransactionAttachmentItem) => void
  onOpenActions: (attachment: TransactionAttachmentItem) => void
}

/**
 * Attachment rows. Clicking a row previews it big (for any type the browser can display); the
 * trailing button opens the actions drawer (`AttachmentActions`) with preview/download/delete.
 *
 * The attachment *metadata* arrives with the transaction, but an image's bytes are fetched by the
 * browser afterwards, so the thumbnail slot shows a spinner until the `<img>` loads. Some allowed
 * media types (notably DNG, and HEIC outside Safari) cannot be decoded at all, so a failed load
 * falls back to the paperclip icon.
 */
export function AttachmentList({ attachments, downloadError, onPreview, onOpenActions }: AttachmentListProps) {
  const [failedPreviewIds, setFailedPreviewIds] = useState<ReadonlySet<string>>(() => new Set())
  const [loadedPreviewIds, setLoadedPreviewIds] = useState<ReadonlySet<string>>(() => new Set())

  function markPreviewFailed(id: string) {
    setFailedPreviewIds((current) => {
      if (current.has(id)) return current
      const next = new Set(current)
      next.add(id)
      return next
    })
  }

  function markPreviewLoaded(id: string) {
    setLoadedPreviewIds((current) => {
      if (current.has(id)) return current
      const next = new Set(current)
      next.add(id)
      return next
    })
  }

  return (
    <div className="flex flex-col gap-2">
      <ul className="m-0 flex list-none flex-col gap-2 p-0">
        {attachments.map((attachment) => {
          const showThumbnail =
            Boolean(attachment.previewUrl) && attachment.mediaType.startsWith('image/') && !failedPreviewIds.has(attachment.id)
          // Metadata is in hand, but the image bytes may still be in flight: keep the slot busy
          // until onLoad, rather than showing an empty box.
          const previewLoading = showThumbnail && !loadedPreviewIds.has(attachment.id)
          const canPreview = previewKindFor(attachment.mediaType) !== null
          return (
            <li
              key={attachment.id}
              className="flex items-center gap-3 rounded-xl border border-line bg-surface-sunken px-3.5 py-2.5"
            >
              <button
                type="button"
                onClick={() => (canPreview ? onPreview(attachment) : onOpenActions(attachment))}
                aria-label={canPreview ? `Preview ${attachment.filename}` : `${attachment.filename} actions`}
                className="flex min-w-0 flex-1 items-center gap-3 text-left"
              >
                {showThumbnail && attachment.previewUrl ? (
                  <span className="relative flex h-12 w-12 flex-none">
                    <img
                      src={attachment.previewUrl}
                      alt={`Preview of ${attachment.filename}`}
                      onLoad={() => markPreviewLoaded(attachment.id)}
                      onError={() => markPreviewFailed(attachment.id)}
                      className={`h-12 w-12 rounded-lg object-cover ${previewLoading ? 'invisible' : ''}`}
                    />
                    {previewLoading && (
                      <span
                        role="status"
                        aria-label={`Loading preview of ${attachment.filename}`}
                        className="absolute inset-0 flex items-center justify-center text-muted"
                      >
                        <IconSpinner className="h-4 w-4 animate-spin" />
                      </span>
                    )}
                  </span>
                ) : (
                  <span className="flex h-12 w-12 flex-none items-center justify-center rounded-lg bg-surface text-muted">
                    <IconPaperclip className="h-4 w-4" />
                  </span>
                )}

                <span className="flex min-w-0 flex-1 flex-col">
                  <span className="truncate text-[13px] font-medium text-ink">{attachment.filename}</span>
                  <span className="truncate text-xs text-muted">
                    {formatBytes(attachment.byteSize)} · {attachment.mediaType}
                  </span>
                </span>
              </button>

              <IconButton
                aria-label={`Actions for ${attachment.filename}`}
                onClick={() => onOpenActions(attachment)}
                icon={<IconDots className="h-4 w-4" />}
                className="h-8 w-8 flex-none"
              />
            </li>
          )
        })}
      </ul>
      {downloadError && <p className="m-0 text-[11.5px] text-critical">Couldn't download that file — try again.</p>}
    </div>
  )
}
