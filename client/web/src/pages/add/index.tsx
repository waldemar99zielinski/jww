import { lazy, useEffect, useRef, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { deleteDraftAttachment } from '@/api/attachments'
import type { TransactionPrefillDto, TransactionType } from '@/api/types'
import { ResponsiveView } from '@/components/ResponsiveView'
import { useAuth } from '@/hooks/useAuth'
import { useCategories } from '@/hooks/useCategories'
import { useDefaultCurrency } from '@/hooks/useDefaultCurrency'
import { useDeleteDraftAttachment, useExtractTransactionPrefill, useUploadDraftAttachment } from '@/hooks/useAttachments'
import { useCreateTransaction } from '@/hooks/useTransactions'
import { useUsers } from '@/hooks/useUsers'
import { formatShortDate, formatTime, toISODate, toISODateTime } from '@/lib/dates'
import { amountToCents, decimalStringToAmountDisplay } from '@/lib/amount'
import { PrefillProcessing } from './components/PrefillProcessing'
import type { AddTransactionViewProps, TransactionAttachmentDraft, TransactionAttachmentFlow } from './types'

const Mobile = lazy(() => import('./AddTransaction.mobile'))
const Desktop = lazy(() => import('./AddTransaction.desktop'))

/** Form fields a read can fill; `touchedRef` records the ones the user edited so we never overwrite them. */
type PrefillFieldKey = 'type' | 'amount' | 'name' | 'description' | 'category' | 'currency' | 'date'

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
  const [moreOpen, setMoreOpen] = useState(false)
  const [currencyCode, setCurrencyCode] = useState('')
  const [userOverride, setUserOverride] = useState('')

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
  /** Fields the user edited; filled only by the change handlers so a read never clobbers typed input. */
  const touchedRef = useRef<Set<PrefillFieldKey>>(new Set())
  /** Cancels each row's in-flight read when the user cancels processing or leaves the screen. */
  const abortRef = useRef<Map<string, AbortController>>(new Map())

  // Any draft the user uploaded but left without saving is deleted on unmount so its bytes don't
  // linger in object storage. Called directly (not through the hook) because this is a one-shot
  // cleanup, not a rendered mutation. `deleteDraftAttachment` is a module import, so the empty
  // dep list is correct. `pendingDraftIdsRef.current` is a Set created once and mutated in place,
  // so reading it here (rather than inside the cleanup) is safe.
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

  function handleSelectType(next: TransactionType) {
    touchedRef.current.add('type')
    setType(next)
    setCategoryId(null)
  }

  function handleAmountChange(value: string) {
    touchedRef.current.add('amount')
    setAmount(value)
  }

  function handleNameChange(value: string) {
    touchedRef.current.add('name')
    setName(value)
  }

  function handleDescriptionChange(value: string) {
    touchedRef.current.add('description')
    setDescription(value)
  }

  function handleSelectCategory(id: string) {
    touchedRef.current.add('category')
    setCategoryId(id || null)
  }

  function handleDateChange(value: string) {
    touchedRef.current.add('date')
    setDate(value)
  }

  function handleUseToday() {
    touchedRef.current.add('date')
    setDate(toISODate(new Date()))
    setTime(formatTime(new Date().toISOString()))
  }

  /** Merges a patch into the one attached file the patch targets, leaving its siblings untouched. */
  function updateAttachment(id: string, patch: Partial<TransactionAttachmentDraft>) {
    setAttachments((current) => current.map((a) => (a.id === id ? { ...a, ...patch } : a)))
  }

  function handleRemoveAttachment(id: string) {
    const target = attachments.find((a) => a.id === id)
    if (!target) return
    const controller = abortRef.current.get(id)
    if (controller) {
      controller.abort()
      abortRef.current.delete(id)
    }
    if (target.attachmentId) {
      pendingDraftIdsRef.current.delete(target.attachmentId)
      deleteAttachment.mutate(target.attachmentId)
    }
    if (target.previewUrl) URL.revokeObjectURL(target.previewUrl)
    setAttachments((current) => current.filter((a) => a.id !== id))
  }

  /** Writes only the fields the user has not edited since opening the screen. */
  function applyPrefill(result: TransactionPrefillDto, id: string) {
    const touched = touchedRef.current
    const filled: string[] = []

    const nextType = result.type?.value
    const typeChanged = nextType !== undefined && nextType !== type && !touched.has('type')
    if (typeChanged) {
      setType(nextType)
      setCategoryId(null)
      filled.push('type')
    }
    const targetType = typeChanged ? nextType : type

    if (result.amount && !touched.has('amount')) {
      // The reader returns a decimal string ("24.90"); the input works in display form ("24,90").
      setAmount(decimalStringToAmountDisplay(result.amount.value))
      filled.push('amount')
    }
    if (result.name && !touched.has('name')) {
      setName(result.name.value)
      filled.push('name')
    }
    if (result.description && !touched.has('description')) {
      setDescription(result.description.value)
      filled.push('description')
    }
    if (result.currency_code && !touched.has('currency') && currencies.some((c) => c.code === result.currency_code!.value)) {
      setCurrencyCode(result.currency_code.value)
      filled.push('currency')
    }
    if (result.occurred_on && !touched.has('date')) {
      setDate(result.occurred_on.value)
      filled.push('date')
    }
    // The hint is a category *name*; only match it when the loaded list is for the target type.
    if (result.category_hint && !touched.has('category') && targetType === type) {
      const hint = result.category_hint.value.toLowerCase()
      const match = categories.find((c) => c.name.toLowerCase() === hint)
      if (match) {
        setCategoryId(match.id)
        filled.push('category')
      }
    }

    updateAttachment(id, { prefilledCount: filled.length })
  }

  function runPrefill(attachmentId: string, id: string) {
    const controller = new AbortController()
    abortRef.current.set(id, controller)
    // `mutateAsync`, not `mutate` with per-call callbacks: several reads can be in flight at once,
    // and a shared mutation observer only keeps the latest call's options (callbacks would be lost).
    extractPrefill
      .mutateAsync({ attachmentId, signal: controller.signal })
      .then((result) => {
        abortRef.current.delete(id)
        applyPrefill(result, id)
        updateAttachment(id, { status: 'ready' })
      })
      .catch(() => {
        abortRef.current.delete(id)
        // A user-triggered cancel aborts the signal; only a real failure gets the error copy.
        const cancelled = controller.signal.aborted
        updateAttachment(id, { status: 'ready', readFailed: !cancelled })
      })
  }

  function handlePickAttachment(file: File, flow: TransactionAttachmentFlow) {
    setAttachmentPickerOpen(false)
    const id = `row-${nextAttachmentKeyRef.current++}`
    const isImage = file.type.startsWith('image/')
    setAttachments((current) => [
      ...current,
      { id, name: file.name, status: 'uploading', flow, isImage, previewUrl: isImage ? URL.createObjectURL(file) : undefined },
    ])
    // Upload immediately as a draft (transaction_id null); the id is linked on save. The file is
    // streamed to the server, not read/parsed here. A multi-select uploads every file in the same
    // tick, so each uses its own `mutateAsync` promise — per-call `onSuccess` on the shared
    // mutation observer would keep only the last file's callback and strand the earlier rows.
    uploadAttachment
      .mutateAsync(file)
      .then((attachment) => {
        pendingDraftIdsRef.current.add(attachment.id)
        updateAttachment(id, { attachmentId: attachment.id, status: flow === 'prefill' ? 'processing' : 'ready' })
        if (flow === 'prefill') runPrefill(attachment.id, id)
      })
      .catch(() => {
        updateAttachment(id, { status: 'error' })
      })
  }

  function handleCancelProcessing(id: string) {
    abortRef.current.get(id)?.abort()
  }

  const effectiveCurrencyCode = currencyCode || currency?.code || ''
  // The input shows a comma and its decimals; the API always receives whole minor units (cents).
  const amountCents = amountToCents(amount)
  // Every draft the user attached, once its upload finished. Sent together on create.
  const attachmentIds = attachments.map((a) => a.attachmentId).filter((id): id is string => Boolean(id))
  // Save is blocked while any attachment is still uploading or being read.
  const attachmentsBusy = attachments.some((a) => a.status === 'uploading' || a.status === 'processing')
  const canSave =
    amountCents > 0 &&
    name.trim() !== '' &&
    categoryId !== null &&
    user !== null &&
    effectiveCurrencyCode !== '' &&
    !attachmentsBusy

  function handleSave() {
    if (!canSave || !user || categoryId === null || !effectiveCurrencyCode) return
    createTransaction.mutate(
      {
        type,
        amount: amountCents,
        currency_code: effectiveCurrencyCode,
        category_id: categoryId,
        user_id: userId || user.id,
        source: 'manual',
        name: name.trim(),
        description,
        created_at: toISODateTime(date, time),
        attachment_ids: attachmentIds.length > 0 ? attachmentIds : undefined,
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
  const processingAttachment = attachments.find((a) => a.status === 'processing')

  const props: AddTransactionViewProps = {
    type,
    onSelectType: handleSelectType,
    amount,
    onAmountChange: handleAmountChange,
    currencyCode: effectiveCurrencyCode || undefined,
    currencies,
    onCurrencyChange: setCurrencyCode,
    hasCurrency: currencies.length > 0,
    categories,
    categoryId,
    onSelectCategory: handleSelectCategory,
    name,
    onNameChange: handleNameChange,
    description,
    onDescriptionChange: handleDescriptionChange,
    date,
    time,
    onDateChange: handleDateChange,
    onTimeChange: setTime,
    onUseToday: handleUseToday,
    dateLabel,
    dateOpen,
    onToggleDate: () => setDateOpen((v) => !v),
    attachments,
    attachmentPickerOpen,
    onOpenAttachment: () => setAttachmentPickerOpen(true),
    onCloseAttachment: () => setAttachmentPickerOpen(false),
    onPickAttachment: handlePickAttachment,
    onRemoveAttachment: handleRemoveAttachment,
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
      {processingAttachment && (
        <PrefillProcessing fileName={processingAttachment.name} onCancel={() => handleCancelProcessing(processingAttachment.id)} />
      )}
    </>
  )
}
