import type { ReactNode } from 'react'
import type { CurrencyDto, TransactionCategoryDto, TransactionType, UserDto } from '@/api/types'
import { AmountHero } from './AmountHero'
import { MoreOptionsPanel } from './MoreOptionsPanel'
import { toCategoryOptions } from './categoryOptions'
import { SearchableSelect } from './ui/SearchableSelect'
import { SegmentedControl } from './ui/SegmentedControl'
import { TextInput } from './ui/TextInput'

const typeOptions = [
  { label: 'expense', value: 'expense' },
  { label: 'income', value: 'income' },
] as const

/** The scalar fields an extraction can propose, used to flag low-confidence values for review. */
export type TransactionFormField =
  | 'type'
  | 'amount'
  | 'currency'
  | 'name'
  | 'category'
  | 'description'
  | 'date'
  | 'time'

export type LowConfidenceFields = ReadonlySet<TransactionFormField>

export interface TransactionFormFieldsProps {
  /** Desktop uses the two-column Add/Edit layout; mobile stacks every field. */
  layout: 'desktop' | 'mobile'
  type: TransactionType
  onSelectType: (type: TransactionType) => void
  amount: string
  onAmountChange: (value: string) => void
  amountReadOnly?: boolean
  amountNote?: string
  currencyCode: string | undefined
  currencies: CurrencyDto[]
  onCurrencyChange: (code: string) => void
  hasCurrency: boolean
  name: string
  onNameChange: (value: string) => void
  categories: TransactionCategoryDto[]
  categoryId: string | null
  onSelectCategory: (id: string) => void
  description: string
  onDescriptionChange: (value: string) => void
  users: UserDto[]
  userId: string
  onUserChange: (id: string) => void
  showUserPicker: boolean
  moreOpen: boolean
  onToggleMore: () => void
  /**
   * Fields extracted below the confidence threshold. Only the Add flow passes this; the detail
   * edit form leaves it undefined so editing a saved transaction never shows review outlines.
   */
  lowConfidenceFields?: LowConfidenceFields
  /**
   * The page owns its date affordance: Add shows a quick-date chip that reveals the full fields,
   * while edit mode always shows the date/time inputs.
   */
  dateField: ReactNode
  /** Page-specific line-item surface (collapsible panel on Add, inline list while editing). */
  itemsField?: ReactNode
  /** Page-specific submit/reset controls rendered below the fields. */
  actions?: ReactNode
}

/**
 * The transaction fields shared by the Add screen and the detail screen's edit mode. It owns only
 * the field composition and desktop/mobile arrangement; values, handlers, and the page-specific
 * slots (attachments, line items, save actions) stay in each container.
 */
export function TransactionFormFields({
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
  name,
  onNameChange,
  categories,
  categoryId,
  onSelectCategory,
  description,
  onDescriptionChange,
  users,
  userId,
  onUserChange,
  showUserPicker,
  moreOpen,
  onToggleMore,
  lowConfidenceFields,
  dateField,
  itemsField,
  actions,
}: TransactionFormFieldsProps) {
  const isDesktop = layout === 'desktop'
  const isLow = (field: TransactionFormField) => lowConfidenceFields?.has(field) ?? false

  const typeField = (
    <div className="flex flex-col gap-2">
      <p className="m-0 text-xs font-semibold uppercase tracking-wide text-muted">Transaction type</p>
      <SegmentedControl
        value={type}
        options={typeOptions}
        onChange={onSelectType}
        ariaLabel="Transaction type"
        invalid={isLow('type')}
      />
    </div>
  )

  const amountField = (
    <AmountHero
      amount={amount}
      currencyCode={currencyCode}
      currencies={currencies}
      onCurrencyChange={onCurrencyChange}
      onAmountChange={onAmountChange}
      readOnly={amountReadOnly}
      note={amountNote}
      boxed={isDesktop}
      fullWidth={!isDesktop}
      invalid={isLow('amount')}
      currencyInvalid={isLow('currency')}
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
      invalid={isLow('name')}
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
        invalid={isLow('category')}
      />
    </div>
  )

  const descriptionField = (
    <TextInput
      type="text"
      value={description}
      onChange={(e) => onDescriptionChange(e.target.value)}
      placeholder="Description (optional)"
      aria-label="Transaction description"
      invalid={isLow('description')}
      className="bg-surface sm:text-[13px]"
    />
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
            {dateField}
          </div>
          <div className="flex flex-col gap-4">
            {categoryField}
            {moreFields}
          </div>
        </div>
        {itemsField}
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
      {dateField}
      {moreFields}
      {itemsField}
      {actions}
    </div>
  )
}
