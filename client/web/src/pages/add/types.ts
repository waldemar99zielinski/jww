import type { CurrencyDto, SubTransactionDto, TransactionCategoryDto, TransactionType, UserDto } from '@/api/types'
import type { LowConfidenceFields } from '@/components/TransactionFormFields'

export type TransactionAttachmentStatus = 'uploading' | 'processing' | 'ready' | 'error'

/** One attachment read's lifecycle, shared by the attachment bar and the page container. */
export type ExtractionStatus = 'idle' | 'reading' | 'success' | 'failed'

/** One file the user attached to the transaction being composed on the Add screen. */
export interface TransactionAttachmentDraft {
  /** Client-only stable key for list rendering and targeted removal; never sent to the server. */
  id: string
  name: string
  status: TransactionAttachmentStatus
  /** Server draft id, set once the upload succeeds; linked to the transaction on save. */
  attachmentId?: string
  /** MIME type from the picked file / upload response, used by the preview lightbox. */
  mediaType?: string
  /** Size in bytes from the picked file / upload response, used by the preview lightbox. */
  byteSize?: number
  previewUrl?: string
  isImage?: boolean
}

/** Where a local line item came from, so a re-scan can refresh extracted rows without losing manual ones. */
export type SubTransactionOrigin = 'manual' | 'extraction'

/** One line item the Add screen owns before the transaction exists; `id` is client-only. */
export interface SubTransactionDraft {
  id: string
  name: string
  label: string
  /** Whole minor units (cents), matching the create payload. */
  amount: number
  origin: SubTransactionOrigin
  /** The reader's score for an extracted row; absent for rows added by hand. */
  confidence?: number
}

/** What one successful read added to the draft, shown as a compact confirmation on the main page. */
export interface ExtractionSummary {
  /** Ordered attachments that produced this result. */
  attachmentIds: string[]
  /** Line items the read provided (not the current list length, which may include manual rows). */
  itemCount: number
  /** The reader's own total in cents, kept as a checksum against the item sum. */
  attachmentTotal?: number
  warnings: string[]
  /** A category name the read proposed but the loaded list could not resolve. */
  categoryHint?: string
}

export interface AddTransactionViewProps {
  type: TransactionType
  onSelectType: (type: TransactionType) => void

  amount: string
  onAmountChange: (value: string) => void
  /** True while the local item list is non-empty, so the Amount field shows the derived Σ read-only. */
  amountReadOnly?: boolean
  /** Hint shown under a derived Amount, e.g. "= sum of 3 items". */
  amountNote?: string
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
  extractionStatus: ExtractionStatus
  extractionSummary?: ExtractionSummary
  /** Extracted fields below the confidence threshold; cleared as the user reviews each one. */
  lowConfidenceFields?: LowConfidenceFields
  /** At least one uploaded image is ready and no upload is still in flight. */
  canScanAttachments: boolean
  /** The last successful read can still be rolled back. */
  canUndoExtraction: boolean
  attachmentPickerOpen: boolean
  onOpenAttachment: () => void
  onCloseAttachment: () => void
  onPickAttachment: (file: File) => void
  onRemoveAttachment: (id: string) => void
  /** Open the shared attachment lightbox for one uploaded draft. */
  onPreviewAttachment: (id: string) => void
  /** Read every ready image in one combined request. */
  onScanAttachments: () => void
  /** Abort the grouped read currently in flight. */
  onCancelExtraction: () => void
  /** Restore the form to exactly what it was before the last successful read. */
  onUndoExtraction: () => void

  /** Local line items, whole minor units internally and display-ready here. */
  subTransactionItems?: SubTransactionDto[]
  /** Extracted line-item ids scored below the confidence threshold. */
  lowConfidenceSubTransactionIds?: ReadonlySet<string>
  /** Collapsed by default; the summary row expands the panel inline. */
  subTransactionsExpanded: boolean
  onToggleSubTransactions: () => void
  /** Expand the panel from the attachment bar's Review action. */
  onReviewSubTransactions: () => void
  /** Open the add-item modal. */
  onAddSubTransaction?: () => void
  /** Open the edit modal for one row. */
  onEditSubTransaction?: (id: string) => void
  onRemoveSubTransaction?: (id: string) => void

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
