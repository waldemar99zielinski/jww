import '@testing-library/jest-dom/vitest'
import { cleanup } from '@testing-library/react'
import { afterAll, afterEach, beforeAll } from 'vitest'
import { server } from '@/test/server'

// jsdom implements no CSS media evaluation and, in this setup, no `matchMedia` at all. Several
// components read it (e.g. `lib/pwa.ts`'s `isStandalone`, `Logo`'s reduced-motion check), so
// provide the smallest spec-shaped stub; individual tests override it when `matches` matters.
if (typeof window !== 'undefined' && typeof window.matchMedia !== 'function') {
  window.matchMedia = (query: string): MediaQueryList =>
    ({
      matches: false,
      media: query,
      onchange: null,
      addEventListener: () => {},
      removeEventListener: () => {},
      addListener: () => {},
      removeListener: () => {},
      dispatchEvent: () => false,
    }) as MediaQueryList
}

// RTL's own auto-cleanup only registers against a *global* `afterEach` — this project runs
// Vitest with `globals: false`, so it never fires on its own. Without this, DOM from one test
// leaks into the next within the same file (e.g. two renders both matching the same role/text).
afterEach(() => cleanup())

beforeAll(() => server.listen({ onUnhandledFrame: 'error' }))
afterEach(() => server.resetHandlers())
afterAll(() => server.close())
