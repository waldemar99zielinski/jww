import { ProgressBar } from '@/components/ProgressBar'
import { formatAmount } from '@/lib/currency'

interface GoalCardProps {
  label: string
  saved: number
  target: number
}

export function GoalCard({ label, saved, target }: GoalCardProps) {
  return (
    <div className="flex flex-col gap-2.5 rounded-2xl border border-line bg-surface p-3.5">
      <strong className="text-[13.5px]">{label}</strong>
      <ProgressBar percent={(saved / target) * 100} colorVar="var(--accent)" />
      <span className="font-mono text-xs tabular-nums text-ink-soft">
        {formatAmount(saved)} / {formatAmount(target)}
      </span>
    </div>
  )
}
