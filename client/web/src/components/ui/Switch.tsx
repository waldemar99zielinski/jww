import { cx } from './classes'

export interface SwitchProps {
  checked: boolean
  onChange: (checked: boolean) => void
  label?: string
  className?: string
}

export function Switch({ checked, onChange, label, className }: SwitchProps) {
  return (
    <button
      type="button"
      role="switch"
      aria-checked={checked}
      aria-label={label}
      onClick={() => onChange(!checked)}
      className={cx(
        'group relative h-6 w-11 flex-none rounded-full p-0 transition-colors duration-200 ease-out',
        checked ? 'bg-accent' : 'bg-line',
        className
      )}
    >
      <span
        className={cx(
          'absolute left-0.5 top-0.5 h-5 w-5 rounded-full bg-surface shadow-[var(--shadow-sm)] will-change-transform',
          'transition-[translate,scale] duration-200 ease-[cubic-bezier(0.34,1.4,0.64,1)] group-active:scale-90',
          checked ? 'translate-x-5' : 'translate-x-0'
        )}
      />
    </button>
  )
}
