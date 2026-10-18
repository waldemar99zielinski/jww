import { useRef, useState, type ChangeEvent, type RefObject } from 'react'
import { IconCamera, IconPaperclip, IconSparkle } from '@/components/icons'
import { BottomSheet } from '@/components/ui/BottomSheet'
import { Button } from '@/components/ui/Button'
import type { TransactionAttachmentFlow } from '../types'

export interface TransactionAttachmentPickerProps {
  open: boolean
  onClose: () => void
  /** Invoked once per selected file, so a multi-select produces one call per file. */
  onFile: (file: File, flow: TransactionAttachmentFlow) => void
}

function modeRowClass() {
  return 'flex items-center gap-3 rounded-xl border border-line bg-surface-sunken px-4 py-3.5 text-left'
}

/**
 * Two-step sheet: first pick what the files are for (attach only vs. import & prefill), then pick
 * the source. Both sources allow selecting several files at once; each one is forwarded to `onFile`
 * and closes the sheet. The page owns the uploads and, for `prefill`, the read.
 */
export function TransactionAttachmentPicker({ open, onClose, onFile }: TransactionAttachmentPickerProps) {
  const cameraInputRef = useRef<HTMLInputElement>(null)
  const fileInputRef = useRef<HTMLInputElement>(null)
  const [step, setStep] = useState<'mode' | 'source'>('mode')
  const [flow, setFlow] = useState<TransactionAttachmentFlow>('attach')
  const [prevOpen, setPrevOpen] = useState(open)

  // A fresh open always starts at the mode choice; reset during render (the BottomSheet pattern)
  // so a reopen never flashes the previous step.
  if (open !== prevOpen) {
    setPrevOpen(open)
    if (open) setStep('mode')
  }

  function openInput(ref: RefObject<HTMLInputElement | null>) {
    ref.current?.click()
    onClose()
  }

  function handleChange(event: ChangeEvent<HTMLInputElement>) {
    const files = Array.from(event.target.files ?? [])
    for (const file of files) onFile(file, flow)
    event.target.value = ''
  }

  function chooseMode(next: TransactionAttachmentFlow) {
    setFlow(next)
    setStep('source')
  }

  const choosingMode = step === 'mode'

  return (
    <BottomSheet
      open={open}
      onClose={onClose}
      title={choosingMode ? 'Add attachment' : flow === 'prefill' ? 'Import & prefill' : 'Add attachment'}
      description={
        choosingMode
          ? 'Attach files, or read one to fill the form for you.'
          : 'Take a photo or choose one or more files — any type.'
      }
    >
      {choosingMode ? (
        <div className="grid gap-2 pb-1">
          <Button variant="custom" onClick={() => chooseMode('prefill')} className={modeRowClass()}>
            <span className="flex h-9 w-9 flex-none items-center justify-center rounded-lg bg-surface text-accent">
              <IconSparkle className="h-4 w-4" />
            </span>
            <span className="flex min-w-0 flex-col">
              <span className="text-sm font-semibold text-ink">Import &amp; prefill</span>
              <span className="text-xs text-muted">Reads a file and fills the transaction</span>
            </span>
          </Button>
          <Button variant="custom" onClick={() => chooseMode('attach')} className={modeRowClass()}>
            <span className="flex h-9 w-9 flex-none items-center justify-center rounded-lg bg-surface text-ink-soft">
              <IconPaperclip className="h-4 w-4" />
            </span>
            <span className="flex min-w-0 flex-col">
              <span className="text-sm font-semibold text-ink">Just attach</span>
              <span className="text-xs text-muted">Adds files without reading them</span>
            </span>
          </Button>
        </div>
      ) : (
        <div className="grid gap-2 pb-1">
          <input
            ref={cameraInputRef}
            type="file"
            accept="image/*"
            capture="environment"
            multiple
            className="hidden"
            onChange={handleChange}
          />
          <input ref={fileInputRef} type="file" multiple className="hidden" onChange={handleChange} />
          <Button
            variant="secondary"
            onClick={() => openInput(cameraInputRef)}
            className="flex items-center justify-center gap-2 py-3.5 text-sm font-semibold"
          >
            <IconCamera className="h-4 w-4" />
            Take photo
          </Button>
          <Button
            variant="secondary"
            onClick={() => openInput(fileInputRef)}
            className="flex items-center justify-center gap-2 py-3.5 text-sm font-semibold"
          >
            <IconPaperclip className="h-4 w-4" />
            Choose files
          </Button>
          <Button variant="ghost" onClick={() => setStep('mode')} className="py-2 text-xs">
            ← Back
          </Button>
        </div>
      )}
    </BottomSheet>
  )
}
