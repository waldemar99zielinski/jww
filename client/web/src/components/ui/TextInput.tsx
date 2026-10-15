import { forwardRef, type InputHTMLAttributes, type ReactNode } from 'react'
import { cx } from './classes'

export interface TextInputProps extends InputHTMLAttributes<HTMLInputElement> {
  leftIcon?: ReactNode
}

export const TextInput = forwardRef<HTMLInputElement, TextInputProps>(function TextInput(
  { className, leftIcon, ...props },
  ref
) {
  if (leftIcon) {
    return (
      <span className="relative block w-full min-w-0">
        <span className="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-muted">{leftIcon}</span>
        <input
          ref={ref}
          className={cx(
            'block w-full min-w-0 rounded-xl border border-line bg-surface-sunken px-3.5 py-2.5 pl-9 text-base font-normal text-ink outline-none placeholder:text-muted sm:text-sm',
            className
          )}
          {...props}
        />
      </span>
    )
  }

  return (
    <input
      ref={ref}
      className={cx(
        'block min-w-0 rounded-xl border border-line bg-surface-sunken px-3.5 py-2.5 text-base font-normal text-ink outline-none placeholder:text-muted sm:text-sm',
        className
      )}
      {...props}
    />
  )
})
