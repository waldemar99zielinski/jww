/**
 * Single-flag browser accessors plus service-worker registration, in the spirit of
 * lib/viewport.ts / lib/designMode.ts: no React, no fetch, no app state.
 */

/** Live connectivity per the browser — the same source the offline gate reacts to. */
export function isOnline(): boolean {
  return navigator.onLine
}

/** True when the app runs in an installed/standalone window rather than a browser tab. */
export function isStandalone(): boolean {
  return window.matchMedia('(display-mode: standalone)').matches
}

/**
 * Registers the Workbox-generated worker (auto-update, silent). No-op outside a production
 * build: `npm run dev` and tests must not install a worker, and the dev server serves its own
 * `devOptions` worker separately.
 */
export function registerServiceWorker(): void {
  if (!import.meta.env.PROD) return
  void import('virtual:pwa-register')
    .then(({ registerSW }) => {
      registerSW({ immediate: true })
    })
    // The registration chunk is fetched lazily; on an offline cold start that fetch fails. It is
    // best-effort (the worker is already installed), so swallow the rejection.
    .catch(() => {})
}
