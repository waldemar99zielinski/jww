import { useId, useState } from 'react'
import { IconChevronDown, IconClose, IconInfo, IconPaperclip, IconSparkle, IconSpinner } from '@/components/icons'
import { Button } from '@/components/ui/Button'
import { IconButton } from '@/components/ui/IconButton'
import { cx } from '@/components/ui/classes'
import { centsToAmountDisplay } from '@/lib/amount'
import type { ExtractionStatus, ExtractionSummary, TransactionAttachmentDraft } from '../types'

export interface AttachmentBarProps {
  attachments: TransactionAttachmentDraft[]
  status: ExtractionStatus
  summary?: ExtractionSummary
  /** At least one uploaded image is ready and no upload is in flight. */
  canScan: boolean
  canUndo: boolean
  disabled?: boolean
  onRemove: (id: string) => void
  /** Open the attachment lightbox for one uploaded file. */
  onPreview: (id: string) => void
  onScan: () => void
  onCancel: () => void
  onUndo: () => void
  /** Expand the sub-transactions panel after a scan that produced line items. */
  onReviewItems?: () => void
}

/**
 * Compact attachment status, shown only once a file is attached. The big empty upload surface is
 * gone: the header's attachment button starts the flow, and this bar stays out of the way until
 * there is something to report. Details (file list, warnings, undo) expand on demand.
 */
export function AttachmentBar({
  attachments,
  status,
  summary,
  canScan,
  canUndo,
  disabled = false,
  onRemove,
  onPreview,
  onScan,
  onCancel,
  onUndo,
  onReviewItems,
}: AttachmentBarProps) {
  const [expanded, setExpanded] = useState(false)
  const detailsId = useId()

  // Nothing attached and nothing read: the page stays clean; the header attachment button is the entry point.
  if (attachments.length === 0) return null

  const count = attachments.length
  const uploading = attachments.some((attachment) => attachment.status === 'uploading')
  const reading = status === 'reading'
  const itemCount = summary?.itemCount ?? 0
  const scannedIds = new Set(summary?.attachmentIds ?? [])
  const readyImages = attachments.filter(
    (attachment) => attachment.isImage && attachment.status === 'ready' && attachment.attachmentId,
  )
  const hasUnscannedFiles = Boolean(summary) && readyImages.some((attachment) => !scannedIds.has(attachment.attachmentId!))
  const needsAttention = Boolean(summary && (summary.warnings.length > 0 || summary.categoryHint))

  const title = reading
    ? 'Reading attachment…'
    : status === 'failed'
      ? "Couldn't read attachment"
      : status === 'success'
        ? itemCount > 0
          ? `${itemCount} ${itemCount === 1 ? 'item' : 'items'} added`
          : 'Attachment read'
        : uploading
          ? 'Uploading attachment…'
          : `${count} ${count === 1 ? 'attachment' : 'attachments'}`

  const subtitle = reading
    ? `Filling the form from ${count} ${count === 1 ? 'file' : 'files'}.`
    : status === 'failed'
      ? 'Files are still attached — try again.'
      : status === 'success'
        ? hasUnscannedFiles
          ? 'New files attached — scan again to include them.'
          : needsAttention
            ? 'Form filled — check the attachment details.'
            : summary?.attachmentTotal
              ? `Form filled · attachment total ${centsToAmountDisplay(summary.attachmentTotal)}`
              : 'Form filled from the attachment.'
        : uploading
          ? 'Sending your file to the server.'
          : 'Ready to scan.'

  const announcement = reading
    ? 'Reading the attachment.'
    : status === 'failed'
      ? "Couldn't read the attachment. The files remain attached."
      : status === 'success'
        ? `Attachment scanned. ${itemCount} ${itemCount === 1 ? 'line item' : 'line items'} added. The form has been filled in.`
        : ''

  return (
    <section
      className={cx(
        'flex flex-col rounded-xl border bg-surface',
        status === 'failed' ? 'border-critical/60' : 'border-line',
      )}
      aria-label="Attachment"
    >
      {announcement && (
        <p role="status" aria-live="polite" className="sr-only">
          {announcement}
        </p>
      )}

      <div
        className="flex cursor-pointer items-center gap-2.5 px-3 py-2.5"
        onClick={(event) => {
          // The whole row toggles details for pointer users; the buttons inside keep their own actions.
          if ((event.target as HTMLElement).closest('button')) return
          setExpanded((value) => !value)
        }}
      >
        <span
          className={cx(
            'flex h-8 w-8 flex-none items-center justify-center rounded-lg',
            status === 'success'
              ? 'bg-accent-soft text-accent-strong'
              : status === 'failed'
                ? 'bg-critical-soft text-critical'
                : 'bg-surface-sunken text-muted',
          )}
        >
          {reading ? (
            <IconSpinner className="h-4 w-4 animate-spin" />
          ) : status === 'failed' ? (
            <IconInfo className="h-4 w-4" />
          ) : status === 'success' ? (
            <IconSparkle className="h-4 w-4" />
          ) : (
            <IconPaperclip className="h-4 w-4" />
          )}
        </span>

        <div className="flex min-w-0 flex-1 flex-col">
          <span className="truncate text-[12.5px] font-semibold text-ink">{title}</span>
          <span className="truncate text-[11px] text-muted">{subtitle}</span>
        </div>

        <div className="flex flex-none items-center gap-1">
          {reading ? (
            <Button variant="ghost" onClick={onCancel} disabled={disabled} className="inline-flex h-8 items-center justify-center px-2.5 py-0 text-xs">
              Cancel
            </Button>
          ) : (
            <>
              {status === 'success' && itemCount > 0 && onReviewItems && (
                <Button
                  variant="ghost"
                  onClick={onReviewItems}
                  disabled={disabled}
                  aria-label="Review items"
                  className="inline-flex h-8 items-center justify-center px-2.5 py-0 text-xs"
                >
                  Review
                </Button>
              )}
              {canScan && (
                <Button
                  variant="primary"
                  onClick={onScan}
                  disabled={disabled}
                  aria-label={status === 'success' ? 'Scan again' : status === 'failed' ? 'Try again' : 'Scan attachment'}
                  className="inline-flex h-8 items-center justify-center px-2.5 py-0 text-xs font-bold"
                >
                  {status === 'success' ? 'Rescan' : status === 'failed' ? 'Retry' : 'Scan'}
                </Button>
              )}
            </>
          )}
          <IconButton
            aria-label={expanded ? 'Hide attachment details' : 'Show attachment details'}
            aria-expanded={expanded}
            aria-controls={detailsId}
            onClick={() => setExpanded((value) => !value)}
            icon={<IconChevronDown className={cx('h-4 w-4 transition-transform', expanded && 'rotate-180')} />}
            className="h-8 w-8"
          />
        </div>
      </div>

      {expanded && (
        <div id={detailsId} className="flex flex-col gap-2 border-t border-line px-3 pb-3 pt-2.5">
          <ul className="m-0 flex list-none flex-col gap-1.5 p-0" aria-label="Attached files">
            {attachments.map((attachment) => (
              <li key={attachment.id} className="flex items-center gap-2.5 rounded-lg bg-surface-sunken px-2.5 py-1.5">
                <button
                  type="button"
                  onClick={() => onPreview(attachment.id)}
                  disabled={!attachment.attachmentId}
                  aria-label={`Preview ${attachment.name}`}
                  className="flex min-w-0 flex-1 cursor-pointer items-center gap-2.5 rounded-md text-left disabled:cursor-default"
                >
                  {attachment.previewUrl ? (
                    <img src={attachment.previewUrl} alt="" className="h-8 w-8 flex-none rounded-md object-cover" />
                  ) : (
                    <span className="flex h-8 w-8 flex-none items-center justify-center rounded-md bg-surface text-muted">
                      <IconPaperclip className="h-3.5 w-3.5" />
                    </span>
                  )}
                  <span className="flex min-w-0 flex-1 flex-col">
                    <span className="truncate text-[12px] font-semibold text-ink">{attachment.name}</span>
                    <span
                      className={cx(
                        'truncate text-[10.5px]',
                        attachment.status === 'error' ? 'text-critical' : 'text-muted',
                      )}
                    >
                      {attachmentStatusLabel(attachment, scannedIds.has(attachment.attachmentId ?? ''))}
                    </span>
                  </span>
                </button>
                <IconButton
                  aria-label={`Remove attachment ${attachment.name}`}
                  onClick={() => onRemove(attachment.id)}
                  icon={<IconClose className="h-3.5 w-3.5" />}
                  disabled={disabled}
                  className="h-7 w-7"
                />
              </li>
            ))}
          </ul>

          {status === 'success' && summary && (
            <div className="flex flex-col gap-1.5 rounded-lg bg-surface-sunken px-3 py-2 text-[11px] text-muted">
              {summary.warnings.length > 0 && (
                <ul className="m-0 flex list-disc flex-col gap-1 pl-4">
                  {summary.warnings.map((warning) => (
                    <li key={warning}>{warning}</li>
                  ))}
                </ul>
              )}
              {summary.categoryHint && <p className="m-0">Category “{summary.categoryHint}” isn't in your list.</p>}
            </div>
          )}

          {canUndo && (
            <Button variant="link" onClick={onUndo} disabled={disabled} className="self-start p-0">
              Undo scan
            </Button>
          )}
        </div>
      )}
    </section>
  )
}

function attachmentStatusLabel(attachment: TransactionAttachmentDraft, scanned: boolean): string {
  switch (attachment.status) {
    case 'uploading':
      return 'Uploading…'
    case 'processing':
      return 'Reading…'
    case 'error':
      return 'Upload failed — remove and try again'
    default:
      if (!attachment.isImage) return 'Attached'
      return scanned ? 'Scanned' : 'Ready to scan'
  }
}
