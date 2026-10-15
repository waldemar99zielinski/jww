import type { ReactNode, SelectHTMLAttributes } from 'react'
import { cx } from './classes'

export interface SelectInputProps extends SelectHTMLAttributes<HTMLSelectElement> {
  icon?: ReactNode
}

export function SelectInput({ className, icon, children, ...props }: SelectInputProps) {
  return (
    <span className="relative">
      <select
        className={cx(
          'w-full appearance-none rounded-xl border border-line bg-surface-sunken px-3.5 py-2.5 pr-10 text-base font-normal text-ink outline-none sm:text-sm',
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
