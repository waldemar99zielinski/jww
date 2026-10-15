import { cssVars } from '@/lib/cssVars'

interface ProgressBarProps {
  percent: number
  colorVar?: string
  critical?: boolean
}

export function ProgressBar({ percent, colorVar, critical }: ProgressBarProps) {
  return (
    <div className="h-1.5 overflow-hidden rounded-full bg-surface-sunken">
      <div
        className={`h-full rounded-full bar-fill ${critical ? 'critical' : ''}`}
        style={{ width: `${Math.min(100, Math.max(0, percent))}%`, ...cssVars({ '--tint': colorVar ?? 'var(--ink)' }) }}
      />
    </div>
  )
}
