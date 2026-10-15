import type { TransactionCategoryDto, TransactionType } from '@/api/types'

export interface SettingsCategoriesViewProps {
  type: TransactionType
  onSelectType: (type: TransactionType) => void
  categories: TransactionCategoryDto[] | undefined
  isLoading: boolean
  onToggleActive: (category: TransactionCategoryDto) => void
  name: string
  onNameChange: (value: string) => void
  onCreate: () => void
  creating: boolean
  createError: boolean
  onClose: () => void
}
