import { useRef, type ChangeEvent, type RefObject } from 'react'
import { IconCamera, IconPaperclip } from '@/components/icons'
import { BottomSheet } from '@/components/ui/BottomSheet'
import { Button } from '@/components/ui/Button'

export interface TransactionAttachmentPickerProps {
  open: boolean
  onClose: () => void
  /** Invoked once per selected file, so a multi-select produces one call per file. */
  onFile: (file: File) => void
  /** Optional copy for pages whose attachment flow differs (e.g. the Add screen's scan hint). */
  description?: string
}

/**
 * Source picker for new attachments. Files are always uploaded first; extraction on the Add screen
 * remains an explicit action on the resulting attachment row, where the user can see exactly which
 * file will be read.
 */
export function TransactionAttachmentPicker({
  open,
  onClose,
  onFile,
  description = 'Take a photo or choose one or more files.',
}: TransactionAttachmentPickerProps) {
  const cameraInputRef = useRef<HTMLInputElement>(null)
  const fileInputRef = useRef<HTMLInputElement>(null)

  function openInput(ref: RefObject<HTMLInputElement | null>) {
    ref.current?.click()
    onClose()
  }

  function handleChange(event: ChangeEvent<HTMLInputElement>) {
    const files = Array.from(event.target.files ?? [])
    for (const file of files) onFile(file)
    event.target.value = ''
  }

  return (
    <BottomSheet
      open={open}
      onClose={onClose}
      title="Add attachment"
      description={description}
    >
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
      </div>
    </BottomSheet>
  )
}
