import { useCallback, useEffect, useRef, useState } from 'react'
import {
  detectPlatform,
  isInAppWebView,
  resolveInstallUiState,
  type InstallEnvironment,
  type InstallUiState,
} from '@/lib/installEnvironment'
import { isStandalone } from '@/lib/pwa'

export interface InstallPrompt {
  state: InstallUiState
  startInstall: () => Promise<'accepted' | 'dismissed' | 'unavailable'>
}

/** The slice of BeforeInstallPromptEvent we use — it has no DOM lib type in current TypeScript. */
interface BeforeInstallPromptEvent extends Event {
  prompt: () => Promise<{ outcome: 'accepted' | 'dismissed' }>
}

function readEnvironment(promptEvent: BeforeInstallPromptEvent | null): InstallEnvironment {
  const userAgent = navigator.userAgent
  return {
    platform: detectPlatform(userAgent),
    isStandalone: isStandalone(),
    hasPrompt: promptEvent !== null,
    isWebView: isInAppWebView(userAgent),
  }
}

/**
 * Captures Chromium's `beforeinstallprompt`, tracks `appinstalled` and standalone-mode changes,
 * and exposes the resolved row state plus the imperative "open the browser dialog" action.
 * Nothing is persisted: the state recomputes from the environment on every relevant event.
 */
export function useInstallPrompt(): InstallPrompt {
  const promptEventRef = useRef<BeforeInstallPromptEvent | null>(null)
  const installedRef = useRef(false)
  const [state, setState] = useState<InstallUiState>({ kind: 'hidden' })

  const computeState = useCallback((): InstallUiState => {
    if (installedRef.current) return { kind: 'hidden' }
    return resolveInstallUiState(readEnvironment(promptEventRef.current))
  }, [])

  useEffect(() => {
    const recompute = () => setState(computeState())

    function handleBeforeInstallPrompt(event: Event) {
      event.preventDefault()
      promptEventRef.current = event as BeforeInstallPromptEvent
      recompute()
    }
    function handleAppInstalled() {
      installedRef.current = true
      promptEventRef.current = null
      recompute()
    }

    const standaloneQuery = window.matchMedia('(display-mode: standalone)')
    window.addEventListener('beforeinstallprompt', handleBeforeInstallPrompt)
    window.addEventListener('appinstalled', handleAppInstalled)
    standaloneQuery.addEventListener('change', recompute)

    recompute()

    return () => {
      window.removeEventListener('beforeinstallprompt', handleBeforeInstallPrompt)
      window.removeEventListener('appinstalled', handleAppInstalled)
      standaloneQuery.removeEventListener('change', recompute)
    }
  }, [computeState])

  const startInstall = useCallback(async (): Promise<'accepted' | 'dismissed' | 'unavailable'> => {
    const event = promptEventRef.current
    if (!event) return 'unavailable'
    try {
      const { outcome } = await event.prompt()
      return outcome
    } catch {
      return 'unavailable'
    } finally {
      // A captured event can only be used once; recompute so a dismissal leaves the row as-is
      // and a failure falls back to instructions (Safari) or a passive row.
      promptEventRef.current = null
      setState(computeState())
    }
  }, [computeState])

  return { state, startInstall }
}
