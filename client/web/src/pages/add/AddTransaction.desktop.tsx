import { IconClose, IconPaperclip } from '@/components/icons'
import { TransactionAttachmentPicker } from '@/components/TransactionAttachmentPicker'
import { IconButton } from '@/components/ui/IconButton'
import { AddTransactionForm } from './components/AddTransactionForm'
import { AttachmentBar } from './components/AttachmentBar'
import type { AddTransactionViewProps } from './types'

export default function AddTransactionDesktop({
  attachments,
  extractionStatus,
  extractionSummary,
  canScanAttachments,
  canUndoExtraction,
  attachmentPickerOpen,
  onOpenAttachment,
  onCloseAttachment,
  onPickAttachment,
  onRemoveAttachment,
  onPreviewAttachment,
  onScanAttachments,
  onCancelExtraction,
  onUndoExtraction,
  onReviewSubTransactions,
  saving,
  onClose,
  ...formProps
}: AddTransactionViewProps) {
  return (
    <div className="mx-auto flex max-w-2xl flex-col gap-6 rounded-2xl border border-line bg-surface p-8">
      <div className="flex items-center justify-between">
        <h1 className="m-0 text-2xl font-bold">Add transaction</h1>
        <div className="flex items-center gap-2">
          <IconButton
            aria-label="Add attachment"
            onClick={onOpenAttachment}
            icon={<IconPaperclip className="h-4 w-4" />}
            className="h-9 w-9"
          />
          <IconButton aria-label="Close" onClick={onClose} icon={<IconClose className="h-4 w-4" />} className="h-9 w-9" />
        </div>
      </div>

      <AttachmentBar
        attachments={attachments}
        status={extractionStatus}
        summary={extractionSummary}
        canScan={canScanAttachments}
        canUndo={canUndoExtraction}
        disabled={saving}
        onRemove={onRemoveAttachment}
        onPreview={onPreviewAttachment}
        onScan={onScanAttachments}
        onCancel={onCancelExtraction}
        onUndo={onUndoExtraction}
        onReviewItems={onReviewSubTransactions}
      />

      <AddTransactionForm layout="desktop" {...formProps} extractionSummary={extractionSummary} saving={saving} />

      <TransactionAttachmentPicker
        open={attachmentPickerOpen}
        onClose={onCloseAttachment}
        onFile={onPickAttachment}
        description="Take a photo or choose one or more files. You can scan the uploaded image into the form."
      />
    </div>
  )
}
