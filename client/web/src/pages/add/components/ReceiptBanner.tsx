import { IconClose, IconPaperclip } from '@/components/icons'
import { IconButton } from '@/components/ui/IconButton'
import type { ReceiptDraft } from '../types'

export interface ReceiptBannerProps {
  receipt: ReceiptDraft
  onRemove: () => void
}

export function ReceiptBanner({ receipt, onRemove }: ReceiptBannerProps) {
  const parsing = receipt.status === 'parsing'
  const showImage = receipt.isImage && receipt.previewUrl

  return (
    <div className="flex items-center gap-3 rounded-2xl border border-line bg-surface-sunken px-3.5 py-3">
      {showImage ? (
        <img src={receipt.previewUrl} alt="" className="h-11 w-11 flex-none rounded-xl object-cover" />
      ) : (
        <span className="flex h-11 w-11 flex-none items-center justify-center rounded-xl bg-surface text-muted">
          <IconPaperclip className="h-4 w-4" />
        </span>
      )}
      <div className="flex min-w-0 flex-1 flex-col">
        <span className="truncate text-[13px] font-semibold text-ink">{parsing ? 'Reading file…' : 'Attachment ready'}</span>
        <span className="truncate text-xs text-muted">
          {parsing ? 'Detecting amount and details' : `${receipt.name} — review the details`}
        </span>
      </div>
      <IconButton aria-label="Remove attachment" onClick={onRemove} icon={<IconClose className="h-4 w-4" />} className="h-8 w-8" />
    </div>
  )
}
