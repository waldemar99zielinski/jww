import { useSyncExternalStore } from 'react'
import { getViewport, type Viewport } from '@/lib/viewport'

function subscribe(onChange: () => void): () => void {
  window.addEventListener('resize', onChange)
  return () => window.removeEventListener('resize', onChange)
}

/** Mobile is the safe default before the first resize-driven read runs. */
function getServerSnapshot(): Viewport {
  return 'mobile'
}

export function useViewport(): Viewport {
  return useSyncExternalStore(subscribe, getViewport, getServerSnapshot)
}
