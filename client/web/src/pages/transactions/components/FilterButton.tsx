import { IconChevronDown, IconSearch } from '@/components/icons'
import { Button } from '@/components/ui/Button'

export interface FilterButtonProps {
  compact: boolean
  activeCount: number
  open: boolean
  onClick: () => void
}

export function FilterButton({ compact, activeCount, open, onClick }: FilterButtonProps) {
  return (
    <Button
      variant="custom"
      aria-expanded={open}
      onClick={onClick}
      className={`flex items-center gap-2 rounded-full border border-line bg-surface font-bold text-ink-soft shadow-[var(--shadow-sm)] ${
        compact ? 'px-3 py-1.5 text-xs' : 'px-4 py-2 text-xs'
      }`}
    >
      <IconSearch className={compact ? 'h-3.5 w-3.5' : 'h-4 w-4'} />
      Filters
      {activeCount > 0 && (
        <span className="rounded-full bg-accent px-2 py-0.5 text-[10px] font-bold text-on-accent">{activeCount}</span>
      )}
      <IconChevronDown className={`${compact ? 'h-3.5 w-3.5' : 'h-4 w-4'} transition-transform ${open ? 'rotate-180' : ''}`} />
    </Button>
  )
}
