import { lazy, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import type { TransactionType } from '@/api/types'
import { ResponsiveView } from '@/components/ResponsiveView'
import { useAuth } from '@/hooks/useAuth'
import { useCategories } from '@/hooks/useCategories'
import { useDefaultCurrency } from '@/hooks/useDefaultCurrency'
import { useCreateTransaction } from '@/hooks/useTransactions'
import { useUsers } from '@/hooks/useUsers'
import { formatShortDate, formatTime, toISODate } from '@/lib/dates'
import type { AddTransactionViewProps, ReceiptDraft } from './types'

const Mobile = lazy(() => import('./AddTransaction.mobile'))
const Desktop = lazy(() => import('./AddTransaction.desktop'))

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
  const [receipt, setReceipt] = useState<ReceiptDraft | null>(null)
  const [receiptPickerOpen, setReceiptPickerOpen] = useState(false)
  const [moreOpen, setMoreOpen] = useState(false)
  const [currencyCode, setCurrencyCode] = useState('')
  const [userOverride, setUserOverride] = useState('')

  const { user } = useAuth()
  const { currency, currencies } = useDefaultCurrency()
  const usersQuery = useUsers()
  const categoriesQuery = useCategories(type)
  const categories = categoriesQuery.data ?? []
  const createTransaction = useCreateTransaction()

  const users = usersQuery.data ?? (user ? [user] : [])
  const userId = userOverride || user?.id || ''

  function handleSelectType(next: TransactionType) {
    setType(next)
    setCategoryId(null)
  }

  function handleUseToday() {
    setDate(toISODate(new Date()))
    setTime(formatTime(new Date().toISOString()))
  }

  function handleRemoveReceipt() {
    setReceipt((current) => {
      if (current?.previewUrl) URL.revokeObjectURL(current.previewUrl)
      return null
    })
  }

  function handlePickReceipt(file: File) {
    setReceiptPickerOpen(false)
    setReceipt((current) => {
      if (current?.previewUrl) URL.revokeObjectURL(current.previewUrl)
      const isImage = file.type.startsWith('image/')
      return { name: file.name, status: 'parsing', isImage, previewUrl: isImage ? URL.createObjectURL(file) : undefined }
    })
    window.setTimeout(() => {
      setReceipt((current) => (current ? { ...current, status: 'ready' } : current))
      setAmount((current) => current || '184,32')
      setName((current) => current || 'Supermarket receipt')
    }, 1200)
  }

  const effectiveCurrencyCode = currencyCode || currency?.code || ''
  const amountValue = Number(amount.replace(',', '.'))
  const canSave = amountValue > 0 && name.trim() !== '' && categoryId !== null && user !== null && effectiveCurrencyCode !== ''

  function handleSave() {
    if (!canSave || !user || categoryId === null || !effectiveCurrencyCode) return
    createTransaction.mutate(
      {
        type,
        amount: amountValue,
        currency_code: effectiveCurrencyCode,
        category_id: categoryId,
        user_id: userId || user.id,
        source: 'manual',
        name: name.trim(),
        description,
      },
      { onSuccess: () => navigate('/') },
    )
  }

  const dateLabel = date === today ? 'Today' : formatShortDate(date)

  const props: AddTransactionViewProps = {
    type,
    onSelectType: handleSelectType,
    amount,
    onAmountChange: setAmount,
    currencyCode: effectiveCurrencyCode || undefined,
    currencies,
    onCurrencyChange: setCurrencyCode,
    hasCurrency: currencies.length > 0,
    categories,
    categoryId,
    onSelectCategory: (id: string) => setCategoryId(id || null),
    name,
    onNameChange: setName,
    description,
    onDescriptionChange: setDescription,
    date,
    time,
    onDateChange: setDate,
    onTimeChange: setTime,
    onUseToday: handleUseToday,
    dateLabel,
    dateOpen,
    onToggleDate: () => setDateOpen((v) => !v),
    receipt,
    receiptPickerOpen,
    onOpenReceipt: () => setReceiptPickerOpen(true),
    onCloseReceipt: () => setReceiptPickerOpen(false),
    onPickReceipt: handlePickReceipt,
    onRemoveReceipt: handleRemoveReceipt,
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

  return <ResponsiveView mobile={Mobile} desktop={Desktop} props={props} />
}
