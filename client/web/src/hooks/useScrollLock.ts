import { useEffect } from 'react'

let lockCount = 0
let saved: Array<{ element: HTMLElement; overflow: string }> = []

/**
 * Every element that needs freezing: the document, plus any element opted in with a
 * `data-scroll-lock` attribute. The desktop app shell scrolls an inner `<main>` rather than the
 * body, so locking `document.body` alone would leave the page behind an overlay scrollable.
 */
function lockTargets(): HTMLElement[] {
  const targets: HTMLElement[] = [document.documentElement, document.body]
  document.querySelectorAll<HTMLElement>('[data-scroll-lock]').forEach((element) => targets.push(element))
  return targets
}

/**
 * Freezes background scrolling while an overlay is open, restoring the previous overflow on close.
 * Reference-counted, so overlapping overlays don't unlock each other early.
 */
export function useScrollLock(active = true) {
  useEffect(() => {
    if (!active) return
    if (lockCount === 0) {
      saved = lockTargets().map((element) => ({ element, overflow: element.style.overflow }))
      for (const { element } of saved) element.style.overflow = 'hidden'
    }
    lockCount += 1
    return () => {
      lockCount -= 1
      if (lockCount === 0) {
        for (const { element, overflow } of saved) element.style.overflow = overflow
        saved = []
      }
    }
  }, [active])
}
