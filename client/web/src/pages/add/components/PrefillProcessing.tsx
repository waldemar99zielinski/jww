import { Loader } from '@/components/ui/Loader'
import { Button } from '@/components/ui/Button'

export interface PrefillProcessingProps {
  fileName: string
  onCancel: () => void
}

/** Full-screen wait shown while a picked file is being read for prefill. */
export function PrefillProcessing({ fileName, onCancel }: PrefillProcessingProps) {
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-paper/95 px-6">
      <div className="flex w-full max-w-sm flex-col items-center gap-5 rounded-2xl border border-line bg-surface p-8 text-center">
        <Loader />
        <div className="flex flex-col gap-1">
          <p className="m-0 text-sm font-semibold text-ink">Reading {fileName}…</p>
          <p className="m-0 text-xs text-muted">
            Looking for the amount, merchant and date. Larger files take longer.
          </p>
        </div>
        <Button variant="secondary" onClick={onCancel} className="w-full py-3.5 text-sm font-semibold">
          Cancel
        </Button>
      </div>
    </div>
  )
}
