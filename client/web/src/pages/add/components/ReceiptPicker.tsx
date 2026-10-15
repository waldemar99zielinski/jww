import { useRef, type ChangeEvent, type RefObject } from 'react'
import { IconCamera, IconPaperclip } from '@/components/icons'
import { BottomSheet } from '@/components/ui/BottomSheet'
import { Button } from '@/components/ui/Button'

export interface ReceiptPickerProps {
  open: boolean
  onClose: () => void
  onFile: (file: File) => void
}

export function ReceiptPicker({ open, onClose, onFile }: ReceiptPickerProps) {
  const cameraInputRef = useRef<HTMLInputElement>(null)
  const fileInputRef = useRef<HTMLInputElement>(null)

  function openInput(ref: RefObject<HTMLInputElement | null>) {
    ref.current?.click()
    onClose()
  }

  function handleChange(event: ChangeEvent<HTMLInputElement>) {
    const file = event.target.files?.[0]
    if (file) onFile(file)
    event.target.value = ''
  }

  return (
    <BottomSheet
      open={open}
      onClose={onClose}
      title="Add attachment"
      description="Attach a photo, PDF, or any other file to this transaction."
    >
      <div className="grid gap-2 pb-1">
        <input
          ref={cameraInputRef}
          type="file"
          accept="image/*"
          capture="environment"
          className="hidden"
          onChange={handleChange}
        />
        <input ref={fileInputRef} type="file" className="hidden" onChange={handleChange} />
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
          Choose a file
        </Button>
      </div>
    </BottomSheet>
  )
}
