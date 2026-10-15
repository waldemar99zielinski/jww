import type { ReactNode, SelectHTMLAttributes } from 'react'
import { cx } from './classes'

export interface InlineSelectProps extends SelectHTMLAttributes<HTMLSelectElement> {
  icon?: ReactNode
}

export function InlineSelect({ className, icon, children, ...props }: InlineSelectProps) {
  return (
    <span className="relative inline-flex items-center">
      <select
        className={cx(
          'cursor-pointer appearance-none rounded-lg bg-transparent py-1 pl-2 pr-7 font-semibold text-muted outline-none',
          className
        )}
        {...props}
      >
        {children}
      </select>
      {icon && <span className="pointer-events-none absolute right-1.5 text-muted">{icon}</span>}
    </span>
  )
}
