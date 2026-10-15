import { describe, expect, it } from 'vitest'
import {
  detectPlatform,
  isInAppWebView,
  resolveInstallUiState,
  type InstallEnvironment,
} from '@/lib/installEnvironment'

const UA = {
  androidChrome:
    'Mozilla/5.0 (Linux; Android 14; Pixel 8) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Mobile Safari/537.36',
  androidFirefox: 'Mozilla/5.0 (Android 14; Mobile; rv:126.0) Gecko/126.0 Firefox/126.0',
  iosSafari:
    'Mozilla/5.0 (iPhone; CPU iPhone OS 17_5 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.5 Mobile/15E148 Safari/604.1',
  macSafari:
    'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.5 Safari/605.1.15',
  macChrome:
    'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36',
  windowsChrome:
    'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36',
  windowsEdge:
    'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36 Edg/124.0.0.0',
  desktopFirefox: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:126.0) Gecko/20100101 Firefox/126.0',
  telegramAndroid:
    'Mozilla/5.0 (Linux; Android 14; Pixel 8) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Mobile Safari/537.36 Telegram/10.9',
  instagramIos:
    'Mozilla/5.0 (iPhone; CPU iPhone OS 17_5 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/15E148 Instagram 330.0.0.0',
  facebookAndroid:
    'Mozilla/5.0 (Linux; Android 14; Pixel 8) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Mobile Safari/537.36 [FBAN/FB4A;FBAV/450.0.0.0;]',
  unknown: 'SomeCustomBrowser/1.0',
}

describe('detectPlatform', () => {
  it('detects iOS as ios-safari regardless of the specific iOS browser', () => {
    expect(detectPlatform(UA.iosSafari)).toBe('ios-safari')
    expect(detectPlatform(UA.instagramIos)).toBe('ios-safari')
  })

  it('detects macOS Safari without mistaking macOS Chrome for it', () => {
    expect(detectPlatform(UA.macSafari)).toBe('macos-safari')
    expect(detectPlatform(UA.macChrome)).toBe('desktop-chromium')
  })

  it('detects Android Chrome and desktop Chromium (Chrome/Edge)', () => {
    expect(detectPlatform(UA.androidChrome)).toBe('android-chromium')
    expect(detectPlatform(UA.windowsChrome)).toBe('desktop-chromium')
    expect(detectPlatform(UA.windowsEdge)).toBe('desktop-chromium')
  })

  it('falls back to "other" for browsers with no install path', () => {
    expect(detectPlatform(UA.androidFirefox)).toBe('other')
    expect(detectPlatform(UA.desktopFirefox)).toBe('other')
    expect(detectPlatform(UA.unknown)).toBe('other')
  })
})

describe('isInAppWebView', () => {
  it('flags the known in-app browsers', () => {
    expect(isInAppWebView(UA.telegramAndroid)).toBe(true)
    expect(isInAppWebView(UA.facebookAndroid)).toBe(true)
    expect(isInAppWebView(UA.instagramIos)).toBe(true)
  })

  it('does not flag ordinary browsers', () => {
    expect(isInAppWebView(UA.androidChrome)).toBe(false)
    expect(isInAppWebView(UA.iosSafari)).toBe(false)
  })
})

describe('resolveInstallUiState', () => {
  const base: InstallEnvironment = { platform: 'other', isStandalone: false, hasPrompt: false, isWebView: false }

  it('hides the row when the app is already installed', () => {
    expect(resolveInstallUiState({ ...base, platform: 'android-chromium', hasPrompt: true, isStandalone: true })).toEqual({
      kind: 'hidden',
    })
  })

  it('hides the row when the app is already installed, even in a webview', () => {
    expect(resolveInstallUiState({ ...base, platform: 'ios-safari', isStandalone: true, isWebView: true })).toEqual({
      kind: 'hidden',
    })
  })

  it('offers instructions on iOS and macOS Safari (they cannot be prompted)', () => {
    expect(resolveInstallUiState({ ...base, platform: 'ios-safari' })).toEqual({
      kind: 'instructions',
      platform: 'ios-safari',
    })
    expect(resolveInstallUiState({ ...base, platform: 'macos-safari' })).toEqual({
      kind: 'instructions',
      platform: 'macos-safari',
    })
  })

  it('offers the browser dialog on Chromium once a prompt has been captured', () => {
    expect(resolveInstallUiState({ ...base, platform: 'android-chromium', hasPrompt: true })).toEqual({
      kind: 'prompt',
      run: 'browser-dialog',
    })
    expect(resolveInstallUiState({ ...base, platform: 'desktop-chromium', hasPrompt: true })).toEqual({
      kind: 'prompt',
      run: 'browser-dialog',
    })
  })

  it('stays hidden on Chromium before beforeinstallprompt fires', () => {
    expect(resolveInstallUiState({ ...base, platform: 'android-chromium' })).toEqual({ kind: 'hidden' })
    expect(resolveInstallUiState({ ...base, platform: 'desktop-chromium' })).toEqual({ kind: 'hidden' })
  })

  it('reports a passive "browser" reason on platforms with no install path', () => {
    expect(resolveInstallUiState(base)).toEqual({ kind: 'unavailable', reason: 'browser' })
  })

  it('reports a passive "webview" reason inside in-app browsers, before platform handling', () => {
    expect(resolveInstallUiState({ ...base, platform: 'ios-safari', isWebView: true })).toEqual({
      kind: 'unavailable',
      reason: 'webview',
    })
    expect(resolveInstallUiState({ ...base, platform: 'android-chromium', isWebView: true, hasPrompt: true })).toEqual({
      kind: 'unavailable',
      reason: 'webview',
    })
  })
})
