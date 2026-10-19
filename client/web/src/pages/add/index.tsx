import { lazy, useEffect, useRef, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { deleteDraftAttachment, downloadDraftAttachment } from '@/api/attachments'
import type { SubTransactionDto, TransactionPrefillDto, TransactionType } from '@/api/types'
import { AttachmentPreview, type AttachmentPreviewItem } from '@/components/AttachmentPreview'
import type { SubTransactionListInput } from '@/components/SubTransactionList'
import type { TransactionFormField } from '@/components/TransactionFormFields'
import { ResponsiveView } from '@/components/ResponsiveView'
import { isLowConfidence } from './prefillConfidence'
import { SubTransactionModal } from './components/SubTransactionModal'
import { useAuth } from '@/hooks/useAuth'
import { useCategories } from '@/hooks/useCategories'
import { useDefaultCurrency } from '@/hooks/useDefaultCurrency'
import { useDeleteDraftAttachment, useExtractTransactionPrefill, useUploadDraftAttachment } from '@/hooks/useAttachments'
import { useCreateTransaction } from '@/hooks/useTransactions'
import { useUsers } from '@/hooks/useUsers'
import { formatShortDate, formatTime, toISODate, toISODateTime } from '@/lib/dates'
import { amountToCents, centsToAmountDisplay, decimalStringToAmountDisplay } from '@/lib/amount'
import { saveBlob } from '@/lib/download'
import type {
  AddTransactionViewProps,
  ExtractionStatus,
  ExtractionSummary,
  SubTransactionDraft,
  TransactionAttachmentDraft,
} from './types'

const Mobile = lazy(() => import('./AddTransaction.mobile'))
const Desktop = lazy(() => import('./AddTransaction.desktop'))

const TIME_PATTERN = /^([01]\d|2[0-3]):[0-5]\d$/

/** The form values an extraction may overwrite, captured so "Undo scan" can restore them exactly. */
interface ExtractionUndo {
  type: TransactionType
  amount: string
  name: string
  description: string
  categoryId: string | null
  date: string
  time: string
  currencyCode: string
  subTransactionItems: SubTransactionDraft[]
}

/** A category name proposed by a read whose target type's list is still loading. */
interface PendingCategoryHint {
  value: string
  type: TransactionType
}

/** Which line item the focused editor modal is open for; `create` starts blank. */
type SubTransactionEditor = { mode: 'create' } | { mode: 'edit'; id: string }

export default function AddTransactionPage() {
  const navigate = useNavigate()
  const [type, setType] = useState<TransactionType>('expense')
  const [amount, setAmount] = useState('')
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [categoryId, setCategoryId] = useState<string | null>(null)
  const [today] = useState(() => toISODate(new Date()))
  const [date, setDate] = useState(today)
  const [time, setTime] = useState(() => formatTime(new Date().toISOString()))
  const [dateOpen, setDateOpen] = useState(false)
  const [attachments, setAttachments] = useState<TransactionAttachmentDraft[]>([])
  const [attachmentPickerOpen, setAttachmentPickerOpen] = useState(false)
  /** Client row id of the attachment currently open in the shared lightbox. */
  const [previewAttachmentId, setPreviewAttachmentId] = useState<string | null>(null)
  const [previewDownloading, setPreviewDownloading] = useState(false)
  const [moreOpen, setMoreOpen] = useState(false)
  const [currencyCode, setCurrencyCode] = useState('')
  const [userOverride, setUserOverride] = useState('')
  /** Line items the user added by hand or accepted from a read; local until the transaction is created. */
  const [subTransactionItems, setSubTransactionItems] = useState<SubTransactionDraft[]>([])
  /** The panel starts collapsed so a scan only occupies one summary row until the user inspects it. */
  const [subTransactionsExpanded, setSubTransactionsExpanded] = useState(false)
  const [subTransactionEditor, setSubTransactionEditor] = useState<SubTransactionEditor | null>(null)
  const [extractionStatus, setExtractionStatus] = useState<ExtractionStatus>('idle')
  const [extractionSummary, setExtractionSummary] = useState<ExtractionSummary>()
  const [extractionUndo, setExtractionUndo] = useState<ExtractionUndo | null>(null)
  const [pendingCategoryHint, setPendingCategoryHint] = useState<PendingCategoryHint | null>(null)
  /** Extracted scalar fields the reader scored below the configured threshold. */
  const [lowConfidenceFields, setLowConfidenceFields] = useState<ReadonlySet<TransactionFormField>>(() => new Set())
  /** Extracted line-item ids scored below the configured threshold. */
  const [lowConfidenceSubTransactionIds, setLowConfidenceSubTransactionIds] = useState<ReadonlySet<string>>(
    () => new Set(),
  )

  const { user } = useAuth()
  const { currency, currencies } = useDefaultCurrency()
  const usersQuery = useUsers()
  const categoriesQuery = useCategories(type)
  const categories = categoriesQuery.data ?? []
  const createTransaction = useCreateTransaction()
  const uploadAttachment = useUploadDraftAttachment()
  const deleteAttachment = useDeleteDraftAttachment()
  const extractPrefill = useExtractTransactionPrefill()

  /** Draft ids uploaded but not yet linked; deleted on unmount so their bytes don't linger. */
  const pendingDraftIdsRef = useRef<Set<string>>(new Set())
  /** Monotonic key so every attached file keeps a stable identity across uploads and removals. */
  const nextAttachmentKeyRef = useRef(0)
  /** Separate monotonic key for local line-item rows (client-only ids). */
  const nextItemKeyRef = useRef(0)
  /** Cancels the one in-flight grouped read when the user cancels processing or leaves the screen. */
  const abortRef = useRef<Map<string, AbortController>>(new Map())
  /** Latest committed form values, so a read that resolves mid-typing snapshots what is on screen. */
  const formRef = useRef<ExtractionUndo>({
    type,
    amount,
    name,
    description,
    categoryId,
    date,
    time,
    currencyCode,
    subTransactionItems,
  })

  // Any draft the user uploaded but left without saving is deleted on unmount so its bytes don't
  // linger in object storage. Called directly (not through the hook) because this is a one-shot
  // cleanup, not a rendered mutation.
  useEffect(() => {
    const pendingDrafts = pendingDraftIdsRef.current
    const inFlightReads = abortRef.current
    return () => {
      for (const controller of inFlightReads.values()) controller.abort()
      for (const draftId of pendingDrafts) void deleteDraftAttachment(draftId).catch(() => {})
    }
  }, [])

  const users = usersQuery.data ?? (user ? [user] : [])
  const userId = userOverride || user?.id || ''

  /**
   * A read can propose a category name for a type whose list is still loading (a type change
   * swaps the category query). Rather than resolving it in an effect, the selected category is
   * derived during render: the hint wins until the user picks one explicitly, and the attachment
   * bar keeps showing it only while it cannot be matched.
   */
  const categoryHintName = pendingCategoryHint?.type === type ? pendingCategoryHint.value : undefined
  const hintedCategory = categoryHintName
    ? categories.find((category) => category.name.toLowerCase() === categoryHintName.toLowerCase())
    : undefined
  const effectiveCategoryId =
    pendingCategoryHint?.type === type && hintedCategory ? hintedCategory.id : categoryId
  const unresolvedCategoryHint =
    categoryHintName && !hintedCategory && !categoriesQuery.isFetching ? categoryHintName : undefined

  // Keep the pre-scan snapshot in sync with the latest committed render, including a category that
  // was resolved from an extraction hint rather than chosen explicitly.
  useEffect(() => {
    formRef.current = {
      type,
      amount,
      name,
      description,
      categoryId: effectiveCategoryId,
      date,
      time,
      currencyCode,
      subTransactionItems,
    }
  })

  /**
   * A low-confidence outline is a prompt to review the extracted value, so touching that field
   * (or choosing a replacement) retires it.
   */
  function clearLowConfidence(...fields: TransactionFormField[]) {
    setLowConfidenceFields((current) => {
      if (!fields.some((field) => current.has(field))) return current
      const next = new Set(current)
      for (const field of fields) next.delete(field)
      return next
    })
  }

  function clearLowConfidenceSubTransaction(id: string) {
    setLowConfidenceSubTransactionIds((current) => {
      if (!current.has(id)) return current
      const next = new Set(current)
      next.delete(id)
      return next
    })
  }

  function handleSelectType(next: TransactionType) {
    setType(next)
    setCategoryId(null)
    setPendingCategoryHint(null)
    clearLowConfidence('type')
  }

  function handleSelectCategory(id: string) {
    // An explicit choice wins over a still-resolving extraction hint.
    setPendingCategoryHint(null)
    setCategoryId(id || null)
    clearLowConfidence('category')
  }

  function handleUseToday() {
    setDate(toISODate(new Date()))
    setTime(formatTime(new Date().toISOString()))
    clearLowConfidence('date', 'time')
  }

  function handleAmountChange(value: string) {
    setAmount(value)
    clearLowConfidence('amount')
  }

  function handleCurrencyChange(code: string) {
    setCurrencyCode(code)
    clearLowConfidence('currency')
  }

  function handleNameChange(value: string) {
    setName(value)
    clearLowConfidence('name')
  }

  function handleDescriptionChange(value: string) {
    setDescription(value)
    clearLowConfidence('description')
  }

  function handleDateChange(value: string) {
    setDate(value)
    clearLowConfidence('date')
  }

  function handleTimeChange(value: string) {
    setTime(value)
    clearLowConfidence('time')
  }

  /** Merges a patch into the one attached file the patch targets, leaving its siblings untouched. */
  function updateAttachment(id: string, patch: Partial<TransactionAttachmentDraft>) {
    setAttachments((current) => current.map((a) => (a.id === id ? { ...a, ...patch } : a)))
  }

  function handleRemoveAttachment(id: string) {
    const target = attachments.find((a) => a.id === id)
    if (!target) return
    const controller = abortRef.current.get('extraction')
    if (controller) {
      controller.abort()
      abortRef.current.delete('extraction')
    }
    if (target.attachmentId) {
      pendingDraftIdsRef.current.delete(target.attachmentId)
      deleteAttachment.mutate(target.attachmentId)
    }
    if (target.previewUrl) URL.revokeObjectURL(target.previewUrl)
    if (previewAttachmentId === id) setPreviewAttachmentId(null)
    // The summary describes a specific set of files; once one leaves, the result no longer
    // matches what is attached. Keep the values that were already written into the form.
    if (target.attachmentId && extractionSummary?.attachmentIds.includes(target.attachmentId)) {
      setExtractionSummary(undefined)
      setExtractionStatus((current) => (current === 'reading' ? current : 'idle'))
    }
    setAttachments((current) => current.filter((a) => a.id !== id))
  }

  /**
   * The single write path for a successful read. Every proposed scalar fills its field and every
   * proposed line item replaces the previously extracted rows, while rows added by hand survive a
   * re-scan. The pre-read snapshot powers "Undo scan"; nothing is merged behind the user's back.
   */
  function applyPrefill(result: TransactionPrefillDto) {
    const before = formRef.current
    setExtractionUndo({ ...before, subTransactionItems: before.subTransactionItems.map((item) => ({ ...item })) })

    const proposedType = result.type?.value
    const targetType = proposedType ?? before.type
    if (proposedType && proposedType !== before.type) {
      setType(proposedType)
      setCategoryId(null)
    }

    const proposedAmount = result.amount?.value
    const proposedName = result.name?.value.trim()
    const proposedDescription = result.description?.value.trim()
    const proposedCurrency = result.currency_code?.value
    const supportsProposedCurrency = Boolean(proposedCurrency && currencies.some((option) => option.code === proposedCurrency))
    const proposedDate = result.occurred_on?.value
    const proposedTime = result.occurred_at?.value
    const validDate = Boolean(proposedDate && /^\d{4}-\d{2}-\d{2}$/.test(proposedDate))
    const validTime = Boolean(proposedTime && TIME_PATTERN.test(proposedTime))
    if (proposedAmount) setAmount(centsToAmountDisplay(proposedAmount))
    if (proposedName) setName(proposedName)
    if (proposedDescription) setDescription(proposedDescription)
    if (proposedCurrency && supportsProposedCurrency) setCurrencyCode(proposedCurrency)
    if (proposedDate && validDate) setDate(proposedDate)
    if (proposedTime && validTime) setTime(proposedTime)

    const extractedRows: SubTransactionDraft[] = (result.sub_transactions ?? []).map((item) => ({
      id: `item-${nextItemKeyRef.current++}`,
      name: item.name,
      label: item.label,
      amount: item.amount,
      origin: 'extraction',
      confidence: item.confidence,
    }))
    setSubTransactionItems((current) => [...extractedRows, ...current.filter((item) => item.origin === 'manual')])
    setLowConfidenceSubTransactionIds(
      new Set(
        extractedRows
          .filter((item) => isLowConfidence({ confidence: item.confidence }))
          .map((item) => item.id),
      ),
    )

    const categoryHint = result.category_hint?.value.trim()
    setPendingCategoryHint(categoryHint ? { value: categoryHint, type: targetType } : null)

    // Only fields that actually made it into the form are outlined for review. The amount is
    // skipped when line items take over the field (it becomes their derived sum), and absent
    // scores are "unscored" rather than low.
    const lowConfidence = new Set<TransactionFormField>()
    if (proposedType && isLowConfidence(result.type)) lowConfidence.add('type')
    if (proposedAmount && extractedRows.length === 0 && isLowConfidence(result.amount)) lowConfidence.add('amount')
    if (proposedCurrency && supportsProposedCurrency && isLowConfidence(result.currency_code)) lowConfidence.add('currency')
    if (proposedName && isLowConfidence(result.name)) lowConfidence.add('name')
    if (proposedDescription && isLowConfidence(result.description)) lowConfidence.add('description')
    if (validDate && isLowConfidence(result.occurred_on)) lowConfidence.add('date')
    if (validTime && isLowConfidence(result.occurred_at)) lowConfidence.add('time')
    if (categoryHint && isLowConfidence(result.category_hint)) lowConfidence.add('category')
    setLowConfidenceFields(lowConfidence)

    setExtractionSummary({
      attachmentIds: result.attachment_ids ?? [],
      itemCount: extractedRows.length,
      attachmentTotal: result.amount?.value,
      warnings: [
        ...(result.warnings ?? []),
        ...(proposedCurrency && !supportsProposedCurrency
          ? [`Currency ${proposedCurrency} isn't configured — using ${currency?.code ?? 'your base currency'}.`]
          : []),
      ],
    })
    setExtractionStatus('success')
  }

  /**
   * The one extraction path, shared by the initial read and re-extraction. Every ready image is
   * sent in one request and the combined result is written straight into the draft.
   */
  function runPrefill(targets: TransactionAttachmentDraft[]) {
    const attachmentIds = targets.map((target) => target.attachmentId).filter((id): id is string => Boolean(id))
    if (attachmentIds.length === 0) return
    const controller = new AbortController()
    abortRef.current.set('extraction', controller)
    setExtractionStatus('reading')
    setExtractionSummary(undefined)
    setLowConfidenceFields(new Set())
    setLowConfidenceSubTransactionIds(new Set())
    const targetRows = new Set(targets.map((target) => target.id))
    setAttachments((current) =>
      current.map((attachment) =>
        targetRows.has(attachment.id) ? { ...attachment, status: 'processing' } : attachment,
      ),
    )
    extractPrefill
      .mutateAsync({ attachmentIds, signal: controller.signal })
      .then((result) => {
        abortRef.current.delete('extraction')
        setAttachments((current) =>
          current.map((attachment) => (targetRows.has(attachment.id) ? { ...attachment, status: 'ready' } : attachment)),
        )
        applyPrefill(result)
      })
      .catch(() => {
        abortRef.current.delete('extraction')
        setAttachments((current) =>
          current.map((attachment) => (targetRows.has(attachment.id) ? { ...attachment, status: 'ready' } : attachment)),
        )
        // Cancelling is not a failure: drop back to the idle prompt and keep the files attached.
        setExtractionStatus(controller.signal.aborted ? 'idle' : 'failed')
      })
  }

  function handleScanAttachments() {
    const targets = attachments.filter(
      (attachment) => attachment.isImage && attachment.attachmentId && attachment.status === 'ready',
    )
    if (targets.length > 0 && extractionStatus !== 'reading') runPrefill(targets)
  }

  function handleUndoExtraction() {
    if (!extractionUndo) return
    setType(extractionUndo.type)
    setAmount(extractionUndo.amount)
    setName(extractionUndo.name)
    setDescription(extractionUndo.description)
    setCategoryId(extractionUndo.categoryId)
    setDate(extractionUndo.date)
    setTime(extractionUndo.time)
    setCurrencyCode(extractionUndo.currencyCode)
    setSubTransactionItems(extractionUndo.subTransactionItems.map((item) => ({ ...item })))
    setExtractionUndo(null)
    setExtractionSummary(undefined)
    setExtractionStatus('idle')
    setPendingCategoryHint(null)
    setLowConfidenceFields(new Set())
    setLowConfidenceSubTransactionIds(new Set())
  }

  function handlePickAttachment(file: File) {
    setAttachmentPickerOpen(false)
    const id = `row-${nextAttachmentKeyRef.current++}`
    const isImage = file.type.startsWith('image/')
    setAttachments((current) => [
      ...current,
      {
        id,
        name: file.name,
        status: 'uploading',
        isImage,
        mediaType: file.type || 'application/octet-stream',
        byteSize: file.size,
        previewUrl: isImage ? URL.createObjectURL(file) : undefined,
      },
    ])
    // Upload immediately as a draft (transaction_id null); normalization (HEIC/HEIF -> PNG) is
    // entirely server-side. A multi-select uses one mutateAsync promise per resulting file.
    uploadAttachment
      .mutateAsync(file)
      .then((attachment) => {
        pendingDraftIdsRef.current.add(attachment.id)
        updateAttachment(id, {
          attachmentId: attachment.id,
          status: 'ready',
          name: attachment.original_filename,
          mediaType: attachment.media_type,
          byteSize: attachment.byte_size,
        })
      })
      .catch(() => {
        updateAttachment(id, { status: 'error' })
      })
  }

  /** Only uploaded drafts can open the shared lightbox; the row disables its preview until then. */
  function handlePreviewAttachment(id: string) {
    const target = attachments.find((attachment) => attachment.id === id)
    if (!target?.attachmentId) return
    setPreviewAttachmentId(id)
  }

  async function handleDownloadPreview(item: AttachmentPreviewItem) {
    setPreviewDownloading(true)
    try {
      saveBlob(await downloadDraftAttachment(item.id), item.filename)
    } catch {
      // Keep the lightbox open; the row still lets the user retry the download.
    } finally {
      setPreviewDownloading(false)
    }
  }

  function handleCancelProcessing() {
    abortRef.current.get('extraction')?.abort()
  }

  function handleAddSubTransaction(input: SubTransactionListInput) {
    setSubTransactionItems((current) => [...current, { id: `item-${nextItemKeyRef.current++}`, ...input, origin: 'manual' }])
    clearLowConfidence('amount')
  }

  function handleUpdateSubTransaction(id: string, input: SubTransactionListInput) {
    setSubTransactionItems((current) => current.map((item) => (item.id === id ? { ...item, ...input } : item)))
    clearLowConfidence('amount')
    clearLowConfidenceSubTransaction(id)
  }

  function handleSaveSubTransaction(input: SubTransactionListInput) {
    if (!subTransactionEditor) return
    if (subTransactionEditor.mode === 'edit') handleUpdateSubTransaction(subTransactionEditor.id, input)
    else handleAddSubTransaction(input)
    setSubTransactionEditor(null)
  }

  function handleOpenAddSubTransaction() {
    setSubTransactionsExpanded(true)
    setSubTransactionEditor({ mode: 'create' })
  }

  function handleOpenEditSubTransaction(id: string) {
    setSubTransactionEditor({ mode: 'edit', id })
  }

  function handleRemoveSubTransaction(id: string) {
    setSubTransactionItems((current) => current.filter((item) => item.id !== id))
    clearLowConfidence('amount')
    clearLowConfidenceSubTransaction(id)
  }

  const effectiveCurrencyCode = currencyCode || currency?.code || ''
  // While any item exists the transaction amount is derived: Σ(items), and the Amount field is
  // read-only. Emptying the list releases it back to being typed.
  const hasSubTransactions = subTransactionItems.length > 0
  const subTransactionTotalCents = subTransactionItems.reduce((sum, item) => sum + item.amount, 0)
  const effectiveAmount = hasSubTransactions ? decimalStringToAmountDisplay(subTransactionTotalCents / 100) : amount
  // The input shows a comma and its decimals; the API always receives whole minor units (cents).
  const amountCents = hasSubTransactions ? subTransactionTotalCents : amountToCents(amount)
  // Every draft the user attached, once its upload finished. Sent together on create.
  const attachmentIds = attachments.map((a) => a.attachmentId).filter((id): id is string => Boolean(id))
  // Save is blocked while any attachment is still uploading or being read.
  const attachmentsBusy = attachments.some((a) => a.status === 'uploading' || a.status === 'processing')
  const canScanAttachments =
    extractionStatus !== 'reading' &&
    !attachments.some((a) => a.status === 'uploading') &&
    attachments.some((a) => a.isImage && a.status === 'ready' && Boolean(a.attachmentId))
  const canSave =
    amountCents > 0 &&
    name.trim() !== '' &&
    effectiveCategoryId !== null &&
    user !== null &&
    effectiveCurrencyCode !== '' &&
    !attachmentsBusy

  function handleSave() {
    if (!canSave || !user || effectiveCategoryId === null || !effectiveCurrencyCode) return
    createTransaction.mutate(
      {
        type,
        amount: amountCents,
        currency_code: effectiveCurrencyCode,
        category_id: effectiveCategoryId,
        user_id: userId || user.id,
        source: 'manual',
        name: name.trim(),
        description,
        created_at: toISODateTime(date, time),
        attachment_ids: attachmentIds.length > 0 ? attachmentIds : undefined,
        // The server derives the transaction amount from these (ignoring `amount`) whenever present.
        sub_transactions: hasSubTransactions
          ? subTransactionItems.map((item) => ({ name: item.name, label: item.label, amount: item.amount }))
          : undefined,
      },
      {
        onSuccess: () => {
          // The server linked every draft; stop the unmount cleanup from deleting them.
          pendingDraftIdsRef.current.clear()
          navigate('/')
        },
      },
    )
  }

  const dateLabel = date === today ? 'Today' : formatShortDate(date)
  const subTransactionItemsForView: SubTransactionDto[] = subTransactionItems.map((item) => ({
    id: item.id,
    name: item.name,
    label: item.label,
    amount: item.amount,
  }))
  const editingSubTransaction =
    subTransactionEditor?.mode === 'edit'
      ? subTransactionItemsForView.find((item) => item.id === subTransactionEditor.id)
      : undefined
  const previewTarget = attachments.find((attachment) => attachment.id === previewAttachmentId)
  const previewItem: AttachmentPreviewItem | null =
    previewTarget?.attachmentId
      ? {
          id: previewTarget.attachmentId,
          filename: previewTarget.name,
          mediaType: previewTarget.mediaType ?? (previewTarget.isImage ? 'image/jpeg' : 'application/octet-stream'),
          byteSize: previewTarget.byteSize ?? 0,
          previewUrl: previewTarget.previewUrl ?? null,
        }
      : null

  const props: AddTransactionViewProps = {
    type,
    onSelectType: handleSelectType,
    amount: effectiveAmount,
    onAmountChange: handleAmountChange,
    amountReadOnly: hasSubTransactions,
    amountNote: hasSubTransactions
      ? `= sum of ${subTransactionItems.length} ${subTransactionItems.length === 1 ? 'item' : 'items'}`
      : undefined,
    currencyCode: effectiveCurrencyCode || undefined,
    currencies,
    onCurrencyChange: handleCurrencyChange,
    hasCurrency: currencies.length > 0,
    categories,
    categoryId: effectiveCategoryId,
    onSelectCategory: handleSelectCategory,
    name,
    onNameChange: handleNameChange,
    description,
    onDescriptionChange: handleDescriptionChange,
    date,
    time,
    onDateChange: handleDateChange,
    onTimeChange: handleTimeChange,
    onUseToday: handleUseToday,
    dateLabel,
    dateOpen,
    onToggleDate: () => setDateOpen((v) => !v),
    attachments,
    extractionStatus,
    extractionSummary: extractionSummary
      ? { ...extractionSummary, categoryHint: unresolvedCategoryHint }
      : undefined,
    lowConfidenceFields,
    canScanAttachments,
    canUndoExtraction: extractionUndo !== null && extractionStatus !== 'reading',
    attachmentPickerOpen,
    onOpenAttachment: () => setAttachmentPickerOpen(true),
    onCloseAttachment: () => setAttachmentPickerOpen(false),
    onPickAttachment: handlePickAttachment,
    onRemoveAttachment: handleRemoveAttachment,
    onPreviewAttachment: handlePreviewAttachment,
    onScanAttachments: handleScanAttachments,
    onCancelExtraction: handleCancelProcessing,
    onUndoExtraction: handleUndoExtraction,
    subTransactionItems: subTransactionItemsForView,
    lowConfidenceSubTransactionIds,
    subTransactionsExpanded,
    onToggleSubTransactions: () => setSubTransactionsExpanded((value) => !value),
    onReviewSubTransactions: () => setSubTransactionsExpanded(true),
    onAddSubTransaction: handleOpenAddSubTransaction,
    onEditSubTransaction: handleOpenEditSubTransaction,
    onRemoveSubTransaction: handleRemoveSubTransaction,
    users,
    userId,
    onUserChange: setUserOverride,
    showUserPicker: users.length > 1,
    moreOpen,
    onToggleMore: () => setMoreOpen((v) => !v),
    canSave,
    saving: createTransaction.isPending,
    saveError: createTransaction.isError,
    onSave: handleSave,
    onClose: () => navigate(-1),
  }

  return (
    <>
      <ResponsiveView mobile={Mobile} desktop={Desktop} props={props} />
      {previewItem && previewTarget && (
        <AttachmentPreview
          key={previewTarget.id}
          attachment={previewItem}
          isDownloading={previewDownloading}
          onDownload={handleDownloadPreview}
          loadBlob={downloadDraftAttachment}
          onClose={() => setPreviewAttachmentId(null)}
        />
      )}
      {(subTransactionEditor?.mode === 'create' || editingSubTransaction) && (
        <SubTransactionModal
          key={subTransactionEditor?.mode === 'edit' ? subTransactionEditor.id : 'new'}
          open
          item={editingSubTransaction}
          disabled={createTransaction.isPending}
          onClose={() => setSubTransactionEditor(null)}
          onSave={handleSaveSubTransaction}
        />
      )}
    </>
  )
}
