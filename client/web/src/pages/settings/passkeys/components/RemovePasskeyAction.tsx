import { BottomSheet } from '@/components/ui/BottomSheet'
import { Button } from '@/components/ui/Button'

export interface RemovePasskeyActionProps {
  open: boolean
  credentialName: string
  isRemoving: boolean
  error: boolean
  onConfirm: () => void
  onClose: () => void
}

/** Confirmation sheet for removing an enrolled Face ID/Touch ID credential. */
export function RemovePasskeyAction({
  open,
  credentialName,
  isRemoving,
  error,
  onConfirm,
  onClose,
}: RemovePasskeyActionProps) {
  return (
    <BottomSheet open={open} onClose={onClose} title="Remove this device?" description={credentialName}>
      <div className="grid gap-2 pb-1">
        <p className="m-0 text-sm text-ink-soft">
          You'll no longer be able to sign in with Face ID/Touch ID on this device. This can't be undone.
        </p>
        {error && <p className="m-0 text-[11.5px] text-critical">Couldn't remove the device — try again.</p>}
        <Button
          variant="custom"
          onClick={onConfirm}
          disabled={isRemoving}
          className="flex items-center justify-center rounded-xl border border-line bg-surface-sunken px-4 py-3.5 text-sm font-semibold text-critical disabled:opacity-40"
        >
          {isRemoving ? 'Removing…' : 'Remove permanently'}
        </Button>
        <Button variant="ghost" onClick={onClose} disabled={isRemoving} className="py-2.5 text-sm">
          Cancel
        </Button>
      </div>
    </BottomSheet>
  )
}
