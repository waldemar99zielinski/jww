import { useState } from 'react'
import { IconDownload, IconImage, IconTrash } from '@/components/icons'
import { BottomSheet } from '@/components/ui/BottomSheet'
import { Button } from '@/components/ui/Button'
import { formatBytes } from '@/lib/formatBytes'
import type { TransactionAttachmentItem } from '../types'

export interface AttachmentActionsProps {
  open: boolean
  attachment: TransactionAttachmentItem
  canPreview: boolean
  isDownloading: boolean
  isDeleting: boolean
  deleteError: boolean
  onPreview: () => void
  onDownload: () => void
  onDelete: () => void
  onClose: () => void
}

function rowClass(danger = false) {
  return `flex items-center gap-3 rounded-xl border border-line bg-surface-sunken px-4 py-3.5 text-left disabled:opacity-40 ${
    danger ? 'text-critical' : 'text-ink'
  }`
}

/**
 * Bottom drawer for one attachment, opened by the row's actions button: preview it big, download
 * it, or delete it. Delete is a two-step confirm inside the sheet, since removing a linked
 * attachment is irreversible.
 */
export function AttachmentActions({
  open,
  attachment,
  canPreview,
  isDownloading,
  isDeleting,
  deleteError,
  onPreview,
  onDownload,
  onDelete,
  onClose,
}: AttachmentActionsProps) {
  const [confirmingDelete, setConfirmingDelete] = useState(false)
  const [prevOpen, setPrevOpen] = useState(open)

  // A fresh open always starts on the actions list; reset during render (the BottomSheet pattern).
  if (open !== prevOpen) {
    setPrevOpen(open)
    if (open) setConfirmingDelete(false)
  }

  return (
    <BottomSheet
      open={open}
      onClose={onClose}
      title={attachment.filename}
      description={`${formatBytes(attachment.byteSize)} · ${attachment.mediaType}`}
    >
      {confirmingDelete ? (
        <div className="grid gap-2 pb-1">
          <p className="m-0 text-sm text-ink-soft">Delete this attachment? This can't be undone.</p>
          {deleteError && <p className="m-0 text-[11.5px] text-critical">Couldn't delete — try again.</p>}
          <Button
            variant="custom"
            onClick={onDelete}
            disabled={isDeleting}
            className={`${rowClass(true)} justify-center font-semibold`}
          >
            {isDeleting ? 'Deleting…' : 'Delete attachment'}
          </Button>
          <Button
            variant="ghost"
            onClick={() => setConfirmingDelete(false)}
            disabled={isDeleting}
            className="py-2.5 text-sm"
          >
            Cancel
          </Button>
        </div>
      ) : (
        <div className="grid gap-2 pb-1">
          {canPreview && (
            <Button variant="custom" onClick={onPreview} className={rowClass()}>
              <span className="flex h-9 w-9 flex-none items-center justify-center rounded-lg bg-surface text-accent">
                <IconImage className="h-4 w-4" />
              </span>
              <span className="text-sm font-semibold">Preview</span>
            </Button>
          )}
          <Button variant="custom" onClick={onDownload} disabled={isDownloading} className={rowClass()}>
            <span className="flex h-9 w-9 flex-none items-center justify-center rounded-lg bg-surface text-ink-soft">
              <IconDownload className="h-4 w-4" />
            </span>
            <span className="text-sm font-semibold">{isDownloading ? 'Downloading…' : 'Download'}</span>
          </Button>
          <Button variant="custom" onClick={() => setConfirmingDelete(true)} className={rowClass(true)}>
            <span className="flex h-9 w-9 flex-none items-center justify-center rounded-lg bg-surface">
              <IconTrash className="h-4 w-4" />
            </span>
            <span className="text-sm font-semibold">Delete</span>
          </Button>
        </div>
      )}
    </BottomSheet>
  )
}
