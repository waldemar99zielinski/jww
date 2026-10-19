import type { CurrencyDto } from '@/api/types'
import { IconChevronDown } from '@/components/icons'
import { InlineSelect } from '@/components/ui/InlineSelect'
import { TextInput } from '@/components/ui/TextInput'
import { normalizeAmountInput, sanitizeAmountInput } from '@/lib/amount'

export interface AmountHeroProps {
  amount: string
  currencyCode: string | undefined
  onAmountChange: (value: string) => void
  currencies?: CurrencyDto[]
  onCurrencyChange?: (code: string) => void
  boxed?: boolean
  fullWidth?: boolean
  /**
   * Renders the amount as plain text instead of an input — for the derived state when the value
   * comes from a sum (see `note`) rather than from typing. Additive: omit it and the component
   * behaves exactly as before.
   */
  readOnly?: boolean
  /** Optional hint under the amount, e.g. "= sum of 3 items". Rendered as-is; ignored when empty. */
  note?: string
  /** Review outline for a low-confidence extracted amount. */
  invalid?: boolean
  /** Review outline for a low-confidence extracted currency. */
  currencyInvalid?: boolean
}

export function AmountHero({
  amount,
  currencyCode,
  onAmountChange,
  currencies = [],
  onCurrencyChange,
  boxed,
  fullWidth,
  readOnly = false,
  note,
  invalid = false,
  currencyInvalid = false,
}: AmountHeroProps) {
  const selectable = currencies.length > 1 && onCurrencyChange !== undefined
  const reviewRing = 'rounded-lg ring-2 ring-critical'

  const currency = selectable ? (
    <InlineSelect
      value={currencyCode}
      onChange={(e) => onCurrencyChange(e.target.value)}
      aria-label="Currency"
      aria-invalid={currencyInvalid || undefined}
      className={`text-[17px] ${currencyInvalid ? reviewRing : ''}`}
      icon={<IconChevronDown className="h-4 w-4" />}
    >
      {currencies.map((option) => (
        <option key={option.code} value={option.code}>
          {option.code}
        </option>
      ))}
    </InlineSelect>
  ) : (
    <span
      aria-invalid={currencyInvalid || undefined}
      className={`text-[17px] font-semibold text-muted ${currencyInvalid ? reviewRing : ''}`}
    >
      {currencyCode ?? '—'}
    </span>
  )

  const amountClassName = `border-0 bg-transparent p-0 text-center text-[40px] font-semibold tracking-tight sm:text-[40px] ${
    fullWidth ? 'w-full pr-12' : 'w-36 flex-none'
  }`

  // Read-only shows the value as text; editable keeps the exact `TextInput` from before.
  const input = readOnly ? (
    <span className={`${amountClassName} ${invalid ? reviewRing : ''}`}>{amount}</span>
  ) : (
    <TextInput
      type="text"
      inputMode="decimal"
      value={amount}
      onChange={(e) => onAmountChange(sanitizeAmountInput(e.target.value))}
      onBlur={() => {
        // Pad to two decimals once the user is done, e.g. "5,5" -> "5,50".
        const normalized = normalizeAmountInput(amount)
        if (normalized !== amount) onAmountChange(normalized)
      }}
      placeholder="0,00"
      aria-label="Amount"
      autoFocus
      invalid={invalid}
      className={amountClassName}
    />
  )

  if (fullWidth) {
    return (
      <div
        className={`relative flex w-full flex-wrap items-center font-mono tabular-nums ${
          boxed ? 'rounded-xl border border-line bg-surface-sunken px-4 py-4' : ''
        }`}
      >
        {input}
        <div className="absolute right-0 top-1/2 -translate-y-1/2">{currency}</div>
        {note && <span className="w-full pt-1 text-center text-xs font-normal text-muted">{note}</span>}
      </div>
    )
  }

  return (
    <div
      className={`flex flex-wrap items-baseline justify-center gap-x-1.5 gap-y-1 font-mono tabular-nums ${
        boxed ? 'rounded-xl border border-line bg-surface-sunken py-4' : 'py-1.5'
      }`}
    >
      {input}
      {currency}
      {note && <span className="w-full text-center text-xs font-normal text-muted">{note}</span>}
    </div>
  )
}
