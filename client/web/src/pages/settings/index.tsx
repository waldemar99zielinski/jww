import { lazy, useState } from 'react'
import { ResponsiveView } from '@/components/ResponsiveView'
import { useAuth } from '@/hooks/useAuth'
import { useDefaultCurrency } from '@/hooks/useDefaultCurrency'
import { useInstallPrompt } from '@/hooks/useInstallPrompt'
import { usePlatformAuthenticator } from '@/hooks/usePlatformAuthenticator'
import { APP_VERSION } from '@/lib/version'
import type { InstallViewProps, SettingsViewProps } from './types'

const Mobile = lazy(() => import('./Settings.mobile'))
const Desktop = lazy(() => import('./Settings.desktop'))

const UNAVAILABLE_TRAILING: Record<'browser' | 'webview', string> = {
  browser: 'Not available in this browser',
  webview: 'Open in a browser to install',
}

export default function SettingsPage() {
  const { user, logout } = useAuth()
  const { currency } = useDefaultCurrency()
  const { state, startInstall } = useInstallPrompt()
  const showPasskeys = usePlatformAuthenticator()
  const [instructionsOpen, setInstructionsOpen] = useState(false)

  // All install logic is resolved here so both view variants only render the result.
  const install: InstallViewProps = {
    show: state.kind !== 'hidden',
    onClick:
      state.kind === 'prompt'
        ? () => void startInstall()
        : state.kind === 'instructions'
          ? () => setInstructionsOpen(true)
          : undefined,
    trailing: state.kind === 'unavailable' ? UNAVAILABLE_TRAILING[state.reason] : undefined,
    platform: state.kind === 'instructions' ? state.platform : null,
    instructionsOpen,
    onCloseInstructions: () => setInstructionsOpen(false),
  }

  const props: SettingsViewProps = {
    displayName: user?.display_name ?? 'Guest',
    username: user?.username ?? '',
    currencyCode: currency?.code,
    version: APP_VERSION,
    onLogout: logout,
    install,
    showPasskeys,
  }

  return <ResponsiveView mobile={Mobile} desktop={Desktop} props={props} />
}
