/**
 * The install decision tree as pure functions, so every branch is unit-testable without a
 * browser. The caller (hooks/useInstallPrompt) owns the imperative `beforeinstallprompt` API;
 * this module only maps an environment snapshot to what the Settings row should render.
 */

/** Where an install path is available, derived from the user agent. */
export type InstallPlatform = 'android-chromium' | 'ios-safari' | 'desktop-chromium' | 'macos-safari' | 'other'

/** Everything the decision function needs, read once per render. */
export interface InstallEnvironment {
  platform: InstallPlatform
  isStandalone: boolean
  hasPrompt: boolean
  isWebView: boolean
}

/**
 * The row's resolved UI state. `hidden` covers both "already installed" and "Chromium hasn't
 * fired `beforeinstallprompt` yet" — offering a dead action is worse than showing nothing.
 */
export type InstallUiState =
  | { kind: 'hidden' }
  | { kind: 'prompt'; run: 'browser-dialog' }
  | { kind: 'instructions'; platform: 'ios-safari' | 'macos-safari' }
  | { kind: 'unavailable'; reason: 'browser' | 'webview' }

const IOS_RE = /iPhone|iPad|iPod/
const ANDROID_RE = /Android/
const CHROMIUM_RE = /Chrome|Chromium|Edg\//
const SAFARI_RE = /Safari/
const MAC_RE = /Macintosh/
// UA sniffing is the only tool available for "in-app browser"; keep the list short and documented.
const WEBVIEW_RE = /Telegram|FBAN|FBAV|Instagram|Line\//

export function isInAppWebView(userAgent: string): boolean {
  return WEBVIEW_RE.test(userAgent)
}

export function detectPlatform(userAgent: string): InstallPlatform {
  if (IOS_RE.test(userAgent)) return 'ios-safari'

  const chromium = CHROMIUM_RE.test(userAgent)

  if (ANDROID_RE.test(userAgent)) return chromium ? 'android-chromium' : 'other'
  if (MAC_RE.test(userAgent)) {
    if (chromium) return 'desktop-chromium'
    return SAFARI_RE.test(userAgent) ? 'macos-safari' : 'other'
  }
  return chromium ? 'desktop-chromium' : 'other'
}

/**
 * The whole install decision tree. Unknown environments fall back to the passive
 * "not available in this browser" row — never a broken button.
 */
export function resolveInstallUiState(env: InstallEnvironment): InstallUiState {
  if (env.isStandalone) return { kind: 'hidden' }
  if (env.isWebView) return { kind: 'unavailable', reason: 'webview' }

  switch (env.platform) {
    case 'ios-safari':
    case 'macos-safari':
      return { kind: 'instructions', platform: env.platform }
    case 'android-chromium':
    case 'desktop-chromium':
      // Chromium can prompt, but only after its engagement heuristics fire; before then there
      // is no actionable row, so stay hidden rather than offer a button that does nothing.
      return env.hasPrompt ? { kind: 'prompt', run: 'browser-dialog' } : { kind: 'hidden' }
    case 'other':
      return { kind: 'unavailable', reason: 'browser' }
  }
}
