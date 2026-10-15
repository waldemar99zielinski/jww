import { afterEach, describe, expect, it, vi } from 'vitest'
import { isDesignMode } from '@/lib/designMode'

describe('isDesignMode', () => {
  afterEach(() => vi.unstubAllEnvs())

  it('returns false when VITE_DESIGN_MODE is unset', () => {
    vi.stubEnv('VITE_DESIGN_MODE', undefined)
    expect(isDesignMode()).toBe(false)
  })

  it('returns false when VITE_DESIGN_MODE is not the exact string "true"', () => {
    vi.stubEnv('VITE_DESIGN_MODE', 'yes')
    expect(isDesignMode()).toBe(false)
  })

  it('returns true when VITE_DESIGN_MODE is "true"', () => {
    vi.stubEnv('VITE_DESIGN_MODE', 'true')
    expect(isDesignMode()).toBe(true)
  })
})
