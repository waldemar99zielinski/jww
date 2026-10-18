import type { CurrencyDto, TransactionCategoryDto, TransactionType, UserDto } from '@/api/types'

/** One label/value pair in the detail screen's field list, already formatted for display. */
export interface TransactionDetailItem {
  label: string
  value: string
}

/**
 * A ready-to-render attachment row. The container resolves the server's raw attachment (including
 * the short-lived `preview_grant`) into a usable `previewUrl`, so the views stay free of URL logic.
 */
export interface TransactionAttachmentItem {
  id: string
  filename: string
  mediaType: string
  byteSize: number
  /** Inline preview URL for previewable media (image/PDF), or null for download-only files. */
  previewUrl: string | null
}

/**
 * The editable form that replaces the read-only fields while the detail screen is in edit mode.
 * Mirrors the Add Transaction form's field set; the container owns every value and handler, the
 * views only render them (via `TransactionEditForm`).
 */
export interface TransactionDetailEditProps {
  type: TransactionType
  amount: string
  currencyCode: string | undefined
  currencies: CurrencyDto[]
  hasCurrency: boolean
  categories: TransactionCategoryDto[]
  categoryId: string | null
  name: string
  description: string
  date: string
  time: string
  users: UserDto[]
  userId: string
  showUserPicker: boolean
  moreOpen: boolean
  canSave: boolean
  saving: boolean
  saveError: boolean
  /** True once any field differs from the transaction that was loaded — drives Reset vs. Cancel. */
  isDirty: boolean
  onSelectType: (type: TransactionType) => void
  onAmountChange: (value: string) => void
  onCurrencyChange: (code: string) => void
  onSelectCategory: (id: string) => void
  onNameChange: (value: string) => void
  onDescriptionChange: (value: string) => void
  onDateChange: (value: string) => void
  onTimeChange: (value: string) => void
  onUseToday: () => void
  onUserChange: (id: string) => void
  onToggleMore: () => void
  onSave: () => void
  onReset: () => void
  onCancel: () => void
}

export interface TransactionDetailViewProps {
  title: string
  categoryName: string
  amount: string
  isIncome: boolean
  items: TransactionDetailItem[]
  attachments: TransactionAttachmentItem[]
  downloadError: boolean
  isLoading: boolean
  isError: boolean
  /** Whether the editable form is showing in place of the read-only fields. */
  isEditing: boolean
  /** Enter edit mode; seeded by the container from the fetched transaction. */
  onStartEdit: () => void
  edit: TransactionDetailEditProps
  onPreview: (attachment: TransactionAttachmentItem) => void
  onOpenActions: (attachment: TransactionAttachmentItem) => void
  onClose: () => void
}
