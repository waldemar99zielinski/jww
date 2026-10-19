import { IconCalendar } from '@/components/icons'
import { Button } from '@/components/ui/Button'

export interface DateQuickFieldProps {
  dateLabel: string
  dateOpen: boolean
  onToggleDate: () => void
  /** Review outline for a low-confidence extracted date. */
  invalid?: boolean
}

function chipClass(active: boolean, invalid: boolean) {
  return `flex min-w-0 flex-1 items-center justify-center gap-1.5 rounded-xl border px-2 py-[11px] text-[13px] font-semibold ${
    active ? 'border-accent bg-accent-soft text-accent-strong' : 'border-line bg-surface-sunken text-ink-soft'
  } ${invalid ? 'border-critical ring-2 ring-critical' : ''}`
}

/** One date shortcut that reveals the full date/time fields. File upload lives behind the header attachment button. */
export function DateQuickField({ dateLabel, dateOpen, onToggleDate, invalid = false }: DateQuickFieldProps) {
  return (
    <Button
      variant="custom"
      onClick={onToggleDate}
      aria-invalid={invalid || undefined}
      className={chipClass(dateOpen, invalid)}
    >
      <IconCalendar className="h-4 w-4 flex-none" />
      <span className="truncate">{dateLabel}</span>
    </Button>
  )
}
