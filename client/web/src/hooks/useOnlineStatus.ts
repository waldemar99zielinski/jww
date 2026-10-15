import { useSyncExternalStore } from 'react'
import { isOnline } from '@/lib/pwa'

function subscribe(onChange: () => void): () => void {
  window.addEventListener('online', onChange)
  window.addEventListener('offline', onChange)
  return () => {
    window.removeEventListener('online', onChange)
    window.removeEventListener('offline', onChange)
  }
}

/** Assume online until we know otherwise — a false negative would flash the offline screen. */
function getServerSnapshot(): boolean {
  return true
}

/** Mirrors hooks/useViewport.ts: a `useSyncExternalStore` wrapper over a browser flag. */
export function useOnlineStatus(): boolean {
  return useSyncExternalStore(subscribe, isOnline, getServerSnapshot)
}
