import { useEffect, useState } from 'react'
import { LogoWord } from '@/components/LogoWord'

type LogoSize = 'sm' | 'md' | 'lg'

/**
 * Fluid sizes. `md`/`lg` are the "Horizontal Unfold" spec's own clamp() tokens (comparison /
 * expanded-preview sizes); `sm` is a project-specific compact variant for tight nav spaces,
 * not part of the spec.
 */
const sizeFontSizes: Record<LogoSize, string> = {
  sm: 'clamp(14px, 2vw, 20px)',
  md: 'clamp(20px, 2.5vw, 34px)',
  lg: 'clamp(23px, 5vw, 64px)',
}

/** How long "JWW" holds before unfolding. */
const HOLD_MS = 300

function prefersReducedMotion(): boolean {
  return typeof window !== 'undefined' && window.matchMedia('(prefers-reduced-motion: reduce)').matches
}

export interface LogoProps {
  /** Plays the JWW -> Joint Wallet Watch unfold once on mount instead of rendering pre-resolved. */
  animated?: boolean
  size?: LogoSize
  className?: string
}

/** The app's wordmark. Reuse this (not just the JWW string) everywhere the brand name appears. */
export function Logo({ animated = false, size = 'md', className = '' }: LogoProps) {
  const [skipAnimation] = useState(() => !animated || prefersReducedMotion())
  const [resolved, setResolved] = useState(skipAnimation)

  useEffect(() => {
    if (skipAnimation) return
    const timer = setTimeout(() => setResolved(true), HOLD_MS)
    return () => clearTimeout(timer)
  }, [skipAnimation])

  return (
    <span
      role="img"
      aria-label="Joint Wallet Watch"
      className={`jww-wordmark items-baseline justify-center ${animated ? 'flex w-full' : 'inline-flex'} ${className}`}
      style={{ fontSize: sizeFontSizes[size] }}
    >
      <LogoWord word="Joint" resolved={resolved} last={false} />
      <LogoWord word="Wallet" resolved={resolved} last={false} />
      <LogoWord word="Watch" resolved={resolved} last />
    </span>
  )
}
