import { SubTransactionList } from '@/components/SubTransactionList'
import { DateTimeFields } from '@/components/DateTimeFields'
import { TransactionFormFields } from '@/components/TransactionFormFields'
import { Button } from '@/components/ui/Button'
import type { TransactionDetailEditProps } from '../types'

export interface TransactionEditFormProps extends TransactionDetailEditProps {
  /** Drives the same desktop/mobile layout split as the Add Transaction screen. */
  layout: 'desktop' | 'mobile'
}

/**
 * The editable form shown in place of the read-only fields on the transaction detail screen.
 * Pure display: every value and handler arrives as a prop from the page container (see
 * `index.tsx`), which is the only place the edit state and the PATCH mutation live.
 *
 * The core fields come from the same `TransactionFormFields` component the Add screen uses; only
 * this page's line-item controls and Save/Reset actions are supplied as slots.
 */
export function TransactionEditForm({
  layout,
  type,
  onSelectType,
  amount,
  amountReadOnly,
  amountNote,
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
  subTransactions,
  subTransactionsBusy,
  subTransactionError,
  onAddSubTransaction,
  onUpdateSubTransaction,
  onRemoveSubTransaction,
  canSave,
  saving,
  saveError,
  isDirty,
  onSave,
  onReset,
  onCancel,
}: TransactionEditFormProps) {
  // The line items persist per row on this screen (unlike the Add form, where they are local until
  // save): every change goes out as a CRUD call owned by the container, so this stays pure display.
  const itemsField = (
    <div className="flex flex-col gap-2">
      <SubTransactionList
        items={subTransactions ?? []}
        currencyCode={currencyCode}
        onAdd={onAddSubTransaction ?? (() => {})}
        onUpdate={onUpdateSubTransaction ?? (() => {})}
        onRemove={onRemoveSubTransaction ?? (() => {})}
        disabled={saving}
        busy={subTransactionsBusy}
      />
      {subTransactionError && <p className="m-0 text-[11.5px] text-critical">{subTransactionError}</p>}
    </div>
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

  return (
    <TransactionFormFields
      layout={layout}
      type={type}
      onSelectType={onSelectType}
      amount={amount}
      onAmountChange={onAmountChange}
      amountReadOnly={amountReadOnly}
      amountNote={amountNote}
      currencyCode={currencyCode}
      currencies={currencies}
      onCurrencyChange={onCurrencyChange}
      hasCurrency={hasCurrency}
      name={name}
      onNameChange={onNameChange}
      categories={categories}
      categoryId={categoryId}
      onSelectCategory={onSelectCategory}
      description={description}
      onDescriptionChange={onDescriptionChange}
      users={users}
      userId={userId}
      onUserChange={onUserChange}
      showUserPicker={showUserPicker}
      moreOpen={moreOpen}
      onToggleMore={onToggleMore}
      dateField={
        <DateTimeFields
          date={date}
          time={time}
          onDateChange={onDateChange}
          onTimeChange={onTimeChange}
          onUseToday={onUseToday}
        />
      }
      itemsField={itemsField}
      actions={actions}
    />
  )
}
