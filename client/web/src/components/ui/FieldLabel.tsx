import type { ReactNode } from 'react'
import { cx } from './classes'

export interface FieldLabelProps {
  label: ReactNode
  children: ReactNode
  className?: string
}

export function FieldLabel({ label, children, className }: FieldLabelProps) {
  return (
    <label className={cx('flex flex-col gap-1.5 text-[11.5px] font-bold uppercase tracking-wide text-muted', className)}>
      {label}
      {children}
    </label>
  )
}
