import { useId } from 'react'
import { Button } from '@/components/ui/Button'

export interface OfflineScreenProps {
  onRetry: () => void
  /** True after a Retry that found the connection still down; drives the failure line. */
  stillOffline: boolean
}

/** The wallet-and-coin mark (the Loader's static geometry), decorative. */
function OfflineMark() {
  return (
    <svg
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth={2}
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
      className="h-[46px] w-[46px]"
    >
      <circle cx="13.6" cy="4.6" r="2.1" fill="currentColor" stroke="none" />
      <rect x="3.4" y="8.1" width="17.2" height="11" rx="2.9" />
      <path d="M15.4 12.3h3.1a1.9 1.9 0 0 1 0 3.8h-3.1a1.9 1.9 0 0 1 0-3.8Z" />
    </svg>
  )
}

/**
 * The blocking "You're offline" panel. Pure display, props only: focus-on-mount and focus
 * restoration live in OfflineGate, and the polite live region (which also carries the failure
 * message below) lives there too. The visible failure line is a duplicate of that announcement.
 */
export function OfflineScreen({ onRetry, stillOffline }: OfflineScreenProps) {
  const titleId = useId()
  const descriptionId = useId()

  return (
    <div
      role="alertdialog"
      aria-modal="true"
      aria-labelledby={titleId}
      aria-describedby={descriptionId}
      className="screen-enter fixed inset-0 z-50 flex min-h-screen flex-col items-center justify-center gap-4 bg-paper px-6 text-center"
    >
      <span
        className="flex h-[74px] w-[74px] items-center justify-center rounded-[20px] bg-accent text-on-accent"
        aria-hidden="true"
      >
        <OfflineMark />
      </span>
      <h2 id={titleId} className="m-0 text-xl font-bold tracking-tight">
        You&rsquo;re offline
      </h2>
      <p id={descriptionId} className="m-0 max-w-[34ch] text-xs text-ink-soft">
        Joint Wallet Watch needs a connection. Your data isn&rsquo;t stored on this device, so there&rsquo;s nothing to show
        until you&rsquo;re back.
      </p>
      <Button
        variant="primary"
        onClick={onRetry}
        className="mt-1 h-12 w-full max-w-[240px] lg:w-auto lg:min-w-[180px]"
      >
        Retry
      </Button>
      <p className="m-0 min-h-4 text-[11px] text-critical">
        {stillOffline ? 'Still offline — check your connection.' : ''}
      </p>
    </div>
  )
}
