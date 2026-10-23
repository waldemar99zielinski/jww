import type { CurrencyDto, SubTransactionDto, TransactionCategoryDto, TransactionType, UserDto } from '@/api/types'
import type { AttachmentPreviewItem } from '@/components/AttachmentPreview'

/** One label/value pair in the detail screen's field list, already formatted for display. */
export interface TransactionDetailItem {
  label: string
  value: string
}

/**
 * A ready-to-render attachment row. The container resolves the server's raw attachment (including
 * the short-lived `preview_grant`) into a usable `previewUrl`, so the views stay free of URL logic.
 */
export type TransactionAttachmentItem = AttachmentPreviewItem

/**
 * The editable form that replaces the read-only fields while the detail screen is in edit mode.
 * Mirrors the Add Transaction form's field set; the container owns every value and handler, the
 * views only render them (via `TransactionEditForm`).
 */
export interface TransactionDetailEditProps {
  type: TransactionType
  amount: string
  /** True while the loaded transaction has line items, so the Amount shows the derived Σ read-only. */
  amountReadOnly?: boolean
  /** Hint shown under a derived Amount, e.g. "= sum of 2 items". */
  amountNote?: string
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

  /** The loaded transaction's line items. The edit path saves each row immediately (create/patch/delete). */
  subTransactions?: SubTransactionDto[]
  /** Whether the shared line-item panel is expanded (collapsed by default, same as Add). */
  subTransactionsExpanded: boolean
  onToggleSubTransactions: () => void
  /** A line-item mutation is in flight — the panel locks while the request is in flight. */
  subTransactionsBusy?: boolean
  /** Inline message when a line-item mutation failed; the row list is never optimistically changed. */
  subTransactionError?: string | null
  onAddSubTransaction?: () => void
  onEditSubTransaction?: (id: string) => void
  onRemoveSubTransaction?: (id: string) => void

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
  /** The loaded line items, shown in the same collapsible panel the Add screen uses. */
  subTransactions: SubTransactionDto[]
  /** Transaction currency, used to format line-item amounts on the read-only panel. */
  currencyCode: string
  /** Shared line-item panel expansion (collapsed by default, exactly like Add). */
  subTransactionsExpanded: boolean
  onToggleSubTransactions: () => void
  attachments: TransactionAttachmentItem[]
  /** An upload is in flight; the Add attachment control shows its busy state. */
  attachmentsUploading: boolean
  /** Inline error when the last transaction-attachment upload failed. */
  attachmentUploadError: boolean
  downloadError: boolean
  isLoading: boolean
  isError: boolean
  /** Whether the editable form is showing in place of the read-only fields. */
  isEditing: boolean
  /** Enter edit mode; seeded by the container from the fetched transaction. */
  onStartEdit: () => void
  /** Open the file picker and upload the chosen file(s) onto this transaction. */
  onAddAttachment: () => void
  /** Ask for confirmation before deleting the transaction, its items, and its attachments. */
  onRequestDelete: () => void
  edit: TransactionDetailEditProps
  onPreview: (attachment: TransactionAttachmentItem) => void
  onOpenActions: (attachment: TransactionAttachmentItem) => void
  onClose: () => void
}
