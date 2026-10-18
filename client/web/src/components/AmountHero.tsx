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
}

export function AmountHero({
  amount,
  currencyCode,
  onAmountChange,
  currencies = [],
  onCurrencyChange,
  boxed,
  fullWidth,
}: AmountHeroProps) {
  const selectable = currencies.length > 1 && onCurrencyChange !== undefined

  const currency = selectable ? (
    <InlineSelect
      value={currencyCode}
      onChange={(e) => onCurrencyChange(e.target.value)}
      aria-label="Currency"
      className="text-[17px]"
      icon={<IconChevronDown className="h-4 w-4" />}
    >
      {currencies.map((option) => (
        <option key={option.code} value={option.code}>
          {option.code}
        </option>
      ))}
    </InlineSelect>
  ) : (
    <span className="text-[17px] font-semibold text-muted">{currencyCode ?? '—'}</span>
  )

  const input = (
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
      className={`border-0 bg-transparent p-0 text-center text-[40px] font-semibold tracking-tight sm:text-[40px] ${
        fullWidth ? 'w-full pr-12' : 'w-36 flex-none'
      }`}
    />
  )

  if (fullWidth) {
    return (
      <div
        className={`relative flex w-full items-center font-mono tabular-nums ${
          boxed ? 'rounded-xl border border-line bg-surface-sunken px-4 py-4' : ''
        }`}
      >
        {input}
        <div className="absolute right-0 top-1/2 -translate-y-1/2">{currency}</div>
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
    </div>
  )
}
