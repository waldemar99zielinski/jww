import type { CurrencyDto, TransactionCategoryDto, TransactionType, UserDto } from '@/api/types'

export type ReceiptStatus = 'parsing' | 'ready'

export interface ReceiptDraft {
  name: string
  status: ReceiptStatus
  previewUrl?: string
  isImage?: boolean
}

export interface AddTransactionViewProps {
  type: TransactionType
  onSelectType: (type: TransactionType) => void

  amount: string
  onAmountChange: (value: string) => void
  currencyCode: string | undefined
  currencies: CurrencyDto[]
  onCurrencyChange: (code: string) => void
  hasCurrency: boolean

  categories: TransactionCategoryDto[]
  categoryId: string | null
  onSelectCategory: (id: string) => void

  date: string
  time: string
  onDateChange: (value: string) => void
  onTimeChange: (value: string) => void
  onUseToday: () => void
  dateLabel: string
  dateOpen: boolean
  onToggleDate: () => void

  name: string
  onNameChange: (value: string) => void
  description: string
  onDescriptionChange: (value: string) => void

  receipt: ReceiptDraft | null
  receiptPickerOpen: boolean
  onOpenReceipt: () => void
  onCloseReceipt: () => void
  onPickReceipt: (file: File) => void
  onRemoveReceipt: () => void

  users: UserDto[]
  userId: string
  onUserChange: (id: string) => void
  showUserPicker: boolean

  moreOpen: boolean
  onToggleMore: () => void

  canSave: boolean
  saving: boolean
  saveError: boolean
  onSave: () => void
  onClose: () => void
}
