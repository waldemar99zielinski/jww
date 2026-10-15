import { act, renderHook } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { useInstallPrompt } from '@/hooks/useInstallPrompt'

const ANDROID_CHROME_UA =
  'Mozilla/5.0 (Linux; Android 14; Pixel 8) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Mobile Safari/537.36'
const IOS_SAFARI_UA =
  'Mozilla/5.0 (iPhone; CPU iPhone OS 17_5 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.5 Mobile/15E148 Safari/604.1'
const DESKTOP_FIREFOX_UA = 'Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:126.0) Gecko/20100101 Firefox/126.0'

const originalUserAgent = window.navigator.userAgent

function setUserAgent(userAgent: string) {
  Object.defineProperty(window.navigator, 'userAgent', { configurable: true, get: () => userAgent })
}

interface FakePromptEvent extends Event {
  prompt: () => Promise<{ outcome: 'accepted' | 'dismissed' }>
}

function fireBeforeInstallPrompt(prompt: () => Promise<{ outcome: 'accepted' | 'dismissed' }>) {
  const event = new Event('beforeinstallprompt') as FakePromptEvent
  event.prompt = prompt
  window.dispatchEvent(event)
  return event
}

afterEach(() => {
  setUserAgent(originalUserAgent)
  vi.restoreAllMocks()
})

describe('useInstallPrompt', () => {
  it('stays hidden on Chromium until beforeinstallprompt fires', () => {
    setUserAgent(ANDROID_CHROME_UA)
    const { result } = renderHook(() => useInstallPrompt())
    expect(result.current.state).toEqual({ kind: 'hidden' })
  })

  it('becomes an actionable prompt after beforeinstallprompt, and startInstall maps the outcome', async () => {
    setUserAgent(ANDROID_CHROME_UA)
    const prompt = vi.fn().mockResolvedValue({ outcome: 'dismissed' })
    const { result } = renderHook(() => useInstallPrompt())

    act(() => {
      fireBeforeInstallPrompt(prompt)
    })
    expect(result.current.state).toEqual({ kind: 'prompt', run: 'browser-dialog' })

    let outcome: 'accepted' | 'dismissed' | 'unavailable' = 'unavailable'
    await act(async () => {
      outcome = await result.current.startInstall()
    })

    expect(outcome).toBe('dismissed')
    expect(prompt).toHaveBeenCalledTimes(1)
    // The captured event is single-use, so the row recomputes (hidden on Chromium).
    expect(result.current.state).toEqual({ kind: 'hidden' })
  })

  it('returns "unavailable" when there is no captured prompt to run', async () => {
    setUserAgent(ANDROID_CHROME_UA)
    const { result } = renderHook(() => useInstallPrompt())
    await expect(result.current.startInstall()).resolves.toBe('unavailable')
  })

  it('hides the row once appinstalled fires', () => {
    setUserAgent(ANDROID_CHROME_UA)
    const { result } = renderHook(() => useInstallPrompt())

    act(() => {
      fireBeforeInstallPrompt(vi.fn().mockResolvedValue({ outcome: 'accepted' }))
    })
    expect(result.current.state.kind).toBe('prompt')

    act(() => {
      window.dispatchEvent(new Event('appinstalled'))
    })
    expect(result.current.state).toEqual({ kind: 'hidden' })
  })

  it('offers iOS instructions with no prompt involved', () => {
    setUserAgent(IOS_SAFARI_UA)
    const { result } = renderHook(() => useInstallPrompt())
    expect(result.current.state).toEqual({ kind: 'instructions', platform: 'ios-safari' })
  })

  it('falls back to an unavailable row on browsers with no install path', () => {
    setUserAgent(DESKTOP_FIREFOX_UA)
    const { result } = renderHook(() => useInstallPrompt())
    expect(result.current.state).toEqual({ kind: 'unavailable', reason: 'browser' })
  })
})
