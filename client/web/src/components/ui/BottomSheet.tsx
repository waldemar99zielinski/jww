import { useCallback, useEffect, useRef, useState, type ReactNode, type TouchEvent } from 'react'
import { IconClose } from '@/components/icons'
import { useScrollLock } from '@/hooks/useScrollLock'
import { IconButton } from './IconButton'

export interface BottomSheetProps {
  open: boolean
  /** Return `false` to veto closing (e.g. validation failed); the sheet will spring back up. */
  onClose: () => boolean | void
  title: ReactNode
  description?: ReactNode
  footer?: ReactNode
  children: ReactNode
}

const DISMISS_DISTANCE = 100
const SLIDE_TRANSITION = 'transition-transform duration-300 ease-[cubic-bezier(0.32,0.72,0,1)]'

export function BottomSheet({ open, onClose, title, description, footer, children }: BottomSheetProps) {
  const [entered, setEntered] = useState(open)
  const [dragging, setDragging] = useState(false)
  const [dragOffset, setDragOffset] = useState(0)
  const [prevOpen, setPrevOpen] = useState(open)
  const dragStartY = useRef<number | null>(null)

  // Keep the sheet mounted and animate via transform, so both open and close are smooth.
  if (open !== prevOpen) {
    setPrevOpen(open)
    setEntered(open)
    if (!open) {
      setDragging(false)
      setDragOffset(0)
    }
  }

  const dismiss = useCallback(() => {
    const didClose = onClose()
    setDragging(false)
    setDragOffset(0)
    if (didClose === false) setEntered(true)
  }, [onClose])

  useEffect(() => {
    if (!open) return
    function handleKeyDown(event: KeyboardEvent) {
      if (event.key === 'Escape') dismiss()
    }
    document.addEventListener('keydown', handleKeyDown)
    return () => document.removeEventListener('keydown', handleKeyDown)
  }, [dismiss, open])

  // Freeze the page behind the sheet — including the desktop shell's scrolling <main>.
  useScrollLock(open)

  function handleTouchStart(event: TouchEvent) {
    dragStartY.current = event.touches[0].clientY
    setDragging(true)
  }

  function handleTouchMove(event: TouchEvent) {
    if (dragStartY.current === null) return
    const delta = event.touches[0].clientY - dragStartY.current
    setDragOffset(Math.max(0, delta))
  }

  function handleTouchEnd() {
    dragStartY.current = null
    if (dragOffset > DISMISS_DISTANCE) {
      dismiss()
      return
    }
    setDragging(false)
    setDragOffset(0)
  }

  const translateY = dragging ? `${dragOffset}px` : entered ? '0px' : '100%'

  return (
    <div
      className={`fixed inset-0 z-50 flex flex-col justify-end ${entered ? 'pointer-events-auto' : 'pointer-events-none'}`}
      aria-hidden={!entered}
      inert={!entered}
    >
      <div
        className={`absolute inset-0 bg-black/40 transition-opacity duration-300 ease-out ${entered ? 'opacity-100' : 'opacity-0'}`}
        onClick={dismiss}
        role="presentation"
      />
      <div
        className={`sheet-panel relative flex w-full flex-col overflow-hidden rounded-t-3xl border border-line bg-surface shadow-[var(--shadow-md)] ${
          dragging ? 'transition-none will-change-transform' : SLIDE_TRANSITION
        }`}
        style={{ transform: `translateY(${translateY})` }}
      >
        <div
          className="shrink-0 touch-none"
          onTouchStart={handleTouchStart}
          onTouchMove={handleTouchMove}
          onTouchEnd={handleTouchEnd}
          onTouchCancel={handleTouchEnd}
        >
          <div className="flex shrink-0 justify-center pt-2.5">
            <span className="h-1 w-10 rounded-full bg-line" />
          </div>
          <div className="flex shrink-0 items-start justify-between gap-3 px-5 pb-3 pt-2">
            <div>
              <p className="m-0 text-base font-bold text-ink">{title}</p>
              {description && <p className="m-0 mt-0.5 text-xs text-muted">{description}</p>}
            </div>
            <IconButton aria-label="Close" onClick={dismiss} icon={<IconClose className="h-4 w-4" />} className="h-8 w-8" />
          </div>
        </div>
        <div className="min-h-0 flex-1 overflow-y-auto overscroll-contain px-5 pb-4 [touch-action:pan-y]">
          {children}
        </div>
        {footer && (
          <div className="shrink-0 border-t border-line bg-surface px-5 pb-[calc(env(safe-area-inset-bottom)+1rem)] pt-3">
            {footer}
          </div>
        )}
      </div>
    </div>
  )
}
