import { useEffect, useId, type ReactNode } from 'react'
import { IconClose } from '@/components/icons'
import { useScrollLock } from '@/hooks/useScrollLock'
import { IconButton } from './IconButton'

export interface ModalProps {
  open: boolean
  onClose: () => void
  title: ReactNode
  description?: ReactNode
  footer?: ReactNode
  children: ReactNode
}

/**
 * Centered dialog for a short, focused task (e.g. editing one line item). Unlike `BottomSheet`, it
 * is not a docked workspace: it unmounts when closed, freezes the page, and closes on
 * Escape/backdrop. Callers put `autoFocus` on the first field they want focused.
 */
export function Modal({ open, onClose, title, description, footer, children }: ModalProps) {
  const titleId = useId()

  useScrollLock(open)

  useEffect(() => {
    if (!open) return
    function handleKeyDown(event: KeyboardEvent) {
      if (event.key === 'Escape') onClose()
    }
    document.addEventListener('keydown', handleKeyDown)
    return () => document.removeEventListener('keydown', handleKeyDown)
  }, [onClose, open])

  if (!open) return null

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
      <div className="absolute inset-0 bg-black/60" onClick={onClose} role="presentation" />
      <div
        role="dialog"
        aria-modal="true"
        aria-labelledby={titleId}
        tabIndex={-1}
        className="relative flex max-h-[90dvh] w-full max-w-md flex-col overflow-hidden rounded-2xl border border-line bg-surface shadow-[var(--shadow-md)] outline-none"
      >
        <header className="flex items-start justify-between gap-3 border-b border-line px-5 py-4">
          <div className="min-w-0">
            <p id={titleId} className="m-0 text-base font-bold text-ink">
              {title}
            </p>
            {description && <p className="m-0 mt-0.5 text-xs text-muted">{description}</p>}
          </div>
          <IconButton aria-label="Close" onClick={onClose} icon={<IconClose className="h-4 w-4" />} className="h-8 w-8" />
        </header>
        <div className="min-h-0 flex-1 overflow-y-auto px-5 py-4">{children}</div>
        {footer && <div className="shrink-0 border-t border-line px-5 py-3">{footer}</div>}
      </div>
    </div>
  )
}
