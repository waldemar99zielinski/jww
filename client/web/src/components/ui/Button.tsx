import type { ButtonHTMLAttributes } from 'react'
import { cx } from './classes'

type ButtonVariant = 'primary' | 'secondary' | 'ghost' | 'link' | 'danger' | 'custom'

export interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: ButtonVariant
  fullWidth?: boolean
}

const variants: Record<ButtonVariant, string> = {
  primary: 'rounded-xl bg-accent px-4 py-2.5 text-[13px] font-bold text-on-accent shadow-[var(--shadow-sm)] disabled:opacity-40',
  secondary:
    'rounded-xl border border-line bg-surface-sunken px-4 py-2.5 text-[13px] font-semibold text-ink-soft disabled:opacity-40',
  ghost: 'rounded-xl px-3 py-2 text-[13px] font-semibold text-ink-soft disabled:opacity-40',
  link: 'text-xs font-bold text-accent-strong disabled:opacity-40',
  danger: 'rounded-xl border border-line bg-surface-sunken px-4 py-2.5 text-[13px] font-semibold text-critical disabled:opacity-40',
  custom: '',
}

export function Button({ variant = 'secondary', fullWidth, className, type = 'button', ...props }: ButtonProps) {
  return <button type={type} className={cx(variants[variant], fullWidth && 'w-full', className)} {...props} />
}
