import { useCallback, useEffect, useRef, useState, type ReactNode, type TouchEvent } from 'react'

/** Swipes starting this close to the left edge belong to the OS back gesture — leave them alone. */
const EDGE_GUARD_PX = 24
/** Don't commit to a horizontal drag until it clearly beats the vertical axis (so scroll wins). */
const DIRECTION_LOCK_PX = 12
/** A drag must cross this fraction of the viewport width (capped) on release to dismiss. */
const CLOSE_THRESHOLD_RATIO = 0.28
const CLOSE_THRESHOLD_MAX_PX = 140
/** A hair longer than the 200ms exit transition, so it finishes before we navigate away. */
const EXIT_MS = 220

interface SwipeToCloseProps {
  onClose: () => void
  /**
   * Receives an animated `close()` so in-content controls (the header's Close button) share the
   * same exit transition instead of navigating instantly.
   */
  children: (close: () => void) => ReactNode
}

/**
 * Lets the user drag a mobile screen to the right to dismiss it, and animates the exit when a
 * control calls the provided `close`.
 *
 * This **supplements** the OS edge-swipe back gesture, it does not replace it: web content cannot
 * intercept a system gesture, so we simply ignore touches near the left edge and claim only
 * gestures that are clearly horizontal (vertical ones are abandoned so the page keeps scrolling).
 */
export function SwipeToClose({ onClose, children }: SwipeToCloseProps) {
  const [offsetX, setOffsetX] = useState(0)
  const [dragging, setDragging] = useState(false)
  const [closing, setClosing] = useState(false)

  const gesture = useRef<{ startX: number; startY: number; horizontal: boolean } | null>(null)
  const offsetRef = useRef(0)
  const onCloseRef = useRef(onClose)
  useEffect(() => {
    onCloseRef.current = onClose
  })

  // The exit animation is played here; only once it has finished do we actually navigate away.
  // `onClose` is read through a ref so this effect never restarts on a re-render.
  useEffect(() => {
    if (!closing) return
    const timer = window.setTimeout(() => onCloseRef.current(), EXIT_MS)
    return () => window.clearTimeout(timer)
  }, [closing])

  const close = useCallback(() => setClosing(true), [])

  function handleTouchStart(event: TouchEvent<HTMLDivElement>) {
    if (closing) return
    const touch = event.touches[0]
    if (!touch || touch.clientX < EDGE_GUARD_PX) return
    gesture.current = { startX: touch.clientX, startY: touch.clientY, horizontal: false }
    offsetRef.current = 0
  }

  function handleTouchMove(event: TouchEvent<HTMLDivElement>) {
    const current = gesture.current
    if (!current || closing) return
    const touch = event.touches[0]
    if (!touch) return
    const dx = touch.clientX - current.startX
    const dy = touch.clientY - current.startY

    if (!current.horizontal) {
      if (Math.abs(dx) < DIRECTION_LOCK_PX && Math.abs(dy) < DIRECTION_LOCK_PX) return
      // Vertical intent wins — abandon so the page keeps scrolling.
      if (Math.abs(dy) >= Math.abs(dx)) {
        gesture.current = null
        return
      }
      current.horizontal = true
      setDragging(true)
    }

    const next = Math.max(0, dx)
    offsetRef.current = next
    setOffsetX(next)
  }

  function handleTouchEnd() {
    const current = gesture.current
    gesture.current = null
    setDragging(false)
    if (!current?.horizontal) {
      offsetRef.current = 0
      setOffsetX(0)
      return
    }
    const threshold = Math.min(CLOSE_THRESHOLD_MAX_PX, window.innerWidth * CLOSE_THRESHOLD_RATIO)
    if (offsetRef.current > threshold) {
      close()
      return
    }
    offsetRef.current = 0
    setOffsetX(0)
  }

  const transform = closing ? 'translateX(100%)' : dragging ? `translateX(${offsetX}px)` : 'translateX(0px)'

  return (
    <div
      className={`[touch-action:pan-y] ${dragging ? 'transition-none' : 'transition-transform duration-200 ease-out'}`}
      style={{ transform }}
      onTouchStart={handleTouchStart}
      onTouchMove={handleTouchMove}
      onTouchEnd={handleTouchEnd}
      onTouchCancel={handleTouchEnd}
    >
      <div className="screen-enter-right">{children(close)}</div>
    </div>
  )
}
