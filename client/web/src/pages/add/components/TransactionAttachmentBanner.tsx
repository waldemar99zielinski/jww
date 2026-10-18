import { IconClose, IconPaperclip, IconSpinner, IconSparkle } from '@/components/icons'
import { IconButton } from '@/components/ui/IconButton'
import type { TransactionAttachmentDraft } from '../types'

export interface TransactionAttachmentBannerProps {
  attachment: TransactionAttachmentDraft
  onRemove: () => void
}

export function TransactionAttachmentBanner({ attachment, onRemove }: TransactionAttachmentBannerProps) {
  const showImage = attachment.isImage && attachment.previewUrl
  const prefilled = attachment.prefilledCount ?? 0
  // Uploading or reading: not yet actionable, so the remove control is replaced by a spinner.
  const busy = attachment.status === 'uploading' || attachment.status === 'processing'

  const title =
    attachment.status === 'uploading'
      ? 'Uploading…'
      : attachment.status === 'processing'
        ? 'Reading…'
        : attachment.status === 'error'
          ? 'Upload failed'
          : attachment.readFailed
            ? "Couldn't read the file"
            : prefilled > 0
              ? `${prefilled} ${prefilled === 1 ? 'field' : 'fields'} filled from ${attachment.name}`
              : 'Attachment ready'

  const subtitle =
    attachment.status === 'uploading'
      ? 'Sending your file to the server'
      : attachment.status === 'processing'
        ? 'Looking for the amount, merchant and date'
        : attachment.status === 'error'
          ? "Couldn't attach this file — remove it and try again"
          : attachment.readFailed
            ? `${attachment.name} — attached when you save`
            : prefilled > 0
              ? 'Check the form, then save'
              : `${attachment.name} — attached when you save`

  return (
    <div className="flex items-center gap-3 rounded-2xl border border-line bg-surface-sunken px-3.5 py-3">
      {showImage ? (
        <img src={attachment.previewUrl} alt="" className="h-11 w-11 flex-none rounded-xl object-cover" />
      ) : (
        <span className="flex h-11 w-11 flex-none items-center justify-center rounded-xl bg-surface text-muted">
          {prefilled > 0 ? <IconSparkle className="h-4 w-4" /> : <IconPaperclip className="h-4 w-4" />}
        </span>
      )}
      <div className="flex min-w-0 flex-1 flex-col">
        <span className={`truncate text-[13px] font-semibold ${attachment.status === 'error' ? 'text-critical' : 'text-ink'}`}>
          {title}
        </span>
        <span className="truncate text-xs text-muted">{subtitle}</span>
      </div>
      {busy ? (
        <span
          role="status"
          aria-label="Uploading attachment"
          className="flex h-8 w-8 flex-none items-center justify-center text-muted"
        >
          <IconSpinner className="h-4 w-4 animate-spin" />
        </span>
      ) : (
        <IconButton
          aria-label={`Remove attachment ${attachment.name}`}
          onClick={onRemove}
          icon={<IconClose className="h-4 w-4" />}
          className="h-8 w-8"
        />
      )}
    </div>
  )
}
