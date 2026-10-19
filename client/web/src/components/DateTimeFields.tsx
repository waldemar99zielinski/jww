import { Button } from '@/components/ui/Button'
import { FieldLabel } from '@/components/ui/FieldLabel'
import { TextInput } from '@/components/ui/TextInput'

export interface DateTimeFieldsProps {
  date: string
  time: string
  onDateChange: (value: string) => void
  onTimeChange: (value: string) => void
  onUseToday: () => void
  /** Review outlines for low-confidence extracted values. */
  invalidDate?: boolean
  invalidTime?: boolean
}

export function DateTimeFields({
  date,
  time,
  onDateChange,
  onTimeChange,
  onUseToday,
  invalidDate = false,
  invalidTime = false,
}: DateTimeFieldsProps) {
  return (
    <div className="flex flex-col gap-2.5">
      <div className="flex gap-2.5">
        <FieldLabel label="Date" className="flex-1">
          <TextInput type="date" value={date} onChange={(e) => onDateChange(e.target.value)} invalid={invalidDate} />
        </FieldLabel>
        <FieldLabel label="Time" className="w-28">
          <TextInput type="time" value={time} onChange={(e) => onTimeChange(e.target.value)} invalid={invalidTime} />
        </FieldLabel>
      </div>
      <Button variant="link" onClick={onUseToday} className="self-start">
        Use today
      </Button>
    </div>
  )
}
