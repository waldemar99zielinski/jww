import { afterEach, describe, expect, it } from 'vitest'
import { DESKTOP_BREAKPOINT_PX, getViewport } from '@/lib/viewport'

function setInnerWidth(width: number) {
  Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: width })
}

const originalInnerWidth = window.innerWidth

describe('getViewport', () => {
  afterEach(() => setInnerWidth(originalInnerWidth))

  it('returns "desktop" at the breakpoint width', () => {
    setInnerWidth(DESKTOP_BREAKPOINT_PX)
    expect(getViewport()).toBe('desktop')
  })

  it('returns "desktop" above the breakpoint width', () => {
    setInnerWidth(DESKTOP_BREAKPOINT_PX + 200)
    expect(getViewport()).toBe('desktop')
  })

  it('returns "mobile" below the breakpoint width', () => {
    setInnerWidth(DESKTOP_BREAKPOINT_PX - 1)
    expect(getViewport()).toBe('mobile')
  })
})
