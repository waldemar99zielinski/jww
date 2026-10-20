import { afterEach, describe, expect, it, vi } from 'vitest'
import { isOnline, isStandalone, registerServiceWorker } from '@/lib/pwa'
import { registerSW } from 'virtual:pwa-register'

vi.mock('virtual:pwa-register', () => ({ registerSW: vi.fn() }))

function setOnLine(value: boolean) {
  Object.defineProperty(window.navigator, 'onLine', { configurable: true, get: () => value })
}

function setNavigatorStandalone(value: boolean | undefined) {
  Object.defineProperty(window.navigator, 'standalone', { configurable: true, value })
}

const originalMatchMedia = window.matchMedia

afterEach(() => {
  setOnLine(true)
  setNavigatorStandalone(undefined)
  window.matchMedia = originalMatchMedia
  vi.clearAllMocks()
})

describe('isOnline', () => {
  it('reads navigator.onLine', () => {
    setOnLine(false)
    expect(isOnline()).toBe(false)
    setOnLine(true)
    expect(isOnline()).toBe(true)
  })
})

describe('isStandalone', () => {
  it('reads the display-mode: standalone media query', () => {
    window.matchMedia = vi.fn().mockReturnValue({ matches: true }) as unknown as typeof window.matchMedia
    expect(isStandalone()).toBe(true)

    window.matchMedia = vi.fn().mockReturnValue({ matches: false }) as unknown as typeof window.matchMedia
    expect(isStandalone()).toBe(false)
  })

  it('falls back to navigator.standalone when the media query misses on iOS', () => {
    window.matchMedia = vi.fn().mockReturnValue({ matches: false }) as unknown as typeof window.matchMedia
    setNavigatorStandalone(true)

    expect(isStandalone()).toBe(true)
  })

  it('stays false in a browser tab on any platform', () => {
    window.matchMedia = vi.fn().mockReturnValue({ matches: false }) as unknown as typeof window.matchMedia
    setNavigatorStandalone(false)

    expect(isStandalone()).toBe(false)
  })
})

describe('registerServiceWorker', () => {
  it('is a no-op outside a production build', () => {
    registerServiceWorker()
    expect(registerSW).not.toHaveBeenCalled()
  })
})
