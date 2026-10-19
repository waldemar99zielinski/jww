import { Button } from './Button'
import { TextInput } from './TextInput'
import { normalizeSignedAmountInput, sanitizeSignedAmountInput } from '@/lib/amount'

export interface SignedAmountInputProps {
  value: string
  onChange: (value: string) => void
  label: string
  placeholder?: string
  className?: string
  autoFocus?: boolean
}

/**
 * A decimal amount input that can be made negative without a minus key. iOS's `inputMode="decimal"`
 * keypad (and many Android numeric keypads) has no minus, so a visible toggle is the only reliable
 * way to enter discount/adjustment line items. The toggle keeps the numeric keyboard.
 */
export function SignedAmountInput({
  value,
  onChange,
  label,
  placeholder = '0,00',
  className,
  autoFocus = false,
}: SignedAmountInputProps) {
  const negative = value.startsWith('-')

  function toggleSign() {
    if (value === '') {
      onChange('-')
      return
    }
    onChange(negative ? value.slice(1) : `-${value}`)
  }

  function handleBlur() {
    const normalized = normalizeSignedAmountInput(value)
    if (normalized !== value) onChange(normalized)
  }

  return (
    <div className="relative block w-full min-w-0">
      <TextInput
        type="text"
        inputMode="decimal"
        value={value}
        onChange={(event) => onChange(sanitizeSignedAmountInput(event.target.value))}
        onBlur={handleBlur}
        placeholder={placeholder}
        aria-label={label}
        autoFocus={autoFocus}
        className={`w-full pl-9 text-right ${className ?? ''}`}
      />
      <Button
        variant="custom"
        type="button"
        aria-label={negative ? 'Make amount positive' : 'Make amount negative'}
        aria-pressed={negative}
        title={negative ? 'Make positive' : 'Make negative'}
        // Keep focus (and the numeric keyboard) on the input when tapping the toggle.
        onPointerDown={(event) => event.preventDefault()}
        onClick={toggleSign}
        className={`absolute left-1 top-1/2 flex h-7 w-7 -translate-y-1/2 items-center justify-center rounded-md text-sm font-bold ${
          negative ? 'bg-critical-soft text-critical' : 'text-muted hover:bg-surface'
        }`}
      >
        ±
      </Button>
    </div>
  )
}
