import { useEffect, useRef, useState, type ReactNode } from 'react'
import { OfflineScreen } from '@/components/OfflineScreen'
import { useOnlineStatus } from '@/hooks/useOnlineStatus'
import { isOnline } from '@/lib/pwa'

/**
 * Blocks the whole app — login route included — while the device has no connection. The routed
 * tree stays mounted but `inert`, so an in-progress draft and paused TanStack Query work survive
 * a network blip instead of being destroyed by an unmount. It is not a route: browser history is
 * unaffected by connectivity changes.
 */
export function OfflineGate({ children }: { children: ReactNode }) {
  const online = useOnlineStatus()
  // A Retry that found `navigator.onLine` already true lifts the block immediately, even if the
  // 'online' event hasn't landed in `useOnlineStatus` yet.
  const [retryOnline, setRetryOnline] = useState(false)
  const [stillOffline, setStillOffline] = useState(false)
  const [statusMessage, setStatusMessage] = useState('')
  const [prevOnline, setPrevOnline] = useState(online)
  const restoreFocusRef = useRef<HTMLElement | null>(null)
  const overlayRef = useRef<HTMLDivElement | null>(null)
  const wasBlockedRef = useRef(false)

  // Adjust the connectivity-derived state during render (the React-recommended pattern for a
  // value changing out from under the component) rather than in an effect.
  if (online !== prevOnline) {
    setPrevOnline(online)
    if (online) {
      setRetryOnline(false)
      setStillOffline(false)
      setStatusMessage('Back online')
    } else {
      setStillOffline(false)
      setStatusMessage('')
    }
  }

  const blocked = !online && !retryOnline

  // Focus management is the only side effect: move into the overlay when it appears, and restore
  // the previously focused element when it goes.
  useEffect(() => {
    const wasBlocked = wasBlockedRef.current
    wasBlockedRef.current = blocked

    if (blocked) {
      // Capture the pre-block focus only on the transition — StrictMode in dev runs this effect
      // twice on mount, and the second run would otherwise capture the Retry button itself.
      if (!wasBlocked) {
        restoreFocusRef.current = document.activeElement instanceof HTMLElement ? document.activeElement : null
      }
      overlayRef.current?.querySelector<HTMLButtonElement>('button')?.focus()
      return
    }

    if (wasBlocked) {
      const previous = restoreFocusRef.current
      if (previous && document.contains(previous)) previous.focus()
      restoreFocusRef.current = null
    }
  }, [blocked])

  function handleRetry() {
    if (isOnline()) {
      setRetryOnline(true)
      setStillOffline(false)
      setStatusMessage('Back online')
      return
    }
    setStillOffline(true)
    setStatusMessage('Still offline — check your connection.')
  }

  return (
    <>
      <div inert={blocked}>{children}</div>
      {blocked && (
        <div ref={overlayRef}>
          <OfflineScreen onRetry={handleRetry} stillOffline={stillOffline} />
        </div>
      )}
      <p role="status" className="sr-only">
        {statusMessage}
      </p>
    </>
  )
}
