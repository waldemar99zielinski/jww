import type { CurrencyDto, TransactionCategoryDto, TransactionType, UserDto } from '@/api/types'

export type TransactionAttachmentStatus = 'uploading' | 'processing' | 'ready' | 'error'

/** Whether picking a file only attaches it, or also reads it to prefill the form. */
export type TransactionAttachmentFlow = 'attach' | 'prefill'

/** One file the user attached to the transaction being composed on the Add screen. */
export interface TransactionAttachmentDraft {
  /** Client-only stable key for list rendering and targeted removal; never sent to the server. */
  id: string
  name: string
  status: TransactionAttachmentStatus
  flow: TransactionAttachmentFlow
  /** Server draft id, set once the upload succeeds; linked to the transaction on save. */
  attachmentId?: string
  previewUrl?: string
  isImage?: boolean
  /** Fields filled from the file after a successful read; drives the banner summary. */
  prefilledCount?: number
  /** The read failed, timed out or was cancelled — the file is still attached. */
  readFailed?: boolean
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

  attachments: TransactionAttachmentDraft[]
  attachmentPickerOpen: boolean
  onOpenAttachment: () => void
  onCloseAttachment: () => void
  onPickAttachment: (file: File, flow: TransactionAttachmentFlow) => void
  onRemoveAttachment: (id: string) => void

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
