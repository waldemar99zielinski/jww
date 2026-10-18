import { IconCalendar, IconPaperclip } from '@/components/icons'
import { Button } from '@/components/ui/Button'

export interface QuickFieldsRowProps {
  dateLabel: string
  dateOpen: boolean
  onToggleDate: () => void
  hasAttachment: boolean
  onOpenAttachment: () => void
}

function chipClass(active: boolean) {
  return `flex min-w-0 flex-1 items-center justify-center gap-1.5 rounded-xl border px-2 py-[11px] text-[13px] font-semibold ${
    active ? 'border-accent bg-accent-soft text-accent-strong' : 'border-line bg-surface-sunken text-ink-soft'
  }`
}

export function QuickFieldsRow({ dateLabel, dateOpen, onToggleDate, hasAttachment, onOpenAttachment }: QuickFieldsRowProps) {
  return (
    <div className="flex gap-2.5">
      <Button variant="custom" onClick={onToggleDate} className={chipClass(dateOpen)}>
        <IconCalendar className="h-4 w-4 flex-none" />
        <span className="truncate">{dateLabel}</span>
      </Button>
      <Button variant="custom" onClick={onOpenAttachment} className={chipClass(hasAttachment)}>
        <IconPaperclip className="h-4 w-4 flex-none" />
        <span className="truncate">{hasAttachment ? 'Attached' : 'Attach'}</span>
      </Button>
    </div>
  )
}
