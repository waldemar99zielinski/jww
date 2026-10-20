/**
 * Single-flag browser accessors plus service-worker registration, in the spirit of
 * lib/viewport.ts / lib/designMode.ts: no React, no fetch, no app state.
 */

/** Live connectivity per the browser — the same source the offline gate reacts to. */
export function isOnline(): boolean {
  return navigator.onLine
}

/** iOS still exposes this non-standard flag for home-screen apps. */
interface NavigatorWithStandalone extends Navigator {
  standalone?: boolean
}

/**
 * True when the app runs in an installed/standalone window rather than a browser tab.
 * Checks both the standard media query and iOS's legacy `navigator.standalone`, which can
 * disagree on home-screen launches.
 */
export function isStandalone(): boolean {
  return (
    window.matchMedia('(display-mode: standalone)').matches ||
    (window.navigator as NavigatorWithStandalone).standalone === true
  )
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
