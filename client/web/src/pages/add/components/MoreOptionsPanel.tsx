import type { UserDto } from '@/api/types'
import { IconChevronDown } from '@/components/icons'
import { Button } from '@/components/ui/Button'
import { FieldLabel } from '@/components/ui/FieldLabel'
import { SelectInput } from '@/components/ui/SelectInput'

export interface MoreOptionsPanelProps {
  open: boolean
  onToggle: () => void
  users: UserDto[]
  userId: string
  onUserChange: (id: string) => void
  showUserPicker: boolean
}

export function MoreOptionsPanel({ open, onToggle, users, userId, onUserChange, showUserPicker }: MoreOptionsPanelProps) {
  if (!showUserPicker) return null

  return (
    <div className="flex flex-col gap-3">
      <Button variant="custom" onClick={onToggle} className="flex items-center gap-2 self-start text-[13px] font-semibold text-ink-soft">
        <IconChevronDown className={`h-4 w-4 transition-transform ${open ? 'rotate-180' : ''}`} />
        More options
      </Button>

      {open && (
        <div className="flex flex-col gap-4 rounded-2xl border border-line bg-surface p-4">
          <FieldLabel label="Paid by">
            <SelectInput value={userId} onChange={(e) => onUserChange(e.target.value)} icon={<IconChevronDown className="h-4 w-4" />}>
              {users.map((member) => (
                <option key={member.id} value={member.id}>
                  {member.display_name}
                </option>
              ))}
            </SelectInput>
          </FieldLabel>
        </div>
      )}
    </div>
  )
}
