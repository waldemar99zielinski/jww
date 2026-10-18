import { IconClose, IconPaperclip } from '@/components/icons'
import { Button } from '@/components/ui/Button'
import { IconButton } from '@/components/ui/IconButton'
import { SearchableSelect } from '@/components/ui/SearchableSelect'
import { SegmentedControl } from '@/components/ui/SegmentedControl'
import { TextInput } from '@/components/ui/TextInput'
import { AmountHero } from '@/components/AmountHero'
import { toCategoryOptions } from '@/components/categoryOptions'
import { DateTimeFields } from '@/components/DateTimeFields'
import { MoreOptionsPanel } from '@/components/MoreOptionsPanel'
import { QuickFieldsRow } from './components/QuickFieldsRow'
import { TransactionAttachmentBanner } from './components/TransactionAttachmentBanner'
import { TransactionAttachmentPicker } from './components/TransactionAttachmentPicker'
import type { AddTransactionViewProps } from './types'

const typeOptions = [
  { label: 'expense', value: 'expense' },
  { label: 'income', value: 'income' },
] as const

export default function AddTransactionDesktop({
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
  attachments,
  attachmentPickerOpen,
  onOpenAttachment,
  onCloseAttachment,
  onPickAttachment,
  onRemoveAttachment,
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
    <div className="mx-auto flex max-w-2xl flex-col gap-6 rounded-2xl border border-line bg-surface p-8">
      <div className="flex items-center justify-between">
        <h1 className="m-0 text-2xl font-bold">Add transaction</h1>
        <div className="flex items-center gap-2">
          <IconButton
            aria-label="Add attachment"
            onClick={onOpenAttachment}
            icon={<IconPaperclip className="h-4 w-4" />}
            className="h-9 w-9"
          />
          <IconButton aria-label="Close" onClick={onClose} icon={<IconClose className="h-4 w-4" />} className="h-9 w-9" />
        </div>
      </div>

      {attachments.length > 0 && (
        <div className="flex flex-col gap-2">
          {attachments.map((a) => (
            <TransactionAttachmentBanner key={a.id} attachment={a} onRemove={() => onRemoveAttachment(a.id)} />
          ))}
        </div>
      )}

      <div className="grid grid-cols-2 gap-8">
        <div className="flex flex-col gap-5">
          <SegmentedControl value={type} options={typeOptions} onChange={onSelectType} />

          <AmountHero
            amount={amount}
            currencyCode={currencyCode}
            currencies={currencies}
            onCurrencyChange={onCurrencyChange}
            onAmountChange={onAmountChange}
            boxed
          />

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
            hasAttachment={attachments.length > 0}
            onOpenAttachment={onOpenAttachment}
          />

          {dateOpen && (
            <DateTimeFields date={date} time={time} onDateChange={onDateChange} onTimeChange={onTimeChange} onUseToday={onUseToday} />
          )}
        </div>

        <div className="flex flex-col gap-4">
          <p className="m-0 text-xs font-semibold uppercase tracking-wide text-muted">Category</p>
          <SearchableSelect
            value={categoryId ?? ''}
            onChange={onSelectCategory}
            options={toCategoryOptions(categories)}
            placeholder="Select category"
            searchPlaceholder="Search categories"
            emptyMessage="No categories found"
          />
          <MoreOptionsPanel
            open={moreOpen}
            onToggle={onToggleMore}
            users={users}
            userId={userId}
            onUserChange={onUserChange}
            showUserPicker={showUserPicker}
          />
        </div>
      </div>

      <Button
        disabled={!canSave || saving}
        onClick={onSave}
        variant="primary"
        className="rounded-2xl bg-accent py-3.5 text-[14.5px] font-bold text-on-accent shadow-[var(--shadow-sm)] disabled:opacity-40"
      >
        {saving ? 'Saving…' : 'Save transaction'}
      </Button>
      {saveError && <p className="m-0 text-center text-[11.5px] text-critical">Couldn't save — try again.</p>}

      <TransactionAttachmentPicker open={attachmentPickerOpen} onClose={onCloseAttachment} onFile={onPickAttachment} />
    </div>
  )
}
