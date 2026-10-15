type LogoMarkSize = 'sm' | 'md' | 'lg'

const markSizeClasses: Record<LogoMarkSize, string> = {
  sm: 'h-7 w-7 text-xs',
  md: 'h-9 w-9 text-sm',
  lg: 'h-14 w-14 text-xl',
}

export interface LogoMarkProps {
  size?: LogoMarkSize
  className?: string
}

/** Compact "JWW" monogram badge for tight spaces (nav headers, etc.) — the wordmark's initials. */
export function LogoMark({ size = 'md', className = '' }: LogoMarkProps) {
  return (
    <span
      role="img"
      aria-label="Joint Wallet Watch"
      className={`inline-flex flex-none items-center justify-center rounded-xl bg-accent font-bold tracking-tight text-on-accent ${markSizeClasses[size]} ${className}`}
    >
      JWW
    </span>
  )
}
