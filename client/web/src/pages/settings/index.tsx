import { lazy } from 'react'
import { ResponsiveView } from '@/components/ResponsiveView'
import { useAuth } from '@/hooks/useAuth'
import { useDefaultCurrency } from '@/hooks/useDefaultCurrency'
import type { SettingsViewProps } from './types'

const Mobile = lazy(() => import('./Settings.mobile'))
const Desktop = lazy(() => import('./Settings.desktop'))

export default function SettingsPage() {
  const { user, logout } = useAuth()
  const { currency } = useDefaultCurrency()

  const props: SettingsViewProps = {
    displayName: user?.display_name ?? 'Guest',
    username: user?.username ?? '',
    currencyCode: currency?.code,
    onLogout: logout,
  }

  return <ResponsiveView mobile={Mobile} desktop={Desktop} props={props} />
}
