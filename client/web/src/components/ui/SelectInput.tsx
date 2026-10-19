import type { ReactNode, SelectHTMLAttributes } from 'react'
import { cx } from './classes'

export interface SelectInputProps extends SelectHTMLAttributes<HTMLSelectElement> {
  icon?: ReactNode
  /** Draws the review outline used when an extracted value is below the confidence threshold. */
  invalid?: boolean
}

/** Mobile stays at 16px to avoid iOS Safari's focus zoom; `sm:` carries the compact size. */
export function SelectInput({ className, icon, invalid, children, ...props }: SelectInputProps) {
  return (
    <span className="relative">
      <select
        aria-invalid={invalid || undefined}
        className={cx(
          'w-full appearance-none rounded-xl border border-line bg-surface-sunken px-3.5 py-2.5 pr-10 text-base font-normal text-ink outline-none sm:text-sm',
          invalid && 'border-critical ring-2 ring-critical',
          className
        )}
        {...props}
      >
        {children}
      </select>
      {icon && <span className="pointer-events-none absolute right-3 top-1/2 -translate-y-1/2 text-muted">{icon}</span>}
    </span>
  )
}
