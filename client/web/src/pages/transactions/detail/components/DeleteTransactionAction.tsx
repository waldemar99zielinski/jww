import { BottomSheet } from '@/components/ui/BottomSheet'
import { Button } from '@/components/ui/Button'

export interface DeleteTransactionActionProps {
  open: boolean
  transactionName: string
  isDeleting: boolean
  error: boolean
  onConfirm: () => void
  onClose: () => void
}

/**
 * Confirmation sheet for the destructive transaction delete. The copy spells out the cascade so
 * the user knows removing the transaction also removes its line items and attachments.
 */
export function DeleteTransactionAction({
  open,
  transactionName,
  isDeleting,
  error,
  onConfirm,
  onClose,
}: DeleteTransactionActionProps) {
  return (
    <BottomSheet open={open} onClose={onClose} title="Remove transaction?" description={transactionName}>
      <div className="grid gap-2 pb-1">
        <p className="m-0 text-sm text-ink-soft">
          This removes the transaction, all of its items, and all of its attachments. This can't be undone.
        </p>
        {error && <p className="m-0 text-[11.5px] text-critical">Couldn't remove the transaction — try again.</p>}
        <Button
          variant="custom"
          onClick={onConfirm}
          disabled={isDeleting}
          className="flex items-center justify-center rounded-xl border border-line bg-surface-sunken px-4 py-3.5 text-sm font-semibold text-critical disabled:opacity-40"
        >
          {isDeleting ? 'Removing…' : 'Remove permanently'}
        </Button>
        <Button variant="ghost" onClick={onClose} disabled={isDeleting} className="py-2.5 text-sm">
          Cancel
        </Button>
      </div>
    </BottomSheet>
  )
}
