import { useRef, type MouseEvent } from 'react'
import { IconCalendar } from '@/components/icons'
import { FieldLabel } from '@/components/ui/FieldLabel'
import { TextInput } from '@/components/ui/TextInput'

export interface DateInputProps {
  label: string
  value: string
  min?: string
  max?: string
  invalid?: boolean
  onChange: (date: string) => void
  className?: string
}

/** A date TextInput that opens the native picker on any click within the field, not just the icon. */
export function DateInput({ label, value, min, max, invalid, onChange, className }: DateInputProps) {
  const inputRef = useRef<HTMLInputElement>(null)

  function openPicker() {
    inputRef.current?.focus()
    inputRef.current?.showPicker?.()
  }

  function handleFieldClick(event: MouseEvent<HTMLSpanElement>) {
    const input = inputRef.current
    if (input && event.target === input) {
      const rect = input.getBoundingClientRect()
      const insideField =
        event.clientX >= rect.left && event.clientX <= rect.right && event.clientY >= rect.top && event.clientY <= rect.bottom
      // Clicks coming from the native picker popup (e.g. its "Clear" action) are
      // retargeted to the input but happen outside the field — don't reopen.
      if (!insideField) return
    }
    openPicker()
  }

  return (
    <FieldLabel label={label} className={className ?? 'min-w-0'}>
      <span className="min-w-0" onClick={handleFieldClick} onKeyDown={(e) => e.key === 'Enter' && openPicker()} role="presentation">
        <TextInput
          ref={inputRef}
          type="date"
          value={value}
          min={min}
          max={max}
          aria-invalid={invalid || undefined}
          onChange={(e) => onChange(e.target.value)}
          leftIcon={<IconCalendar className="h-4 w-4 text-white" />}
          className={`w-full min-w-0 max-w-full appearance-none cursor-pointer normal-case tracking-normal [&::-webkit-calendar-picker-indicator]:hidden ${
            invalid ? 'border-critical' : ''
          }`}
        />
      </span>
    </FieldLabel>
  )
}
