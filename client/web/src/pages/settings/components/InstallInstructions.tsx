import type { ReactNode } from 'react'
import { IconClose } from '@/components/icons'
import { BottomSheet } from '@/components/ui/BottomSheet'
import { IconButton } from '@/components/ui/IconButton'

export type InstallInstructionsPlatform = 'ios-safari' | 'macos-safari'

export interface InstallInstructionsProps {
  platform: InstallInstructionsPlatform
  /** Mobile renders a dismissible sheet; desktop renders an inline card under the row. */
  compact: boolean
  open: boolean
  onClose: () => void
}

const CONTENT: Record<InstallInstructionsPlatform, { description: string; note: string; steps: ReactNode[] }> = {
  'ios-safari': {
    description: 'Add it to your Home Screen from Safari’s share menu.',
    note: 'Safari does not let the app open this dialog for you. The steps above are the only path on iOS.',
    steps: [
      <>
        Tap the <strong className="font-semibold text-ink">⋯ menu</strong> (or the Share button) in Safari’s toolbar.
      </>,
      <>
        Choose <strong className="font-semibold text-ink">Share</strong>, then{' '}
        <strong className="font-semibold text-ink">Add to Home Screen</strong>.
      </>,
      <>
        Leave <strong className="font-semibold text-ink">“Open as Web App”</strong> switched on, then tap{' '}
        <strong className="font-semibold text-ink">Add</strong>.
      </>,
      <>Open Joint Wallet Watch from your Home Screen — it launches without Safari’s interface.</>,
    ],
  },
  'macos-safari': {
    description: 'Add it to your Dock from Safari’s File menu.',
    note: 'Safari does not let the app open this dialog for you. The steps above are the only path on macOS.',
    steps: [
      <>
        In Safari, choose <strong className="font-semibold text-ink">File → Add to Dock…</strong>.
      </>,
      <>
        Confirm the name, then click <strong className="font-semibold text-ink">Add</strong>.
      </>,
      <>Launch it from the Dock like any other app.</>,
    ],
  },
}

/**
 * Per-platform install steps. One component, two presentations: a `BottomSheet` on mobile and an
 * inline card on desktop — the compact/inline pattern `TransactionFiltersDropdown` already uses.
 * Copy is identical across both; only the surface changes.
 */
export function InstallInstructions({ platform, compact, open, onClose }: InstallInstructionsProps) {
  const { description, note, steps } = CONTENT[platform]

  const stepsList = (
    <ol className="m-0 flex list-none flex-col gap-3 p-0">
      {steps.map((step, index) => (
        <li key={index} className="flex gap-3 text-[12.5px] leading-relaxed text-ink-soft">
          <span className="flex h-[21px] w-[21px] flex-none items-center justify-center rounded-full bg-accent-soft text-[11px] font-bold text-accent-strong">
            {index + 1}
          </span>
          <span>{step}</span>
        </li>
      ))}
    </ol>
  )

  if (compact) {
    return (
      <BottomSheet open={open} onClose={onClose} title="Install Joint Wallet Watch" description={description}>
        {stepsList}
        <p className="mt-4 text-[11px] text-muted">{note}</p>
      </BottomSheet>
    )
  }

  if (!open) return null

  return (
    <div className="rounded-2xl border border-line bg-surface p-4">
      <div className="mb-3 flex items-start justify-between gap-3">
        <div>
          <p className="m-0 text-[13.5px] font-bold text-ink">Install Joint Wallet Watch</p>
          <p className="m-0 mt-0.5 text-xs text-muted">{description}</p>
        </div>
        <IconButton aria-label="Close" onClick={onClose} icon={<IconClose className="h-4 w-4" />} className="h-8 w-8" />
      </div>
      {stepsList}
      <p className="mt-4 text-[11px] text-muted">{note}</p>
    </div>
  )
}
