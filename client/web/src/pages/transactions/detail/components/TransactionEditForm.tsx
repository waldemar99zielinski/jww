import { AmountHero } from '@/components/AmountHero'
import { toCategoryOptions } from '@/components/categoryOptions'
import { DateTimeFields } from '@/components/DateTimeFields'
import { MoreOptionsPanel } from '@/components/MoreOptionsPanel'
import { Button } from '@/components/ui/Button'
import { SearchableSelect } from '@/components/ui/SearchableSelect'
import { SegmentedControl } from '@/components/ui/SegmentedControl'
import { TextInput } from '@/components/ui/TextInput'
import type { TransactionDetailEditProps } from '../types'

const typeOptions = [
  { label: 'expense', value: 'expense' },
  { label: 'income', value: 'income' },
] as const

export interface TransactionEditFormProps extends TransactionDetailEditProps {
  /** Drives the same desktop/mobile layout split as the Add Transaction screen. */
  layout: 'desktop' | 'mobile'
}

/**
 * The editable form shown in place of the read-only fields on the transaction detail screen.
 * Pure display: every value and handler arrives as a prop from the page container (see
 * `index.tsx`), which is the only place the edit state and the PATCH mutation live.
 */
export function TransactionEditForm({
  layout,
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
  users,
  userId,
  onUserChange,
  showUserPicker,
  moreOpen,
  onToggleMore,
  canSave,
  saving,
  saveError,
  isDirty,
  onSave,
  onReset,
  onCancel,
}: TransactionEditFormProps) {
  const isDesktop = layout === 'desktop'

  const typeField = <SegmentedControl value={type} options={typeOptions} onChange={onSelectType} />

  const amountField = (
    <AmountHero
      amount={amount}
      currencyCode={currencyCode}
      currencies={currencies}
      onCurrencyChange={onCurrencyChange}
      onAmountChange={onAmountChange}
      boxed={isDesktop}
      fullWidth={!isDesktop}
    />
  )

  const noCurrency = !hasCurrency ? (
    <p className="m-0 text-center text-[11.5px] text-muted">
      No currency configured yet — add one in Settings before saving.
    </p>
  ) : null

  const nameField = (
    <TextInput
      type="text"
      value={name}
      onChange={(e) => onNameChange(e.target.value)}
      placeholder="Name (e.g. Lidl groceries)"
      aria-label="Transaction name"
      className="bg-surface sm:text-[13px]"
    />
  )

  const descriptionField = (
    <TextInput
      type="text"
      value={description}
      onChange={(e) => onDescriptionChange(e.target.value)}
      placeholder="Description (optional)"
      aria-label="Transaction description"
      className="bg-surface sm:text-[13px]"
    />
  )

  const categoryField = (
    <div className="flex flex-col gap-2">
      <p className="m-0 text-xs font-semibold uppercase tracking-wide text-muted">Category</p>
      <SearchableSelect
        value={categoryId ?? ''}
        onChange={onSelectCategory}
        options={toCategoryOptions(categories)}
        placeholder="Select category"
        searchPlaceholder="Search categories"
        emptyMessage="No categories found"
        inline={!isDesktop}
      />
    </div>
  )

  const dateFields = (
    <DateTimeFields date={date} time={time} onDateChange={onDateChange} onTimeChange={onTimeChange} onUseToday={onUseToday} />
  )

  const moreFields = (
    <MoreOptionsPanel
      open={moreOpen}
      onToggle={onToggleMore}
      users={users}
      userId={userId}
      onUserChange={onUserChange}
      showUserPicker={showUserPicker}
    />
  )

  const actions = (
    <div className="flex flex-col gap-2">
      <div className="flex gap-2.5">
        <Button
          variant="custom"
          onClick={isDirty ? onReset : onCancel}
          disabled={saving}
          className="flex-1 rounded-2xl border border-line bg-surface-sunken py-3.5 text-[14.5px] font-bold text-ink-soft disabled:opacity-40"
        >
          {isDirty ? 'Reset' : 'Cancel'}
        </Button>
        <Button
          variant="primary"
          onClick={onSave}
          disabled={!canSave || saving}
          className="flex-1 rounded-2xl bg-accent py-3.5 text-[14.5px] font-bold text-on-accent shadow-[var(--shadow-sm)] disabled:opacity-40"
        >
          {saving ? 'Saving…' : 'Save changes'}
        </Button>
      </div>
      {saveError && <p className="m-0 text-center text-[11.5px] text-critical">Couldn't save — try again.</p>}
    </div>
  )

  if (isDesktop) {
    return (
      <div className="flex flex-col gap-6">
        <div className="grid grid-cols-2 gap-8">
          <div className="flex flex-col gap-5">
            {typeField}
            {amountField}
            {noCurrency}
            {nameField}
            {descriptionField}
            {dateFields}
          </div>
          <div className="flex flex-col gap-4">
            {categoryField}
            {moreFields}
          </div>
        </div>
        {actions}
      </div>
    )
  }

  return (
    <div className="flex flex-col gap-5">
      {typeField}
      {amountField}
      {noCurrency}
      {nameField}
      {categoryField}
      {descriptionField}
      {dateFields}
      {moreFields}
      {actions}
    </div>
  )
}
