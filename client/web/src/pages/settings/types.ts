import type { ReactNode } from 'react'
import type { InstallInstructionsPlatform } from './components/InstallInstructions'

/**
 * The install row, pre-resolved by the container so both view variants render it identically.
 * `show: false` hides the row entirely (already installed, or Chromium hasn't offered a prompt
 * yet). `trailing` carries the reason for a non-actionable row (unsupported browser / webview).
 */
export interface InstallViewProps {
  show: boolean
  onClick?: () => void
  trailing?: ReactNode
  platform: InstallInstructionsPlatform | null
  instructionsOpen: boolean
  onCloseInstructions: () => void
}

export interface SettingsViewProps {
  displayName: string
  username: string
  currencyCode: string | undefined
  version: string
  onLogout: () => void
  install: InstallViewProps
  /** Hides the Face ID/biometrics row when no platform authenticator is available. */
  showPasskeys: boolean
}
