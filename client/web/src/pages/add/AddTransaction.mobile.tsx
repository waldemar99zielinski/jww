import { IconClose, IconPaperclip } from '@/components/icons'
import { Button } from '@/components/ui/Button'
import { IconButton } from '@/components/ui/IconButton'
import { SearchableSelect } from '@/components/ui/SearchableSelect'
import { SegmentedControl } from '@/components/ui/SegmentedControl'
import { TextInput } from '@/components/ui/TextInput'
import { toCategoryOptions } from './categoryOptions'
import { AmountHero } from './components/AmountHero'
import { DateTimeFields } from './components/DateTimeFields'
import { MoreOptionsPanel } from './components/MoreOptionsPanel'
import { QuickFieldsRow } from './components/QuickFieldsRow'
import { ReceiptBanner } from './components/ReceiptBanner'
import { ReceiptPicker } from './components/ReceiptPicker'
import type { AddTransactionViewProps } from './types'

const typeOptions = [
  { label: 'expense', value: 'expense' },
  { label: 'income', value: 'income' },
] as const

export default function AddTransactionMobile({
  type,
  onSelectType,
  amount,
  onAmountChange,
  currencyCode,
  currencies,
  onCurrencyChange,
  hasCurrency,
  categories,
  categoryId,
  onSelectCategory,
  name,
  onNameChange,
  description,
  onDescriptionChange,
  date,
  time,
  onDateChange,
  onTimeChange,
  onUseToday,
  dateLabel,
  dateOpen,
  onToggleDate,
  receipt,
  receiptPickerOpen,
  onOpenReceipt,
  onCloseReceipt,
  onPickReceipt,
  onRemoveReceipt,
  users,
  userId,
  onUserChange,
  showUserPicker,
  moreOpen,
  onToggleMore,
  canSave,
  saving,
  saveError,
  onSave,
  onClose,
}: AddTransactionViewProps) {
  return (
    <div className="flex flex-col gap-5 px-4 pb-4 pt-6">
      <div className="flex items-center justify-between">
        <h2 className="m-0 text-[19px] font-bold">Add transaction</h2>
        <div className="flex items-center gap-2">
          <IconButton
            aria-label="Add attachment"
            onClick={onOpenReceipt}
            icon={<IconPaperclip className="h-4 w-4" />}
            className="h-[34px] w-[34px]"
          />
          <IconButton aria-label="Close" onClick={onClose} icon={<IconClose className="h-4 w-4" />} className="h-[34px] w-[34px]" />
        </div>
      </div>

      <SegmentedControl value={type} options={typeOptions} onChange={onSelectType} />

      <AmountHero
        amount={amount}
        currencyCode={currencyCode}
        currencies={currencies}
        onCurrencyChange={onCurrencyChange}
        onAmountChange={onAmountChange}
        fullWidth
      />

      {receipt && <ReceiptBanner receipt={receipt} onRemove={onRemoveReceipt} />}

      {!hasCurrency && (
        <p className="m-0 text-center text-[11.5px] text-muted">
          No currency configured yet — add one in Settings before saving.
        </p>
      )}

      <TextInput
        type="text"
        value={name}
        onChange={(e) => onNameChange(e.target.value)}
        placeholder="Name (e.g. Lidl groceries)"
        aria-label="Transaction name"
        className="bg-surface sm:text-[13px]"
      />

      <div className="flex flex-col gap-2">
        <p className="m-0 text-xs font-semibold uppercase tracking-wide text-muted">Category</p>
        <SearchableSelect
          value={categoryId ?? ''}
          onChange={onSelectCategory}
          options={toCategoryOptions(categories)}
          placeholder="Select category"
          searchPlaceholder="Search categories"
          emptyMessage="No categories found"
          inline
        />
      </div>

      <TextInput
        type="text"
        value={description}
        onChange={(e) => onDescriptionChange(e.target.value)}
        placeholder="Description (optional)"
        aria-label="Transaction description"
        className="bg-surface sm:text-[13px]"
      />

      <QuickFieldsRow
        dateLabel={dateLabel}
        dateOpen={dateOpen}
        onToggleDate={onToggleDate}
        hasReceipt={receipt !== null}
        onOpenReceipt={onOpenReceipt}
      />

      {dateOpen && (
        <DateTimeFields date={date} time={time} onDateChange={onDateChange} onTimeChange={onTimeChange} onUseToday={onUseToday} />
      )}

      <MoreOptionsPanel
        open={moreOpen}
        onToggle={onToggleMore}
        users={users}
        userId={userId}
        onUserChange={onUserChange}
        showUserPicker={showUserPicker}
      />

      <p className="m-0 text-center text-[11.5px] leading-relaxed text-muted">
        Photo attachments are previews — only the core transaction is saved today.
      </p>

      <div className="sticky bottom-[calc(72px+env(safe-area-inset-bottom))] z-10 -mx-4 mt-1 bg-paper px-4 pb-3 pt-3">
        <Button
          disabled={!canSave || saving}
          onClick={onSave}
          variant="primary"
          className="w-full rounded-2xl bg-accent py-3.5 text-[14.5px] font-bold text-on-accent shadow-[var(--shadow-sm)] disabled:opacity-40"
        >
          {saving ? 'Saving…' : 'Save transaction'}
        </Button>
        {saveError && <p className="m-0 mt-2 text-center text-[11.5px] text-critical">Couldn't save — try again.</p>}
      </div>

      <ReceiptPicker open={receiptPickerOpen} onClose={onCloseReceipt} onFile={onPickReceipt} />
    </div>
  )
}
