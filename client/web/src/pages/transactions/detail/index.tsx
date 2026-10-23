import { lazy, useEffect, useMemo, useRef, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { attachmentPreviewUrl, downloadTransactionAttachment } from '@/api/attachments'
import type { TransactionDetailsDto, TransactionSource, TransactionType } from '@/api/types'
import { AttachmentPreview } from '@/components/AttachmentPreview'
import { ResponsiveView } from '@/components/ResponsiveView'
import { SubTransactionModal } from '@/components/SubTransactionModal'
import { TransactionAttachmentPicker } from '@/components/TransactionAttachmentPicker'
import type { SubTransactionInput } from '@/components/subTransactionInput'
import { useDeleteTransactionAttachment, useUploadTransactionAttachment } from '@/hooks/useAttachments'
import { useCategories } from '@/hooks/useCategories'
import { useDefaultCurrency } from '@/hooks/useDefaultCurrency'
import { useCreateSubTransaction, useDeleteSubTransaction, useUpdateSubTransaction } from '@/hooks/useSubTransactions'
import { useDeleteTransaction, useTransaction, useUpdateTransaction } from '@/hooks/useTransactions'
import { useUsers } from '@/hooks/useUsers'
import { amountToCents, centsToAmountDisplay, decimalStringToAmountDisplay } from '@/lib/amount'
import { formatSignedAmount } from '@/lib/currency'
import { formatTime, toISODateTime, toLocalISODate } from '@/lib/dates'
import { saveBlob } from '@/lib/download'
import { previewKindFor } from '@/lib/previewKind'
import { AttachmentActions } from './components/AttachmentActions'
import { DeleteTransactionAction } from './components/DeleteTransactionAction'
import type {
  TransactionAttachmentItem,
  TransactionDetailEditProps,
  TransactionDetailItem,
  TransactionDetailViewProps,
} from './types'

const Mobile = lazy(() => import('./TransactionDetail.mobile'))
const Desktop = lazy(() => import('./TransactionDetail.desktop'))

const TYPE_LABELS: Record<TransactionType, string> = { income: 'Income', expense: 'Expense' }
const SOURCE_LABELS: Record<TransactionSource, string> = { manual: 'Manual', recurring: 'Recurring' }

/** The edit form's values for a loaded transaction, formatted the way the inputs expect them. */
function editFormValues(source: TransactionDetailsDto) {
  return {
    type: source.type,
    amount: centsToAmountDisplay(source.amount),
    currencyCode: source.currency_code,
    categoryId: source.category_id,
    name: source.name,
    description: source.description,
    date: toLocalISODate(new Date(source.created_at)),
    time: formatTime(source.created_at),
    userId: source.user_id,
  }
}

export default function TransactionDetailPage() {
  const navigate = useNavigate()
  const { id } = useParams()
  const transactionId = Number(id)
  const hasValidId = Number.isInteger(transactionId)

  const transactionQuery = useTransaction(hasValidId ? transactionId : Number.NaN)
  const usersQuery = useUsers()
  const transaction = transactionQuery.data
  const userName = transaction
    ? (usersQuery.data?.find((user) => user.id === transaction.user_id)?.display_name ?? transaction.user_id)
    : ''

  const [downloadingId, setDownloadingId] = useState<string | null>(null)
  const [downloadError, setDownloadError] = useState(false)
  const [previewAttachment, setPreviewAttachment] = useState<TransactionAttachmentItem | null>(null)
  const [actionsTarget, setActionsTarget] = useState<TransactionAttachmentItem | null>(null)
  const [actionsOpen, setActionsOpen] = useState(false)
  const [deletingId, setDeletingId] = useState<string | null>(null)
  const [deleteError, setDeleteError] = useState(false)
  const deleteAttachment = useDeleteTransactionAttachment()
  const uploadAttachment = useUploadTransactionAttachment()
  const [attachmentPickerOpen, setAttachmentPickerOpen] = useState(false)
  const [attachmentUploadsInFlight, setAttachmentUploadsInFlight] = useState(0)
  const [attachmentUploadError, setAttachmentUploadError] = useState(false)
  const deleteTransaction = useDeleteTransaction()
  const [deleteTransactionOpen, setDeleteTransactionOpen] = useState(false)
  const [deleteTransactionFailed, setDeleteTransactionFailed] = useState(false)

  // Edit mode. The form fields are seeded from the fetched transaction when the user enters edit
  // mode (and again on Reset); the read-only screen keeps reading straight from `transaction`.
  const [editing, setEditing] = useState(false)
  const [type, setType] = useState<TransactionType>('expense')
  const [amount, setAmount] = useState('')
  const [currencyCode, setCurrencyCode] = useState('')
  const [categoryId, setCategoryId] = useState<string | null>(null)
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [date, setDate] = useState('')
  const [time, setTime] = useState('')
  const [userOverride, setUserOverride] = useState('')
  const [moreOpen, setMoreOpen] = useState(false)

  const updateTransaction = useUpdateTransaction()

  // Line items on an existing transaction save immediately, one CRUD call per row (analysis §4,
  // UC6). Each mutation invalidates `['transactions']`, so the detail query refetches the server's
  // recomputed parent amount; a failed call surfaces an inline error and never touches the rows.
  const createSubTransaction = useCreateSubTransaction()
  const updateSubTransaction = useUpdateSubTransaction()
  const deleteSubTransaction = useDeleteSubTransaction()
  const [subTransactionError, setSubTransactionError] = useState<string | null>(null)
  // The line-item panel mirrors Add: collapsed by default and expandable on click. The focused
  // modal edits one row at a time, seeded by `subTransactionEditor`.
  const [subTransactionsExpanded, setSubTransactionsExpanded] = useState(false)
  const [subTransactionEditor, setSubTransactionEditor] = useState<
    { mode: 'create' } | { mode: 'edit'; id: string } | null
  >(null)

  // Only fetched once the form is actually open — the read-only screen needs neither.
  const categoriesQuery = useCategories(type, true, editing)
  const categories = categoriesQuery.data ?? []
  const { currency, currencies } = useDefaultCurrency()
  const users = usersQuery.data ?? []

  const items = useMemo<TransactionDetailItem[]>(() => {
    if (!transaction) return []
    const created = new Date(transaction.created_at)
    const dateLabel = `${created.toLocaleDateString('en-GB', { day: 'numeric', month: 'long', year: 'numeric' })} · ${formatTime(transaction.created_at)}`
    return [
      { label: 'Type', value: TYPE_LABELS[transaction.type] },
      { label: 'Category', value: transaction.category_name },
      { label: 'Date', value: dateLabel },
      { label: 'Added by', value: userName },
      { label: 'Source', value: SOURCE_LABELS[transaction.source] },
      { label: 'Currency', value: transaction.currency_code },
      { label: 'Description', value: transaction.description || '—' },
      { label: 'Reference', value: `#${transaction.id}` },
    ]
  }, [transaction, userName])

  // Resolve the grant the server minted for previewable media into a real inline URL. Non-previewable
  // attachments keep a null URL and only offer download.
  const attachments = useMemo<TransactionAttachmentItem[]>(
    () =>
      (transaction?.attachments ?? []).map((attachment) => ({
        id: attachment.id,
        filename: attachment.original_filename,
        mediaType: attachment.media_type,
        byteSize: attachment.byte_size,
        previewUrl: attachment.preview_grant ? attachmentPreviewUrl(attachment.preview_grant) : null,
      })),
    [transaction],
  )

  // The loaded transaction's line items. While any exist the Amount field is server-derived
  // (read-only); the container reads them straight off the fetched record, so a failed row
  // mutation simply leaves the row unchanged.
  const subTransactions = transaction?.sub_transactions ?? []
  const hasSubTransactions = subTransactions.length > 0
  const subTransactionsBusy =
    createSubTransaction.isPending || updateSubTransaction.isPending || deleteSubTransaction.isPending

  // A line-item add/edit/remove makes the server recompute the parent amount, and the detail query
  // refetches it. While items exist the Amount field is read-only and must show that server value;
  // this also carries the server's last computed amount across the moment the final item is removed
  // and the field unlocks ("last computed value remains until edited", analysis §8). It only writes
  // when items exist (or existed a render ago), so it never clobbers a value being typed into the
  // editable field.
  const hadSubTransactions = useRef(false)
  useEffect(() => {
    if (!editing || !transaction) return
    const previouslyHadItems = hadSubTransactions.current
    hadSubTransactions.current = hasSubTransactions
    if (hasSubTransactions || previouslyHadItems) {
      setAmount(centsToAmountDisplay(transaction.amount))
    }
  }, [editing, transaction, hasSubTransactions])

  /** Adds one item: POST, then refetch so the parent's recomputed amount is what the screen shows. */
  async function handleAddSubTransaction(input: SubTransactionInput): Promise<boolean> {
    if (!transaction) return false
    setSubTransactionError(null)
    try {
      await createSubTransaction.mutateAsync({ transactionId: transaction.id, input })
      return true
    } catch {
      setSubTransactionError("Couldn't add that item — try again.")
      return false
    }
  }

  /** Edits one item: PATCH. On failure the row keeps its previous values and an error shows. */
  async function handleUpdateSubTransaction(id: string, input: SubTransactionInput): Promise<boolean> {
    if (!transaction) return false
    setSubTransactionError(null)
    try {
      await updateSubTransaction.mutateAsync({ transactionId: transaction.id, subTransactionId: id, input })
      return true
    } catch {
      setSubTransactionError("Couldn't save that item — try again.")
      return false
    }
  }

  /** Removes one item: DELETE. The row is only dropped once the refetch confirms the server did it. */
  async function handleRemoveSubTransaction(id: string) {
    if (!transaction) return
    setSubTransactionError(null)
    try {
      await deleteSubTransaction.mutateAsync({ transactionId: transaction.id, subTransactionId: id })
    } catch {
      setSubTransactionError("Couldn't remove that item — try again.")
    }
  }

  function handleOpenAddSubTransaction() {
    setSubTransactionsExpanded(true)
    setSubTransactionEditor({ mode: 'create' })
  }

  function handleOpenEditSubTransaction(id: string) {
    setSubTransactionEditor({ mode: 'edit', id })
  }

  /** Saves the modal's values through the matching CRUD call; the modal stays open on failure. */
  async function handleSaveSubTransaction(input: SubTransactionInput) {
    if (!subTransactionEditor) return
    const saved =
      subTransactionEditor.mode === 'edit'
        ? await handleUpdateSubTransaction(subTransactionEditor.id, input)
        : await handleAddSubTransaction(input)
    if (saved) setSubTransactionEditor(null)
  }

  const editingSubTransaction =
    subTransactionEditor?.mode === 'edit'
      ? subTransactions.find((item) => item.id === subTransactionEditor.id)
      : undefined

  async function handleDownload(attachment: TransactionAttachmentItem) {
    if (!transaction) return
    setDownloadingId(attachment.id)
    setDownloadError(false)
    try {
      // The download route is protected, so an <a href> can't reach it — fetch the bytes with the
      // auth header, then hand the blob to a throwaway object-URL anchor to trigger the save.
      const blob = await downloadTransactionAttachment(transaction.id, attachment.id)
      saveBlob(blob, attachment.filename)
    } catch {
      setDownloadError(true)
    } finally {
      setDownloadingId(null)
    }
  }

  function openActions(attachment: TransactionAttachmentItem) {
    setDeleteError(false)
    setActionsTarget(attachment)
    setActionsOpen(true)
  }

  /** Uploads each picked file straight onto this transaction; the detail query refetches the rows. */
  async function handlePickAttachment(file: File) {
    if (!transaction) return
    setAttachmentPickerOpen(false)
    setAttachmentUploadError(false)
    setAttachmentUploadsInFlight((count) => count + 1)
    try {
      await uploadAttachment.mutateAsync({ transactionId: transaction.id, file })
    } catch {
      setAttachmentUploadError(true)
    } finally {
      setAttachmentUploadsInFlight((count) => count - 1)
    }
  }

  async function handleDelete(attachment: TransactionAttachmentItem) {
    if (!transaction) return
    setDeletingId(attachment.id)
    setDeleteError(false)
    try {
      await deleteAttachment.mutateAsync({ transactionId: transaction.id, attachmentId: attachment.id })
      setActionsOpen(false)
      // If this attachment was open in the lightbox, close it too.
      setPreviewAttachment((current) => (current?.id === attachment.id ? null : current))
    } catch {
      setDeleteError(true)
    } finally {
      setDeletingId(null)
    }
  }

  function handleRequestDeleteTransaction() {
    setDeleteTransactionFailed(false)
    setDeleteTransactionOpen(true)
  }

  function handleCloseDeleteTransaction() {
    if (deleteTransaction.isPending) return
    setDeleteTransactionOpen(false)
    setDeleteTransactionFailed(false)
  }

  /** Deletes the transaction; the server cascades to its line items and attachments. */
  async function handleDeleteTransaction() {
    if (!transaction) return
    setDeleteTransactionFailed(false)
    try {
      await deleteTransaction.mutateAsync(transaction.id)
      navigate('/transactions', { replace: true })
    } catch {
      setDeleteTransactionFailed(true)
    }
  }

  /** Writes the fetched transaction into the edit form. Called on entering edit mode and on Reset. */
  function seedEditForm(source: TransactionDetailsDto) {
    const values = editFormValues(source)
    setType(values.type)
    setAmount(values.amount)
    setCurrencyCode(values.currencyCode)
    setCategoryId(values.categoryId)
    setName(values.name)
    setDescription(values.description)
    setDate(values.date)
    setTime(values.time)
    setUserOverride('')
    setMoreOpen(false)
  }

  function handleStartEdit() {
    if (!transaction) return
    seedEditForm(transaction)
    setEditing(true)
  }

  function handleReset() {
    if (!transaction) return
    // Clear any "couldn't save" state along with the fields, so a fresh attempt starts clean.
    updateTransaction.reset()
    seedEditForm(transaction)
  }

  /** Leave edit mode without changing anything (the button reads "Cancel" while the form is clean). */
  function handleCancel() {
    updateTransaction.reset()
    setSubTransactionEditor(null)
    setEditing(false)
  }

  function handleSelectType(next: TransactionType) {
    setType(next)
    // Categories are per-type; a new type invalidates the previous pick, same as on the Add screen.
    setCategoryId(null)
  }

  // The pristine form values, to tell Reset (edited) from Cancel (untouched) and to seed on entry.
  const baseline = useMemo(() => (transaction ? editFormValues(transaction) : null), [transaction])

  const effectiveCurrencyCode = currencyCode || currency?.code || ''
  const userId = userOverride || transaction?.user_id || ''
  // While items exist the transaction amount is derived from them and the field is read-only, so
  // both the shown value and the value compared for dirtiness come from the item sum, not the field
  // state (which is seeded from the server's amount on entry and kept in step by the effect above).
  const subTransactionTotalCents = subTransactions.reduce((sum, item) => sum + item.amount, 0)
  const amountCents = hasSubTransactions ? subTransactionTotalCents : amountToCents(amount)
  const amountDisplay = hasSubTransactions ? decimalStringToAmountDisplay(amountCents / 100) : amount
  const canSave =
    amountCents > 0 &&
    name.trim() !== '' &&
    categoryId !== null &&
    userId !== '' &&
    effectiveCurrencyCode !== '' &&
    date !== ''
  const isDirty =
    baseline !== null &&
    (type !== baseline.type ||
      amountDisplay !== baseline.amount ||
      currencyCode !== baseline.currencyCode ||
      categoryId !== baseline.categoryId ||
      name !== baseline.name ||
      description !== baseline.description ||
      date !== baseline.date ||
      time !== baseline.time ||
      userId !== baseline.userId)

  function handleSave() {
    if (!transaction || !canSave || categoryId === null) return
    updateTransaction.mutate(
      {
        id: transaction.id,
        input: {
          type,
          currency_code: effectiveCurrencyCode,
          category_id: categoryId,
          user_id: userId,
          name: name.trim(),
          description,
          created_at: toISODateTime(date, time),
          // The server derives (and rejects setting) amount while any item exists.
          ...(hasSubTransactions ? {} : { amount: amountCents }),
        },
      },
      {
        // Leave edit mode; the mutation invalidates the transaction query, so the read-only screen
        // re-renders from the server's fresh copy.
        onSuccess: () => setEditing(false),
      },
    )
  }

  const edit: TransactionDetailEditProps = {
    type,
    amount: amountDisplay,
    amountReadOnly: hasSubTransactions,
    amountNote: hasSubTransactions
      ? `= sum of ${subTransactions.length} ${subTransactions.length === 1 ? 'item' : 'items'}`
      : undefined,
    currencyCode: effectiveCurrencyCode || undefined,
    currencies,
    hasCurrency: currencies.length > 0,
    categories,
    categoryId,
    name,
    description,
    date,
    time,
    users,
    userId,
    showUserPicker: users.length > 1,
    moreOpen,
    canSave,
    saving: updateTransaction.isPending,
    saveError: updateTransaction.isError,
    isDirty,
    onSelectType: handleSelectType,
    onAmountChange: setAmount,
    onCurrencyChange: setCurrencyCode,
    onSelectCategory: (id) => setCategoryId(id || null),
    onNameChange: setName,
    onDescriptionChange: setDescription,
    onDateChange: setDate,
    onTimeChange: setTime,
    onUseToday: () => {
      const now = new Date()
      setDate(toLocalISODate(now))
      setTime(formatTime(now.toISOString()))
    },
    onUserChange: setUserOverride,
    onToggleMore: () => setMoreOpen((v) => !v),
    subTransactions,
    subTransactionsExpanded,
    onToggleSubTransactions: () => setSubTransactionsExpanded((value) => !value),
    subTransactionsBusy,
    subTransactionError,
    onAddSubTransaction: handleOpenAddSubTransaction,
    onEditSubTransaction: handleOpenEditSubTransaction,
    onRemoveSubTransaction: handleRemoveSubTransaction,
    onSave: handleSave,
    onReset: handleReset,
    onCancel: handleCancel,
  }

  const props: TransactionDetailViewProps = {
    title: transaction ? transaction.name || transaction.description || transaction.category_name : '',
    categoryName: transaction?.category_name ?? '',
    amount: transaction
      ? formatSignedAmount(transaction.amount / 100, transaction.type, transaction.currency_code)
      : '',
    isIncome: transaction?.type === 'income',
    items,
    subTransactions,
    currencyCode: transaction?.currency_code ?? '',
    subTransactionsExpanded,
    onToggleSubTransactions: () => setSubTransactionsExpanded((value) => !value),
    attachments,
    attachmentsUploading: attachmentUploadsInFlight > 0,
    attachmentUploadError,
    downloadError,
    isLoading: hasValidId && transactionQuery.isPending,
    isError: !hasValidId || transactionQuery.isError,
    isEditing: editing,
    onStartEdit: handleStartEdit,
    onAddAttachment: () => {
      setAttachmentUploadError(false)
      setAttachmentPickerOpen(true)
    },
    onRequestDelete: handleRequestDeleteTransaction,
    edit,
    onPreview: setPreviewAttachment,
    onOpenActions: openActions,
    onClose: () => navigate(-1),
  }

  return (
    <>
      <ResponsiveView mobile={Mobile} desktop={Desktop} props={props} />
      {actionsTarget && (
        <AttachmentActions
          open={actionsOpen}
          attachment={actionsTarget}
          canPreview={previewKindFor(actionsTarget.mediaType) !== null}
          isDownloading={downloadingId === actionsTarget.id}
          isDeleting={deletingId === actionsTarget.id}
          deleteError={deleteError}
          onPreview={() => {
            setActionsOpen(false)
            setPreviewAttachment(actionsTarget)
          }}
          onDownload={() => handleDownload(actionsTarget)}
          onDelete={() => handleDelete(actionsTarget)}
          onClose={() => setActionsOpen(false)}
        />
      )}
      <DeleteTransactionAction
        open={deleteTransactionOpen}
        transactionName={deleteTransactionOpen ? props.title : ''}
        isDeleting={deleteTransaction.isPending}
        error={deleteTransactionFailed}
        onConfirm={handleDeleteTransaction}
        onClose={handleCloseDeleteTransaction}
      />
      <TransactionAttachmentPicker
        open={attachmentPickerOpen}
        onClose={() => setAttachmentPickerOpen(false)}
        onFile={handlePickAttachment}
        description="Take a photo or choose one or more files to attach to this transaction."
      />
      {previewAttachment && transaction && (
        <AttachmentPreview
          key={previewAttachment.id}
          attachment={previewAttachment}
          isDownloading={downloadingId === previewAttachment.id}
          onDownload={handleDownload}
          loadBlob={(attachmentId) => downloadTransactionAttachment(transaction.id, attachmentId)}
          onClose={() => setPreviewAttachment(null)}
        />
      )}
      {(subTransactionEditor?.mode === 'create' || editingSubTransaction) && (
        <SubTransactionModal
          key={subTransactionEditor?.mode === 'edit' ? subTransactionEditor.id : 'new'}
          open
          item={editingSubTransaction}
          disabled={subTransactionsBusy}
          onClose={() => setSubTransactionEditor(null)}
          onSave={handleSaveSubTransaction}
        />
      )}
    </>
  )
}
