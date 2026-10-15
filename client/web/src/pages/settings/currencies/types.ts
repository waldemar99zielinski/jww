import type { CurrencyDto } from '@/api/types'

export interface SettingsCurrenciesViewProps {
  currencies: CurrencyDto[] | undefined
  isLoading: boolean
  onMakeBase: (currency: CurrencyDto) => void
  code: string
  onCodeChange: (value: string) => void
  name: string
  onNameChange: (value: string) => void
  onCreate: () => void
  creating: boolean
  hasError: boolean
  onClose: () => void
}
