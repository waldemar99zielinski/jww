import { DateTimeFields } from '@/components/DateTimeFields'
import { TransactionFormFields } from '@/components/TransactionFormFields'
import { Button } from '@/components/ui/Button'
import { DateQuickField } from './DateQuickField'
import { SubTransactionsPanel } from './SubTransactionsPanel'
import type { AddTransactionViewProps } from '../types'

type AddTransactionFormFields = Pick<
  AddTransactionViewProps,
  | 'type'
  | 'onSelectType'
  | 'amount'
  | 'onAmountChange'
  | 'amountReadOnly'
  | 'amountNote'
  | 'currencyCode'
  | 'currencies'
  | 'onCurrencyChange'
  | 'hasCurrency'
  | 'categories'
  | 'categoryId'
  | 'onSelectCategory'
  | 'name'
  | 'onNameChange'
  | 'description'
  | 'onDescriptionChange'
  | 'date'
  | 'time'
  | 'onDateChange'
  | 'onTimeChange'
  | 'onUseToday'
  | 'dateLabel'
  | 'dateOpen'
  | 'onToggleDate'
  | 'extractionSummary'
  | 'subTransactionItems'
  | 'lowConfidenceSubTransactionIds'
  | 'subTransactionsExpanded'
  | 'onToggleSubTransactions'
  | 'onAddSubTransaction'
  | 'onEditSubTransaction'
  | 'onRemoveSubTransaction'
  | 'users'
  | 'userId'
  | 'onUserChange'
  | 'showUserPicker'
  | 'moreOpen'
  | 'onToggleMore'
  | 'lowConfidenceFields'
  | 'canSave'
  | 'saving'
  | 'saveError'
  | 'onSave'
>

export interface AddTransactionFormProps extends AddTransactionFormFields {
  layout: 'desktop' | 'mobile'
}

/**
 * The Add screen's form body. Keeps the two layout variants from drifting: the shared core fields
 * live in `TransactionFormFields`, and this component owns the Add-only date shortcut, line-item
 * panel and save action. Only page chrome (header, attachment bar, picker) stays in the variants.
 */
export function AddTransactionForm({
  layout,
  type,
  onSelectType,
  amount,
  onAmountChange,
  amountReadOnly,
  amountNote,
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
  extractionSummary,
  subTransactionItems,
  lowConfidenceSubTransactionIds,
  subTransactionsExpanded,
  onToggleSubTransactions,
  onAddSubTransaction,
  onEditSubTransaction,
  onRemoveSubTransaction,
  users,
  userId,
  onUserChange,
  showUserPicker,
  moreOpen,
  onToggleMore,
  lowConfidenceFields,
  canSave,
  saving,
  saveError,
  onSave,
}: AddTransactionFormProps) {
  const saveButton = (
    <Button
      disabled={!canSave || saving}
      onClick={onSave}
      variant="primary"
      className={`rounded-2xl bg-accent py-3.5 text-[14.5px] font-bold text-on-accent shadow-[var(--shadow-sm)] disabled:opacity-40 ${
        layout === 'mobile' ? 'w-full' : ''
      }`}
    >
      {saving ? 'Saving…' : 'Save transaction'}
    </Button>
  )

  const saveErrorLine = saveError && (
    <p className="m-0 text-center text-[11.5px] text-critical">Couldn't save — try again.</p>
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
      lowConfidenceFields={lowConfidenceFields}
      dateField={
        <>
          <DateQuickField
            dateLabel={dateLabel}
            dateOpen={dateOpen}
            onToggleDate={onToggleDate}
            invalid={!dateOpen && (lowConfidenceFields?.has('date') ?? false)}
          />
          {dateOpen && (
            <DateTimeFields
              date={date}
              time={time}
              onDateChange={onDateChange}
              onTimeChange={onTimeChange}
              onUseToday={onUseToday}
              invalidDate={lowConfidenceFields?.has('date') ?? false}
              invalidTime={lowConfidenceFields?.has('time') ?? false}
            />
          )}
        </>
      }
      itemsField={
        <SubTransactionsPanel
          items={subTransactionItems ?? []}
          attachmentTotal={extractionSummary?.attachmentTotal}
          lowConfidenceIds={lowConfidenceSubTransactionIds}
          expanded={subTransactionsExpanded}
          disabled={saving}
          onToggle={onToggleSubTransactions}
          onAdd={onAddSubTransaction ?? (() => {})}
          onEdit={onEditSubTransaction ?? (() => {})}
          onRemove={onRemoveSubTransaction ?? (() => {})}
        />
      }
      actions={
        layout === 'mobile' ? (
          <div className="sticky bottom-[calc(72px+env(safe-area-inset-bottom))] z-10 -mx-4 mt-1 bg-paper px-4 pb-3 pt-3">
            {saveButton}
            {saveErrorLine && <div className="mt-2">{saveErrorLine}</div>}
          </div>
        ) : (
          <>
            {saveButton}
            {saveErrorLine}
          </>
        )
      }
    />
  )
}
