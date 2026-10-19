import { Button } from './Button'

export interface SegmentedOption<T extends string> {
  label: string
  value: T
}

export interface SegmentedControlProps<T extends string> {
  value: T
  options: readonly SegmentedOption<T>[]
  onChange: (value: T) => void
  /** Accessible name for the button group. */
  ariaLabel?: string
  /** Draws the review outline used when an extracted value is below the confidence threshold. */
  invalid?: boolean
  className?: string
}

export function SegmentedControl<T extends string>({
  value,
  options,
  onChange,
  ariaLabel,
  invalid = false,
  className,
}: SegmentedControlProps<T>) {
  return (
    <div
      role="group"
      aria-label={ariaLabel}
      aria-invalid={invalid || undefined}
      className={`${className ?? 'grid grid-cols-2 gap-[3px] rounded-xl bg-surface-sunken p-[3px]'} ${
        invalid ? 'ring-2 ring-critical' : ''
      }`}
    >
      {options.map((option) => (
        <Button
          key={option.value}
          variant="custom"
          onClick={() => onChange(option.value)}
          className={`rounded-[9px] py-2.5 text-[13.5px] font-semibold capitalize ${
            value === option.value ? 'bg-surface text-ink shadow-[var(--shadow-sm)]' : 'text-ink-soft'
          }`}
        >
          {option.label}
        </Button>
      ))}
    </div>
  )
}
