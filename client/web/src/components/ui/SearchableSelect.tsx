import { useEffect, useMemo, useRef, useState, type ReactNode } from 'react'
import { IconChevronDown } from '@/components/icons'
import { Button } from './Button'
import { TextInput } from './TextInput'

export interface SearchableSelectOption {
  value: string
  label: string
  icon?: ReactNode
  trailing?: ReactNode
}

export interface SearchableSelectProps {
  value: string
  onChange: (value: string) => void
  options: SearchableSelectOption[]
  placeholder?: string
  searchPlaceholder?: string
  emptyOptionLabel?: string
  emptyMessage?: string
  /** Render the list inline (expands in flow) instead of as an absolute popover. */
  inline?: boolean
  className?: string
}

export function SearchableSelect({
  value,
  onChange,
  options,
  placeholder = 'Select…',
  searchPlaceholder = 'Search…',
  emptyOptionLabel,
  emptyMessage = 'No results',
  inline = false,
  className,
}: SearchableSelectProps) {
  const [open, setOpen] = useState(false)
  const [query, setQuery] = useState('')
  const rootRef = useRef<HTMLDivElement>(null)
  const selected = options.find((option) => option.value === value)
  const filtered = useMemo(() => {
    const q = query.trim().toLowerCase()
    if (!q) return options
    return options.filter((option) => option.label.toLowerCase().includes(q))
  }, [options, query])

  useEffect(() => {
    if (!open || inline) return
    function handlePointerDown(event: PointerEvent) {
      if (!rootRef.current?.contains(event.target as Node)) setOpen(false)
    }
    document.addEventListener('pointerdown', handlePointerDown)
    return () => document.removeEventListener('pointerdown', handlePointerDown)
  }, [open, inline])

  function pick(next: string) {
    onChange(next)
    setOpen(false)
    setQuery('')
  }

  const panelClass = inline
    ? 'mt-2 w-full rounded-2xl border border-line bg-surface p-2 shadow-[var(--shadow-sm)]'
    : 'absolute left-0 right-0 top-[calc(100%+8px)] z-20 w-full rounded-2xl border border-line bg-surface p-2 shadow-[var(--shadow-md)]'

  return (
    <div ref={rootRef} className={`relative block w-full ${className ?? ''}`}>
      <Button
        variant="custom"
        type="button"
        aria-expanded={open}
        onClick={() => setOpen((v) => !v)}
        className="flex w-full items-center justify-between gap-2 rounded-xl border border-line bg-surface-sunken px-3.5 py-2.5 text-left text-sm font-normal normal-case tracking-normal text-ink outline-none"
      >
        <span className="flex min-w-0 items-center gap-2">
          {selected?.icon}
          <span className={`truncate ${selected ? '' : 'text-muted'}`}>{selected ? selected.label : placeholder}</span>
        </span>
        <IconChevronDown className={`h-4 w-4 flex-none text-muted transition-transform ${open ? 'rotate-180' : ''}`} />
      </Button>

      {open && (
        <div className={panelClass}>
          <TextInput
            type="search"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder={searchPlaceholder}
            className="mb-2 w-full"
          />
          <div className="max-h-56 w-full overflow-auto">
            {emptyOptionLabel && (
              <Button
                variant="custom"
                onClick={() => pick('')}
                className="w-full rounded-xl px-3 py-2 text-left text-sm font-semibold text-ink hover:bg-surface-sunken"
              >
                {emptyOptionLabel}
              </Button>
            )}
            {filtered.map((option) => (
              <Button
                key={option.value}
                variant="custom"
                onClick={() => pick(option.value)}
                className="flex w-full items-center gap-2 rounded-xl px-3 py-2 text-left text-sm font-semibold text-ink hover:bg-surface-sunken"
              >
                {option.icon}
                <span className="min-w-0 flex-1 truncate">{option.label}</span>
                {option.trailing}
              </Button>
            ))}
            {filtered.length === 0 && <p className="m-0 px-3 py-2 text-xs font-normal text-muted">{emptyMessage}</p>}
          </div>
        </div>
      )}
    </div>
  )
}
