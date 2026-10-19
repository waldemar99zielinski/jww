import { forwardRef, type InputHTMLAttributes, type ReactNode } from 'react'
import { cx } from './classes'

export interface TextInputProps extends InputHTMLAttributes<HTMLInputElement> {
  leftIcon?: ReactNode
  /** Draws the review outline used when an extracted value is below the confidence threshold. */
  invalid?: boolean
}

/**
 * Keep mobile controls at `text-base` (16px): iOS Safari zooms the viewport when a focused input
 * is smaller than that. `sm:` is where compact desktop sizing belongs; never override the mobile
 * size with a smaller arbitrary value.
 */
export const TextInput = forwardRef<HTMLInputElement, TextInputProps>(function TextInput(
  { className, leftIcon, invalid, ...props },
  ref
) {
  const invalidClass = invalid ? 'border-critical ring-2 ring-critical' : ''

  if (leftIcon) {
    return (
      <span className="relative block w-full min-w-0">
        <span className="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-muted">{leftIcon}</span>
        <input
          ref={ref}
          aria-invalid={invalid || undefined}
          className={cx(
            'block w-full min-w-0 rounded-xl border border-line bg-surface-sunken px-3.5 py-2.5 pl-9 text-base font-normal text-ink outline-none placeholder:text-muted sm:text-sm',
            invalidClass,
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
      aria-invalid={invalid || undefined}
      className={cx(
        'block min-w-0 rounded-xl border border-line bg-surface-sunken px-3.5 py-2.5 text-base font-normal text-ink outline-none placeholder:text-muted sm:text-sm',
        invalidClass,
        className
      )}
      {...props}
    />
  )
})
