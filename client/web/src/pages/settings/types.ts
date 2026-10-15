export interface SettingsViewProps {
  displayName: string
  username: string
  currencyCode: string | undefined
  onLogout: () => void
}
