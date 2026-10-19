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

// jsdom implements no image decoding. Attachment preparation downscales raster images with
// `createImageBitmap`; provide a no-op 1×1 bitmap so uploads can flow through tests. (JPEGs at or
// below the resize threshold are returned unchanged before any canvas work.)
if (typeof globalThis.createImageBitmap !== 'function') {
  globalThis.createImageBitmap = (async () => ({ width: 1, height: 1, close: () => {} })) as typeof createImageBitmap
}

// RTL's own auto-cleanup only registers against a *global* `afterEach` — this project runs
// Vitest with `globals: false`, so it never fires on its own. Without this, DOM from one test
// leaks into the next within the same file (e.g. two renders both matching the same role/text).
afterEach(() => cleanup())

beforeAll(() => server.listen({ onUnhandledFrame: 'error' }))
afterEach(() => server.resetHandlers())
afterAll(() => server.close())
