import type { WebAuthnCredentialDto } from '@/api/types'

export interface SettingsPasskeysViewProps {
  credentials: WebAuthnCredentialDto[] | undefined
  isLoading: boolean
  name: string
  onNameChange: (value: string) => void
  onEnroll: () => void
  enrolling: boolean
  enrollError: string | null
  onDelete: (credential: WebAuthnCredentialDto) => void
  onClose: () => void
}
