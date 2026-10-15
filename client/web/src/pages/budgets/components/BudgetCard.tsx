import { ProgressBar } from '@/components/ProgressBar'
import { formatAmount } from '@/lib/currency'
import { getPlaceholderCategory, type PlaceholderCategoryId } from '@/lib/placeholderCategories'

interface BudgetCardProps {
  categoryId: PlaceholderCategoryId
  spent: number
  limit: number
}

export function BudgetCard({ categoryId, spent, limit }: BudgetCardProps) {
  const category = getPlaceholderCategory(categoryId)
  const overBudget = spent > limit
  const percent = (spent / limit) * 100

  return (
    <div className="flex flex-col gap-2.5 rounded-2xl border border-line bg-surface p-3.5">
      <div className="flex items-center gap-2 text-[13.5px]">
        <span className="h-2 w-2 rounded-full" style={{ background: category.colorVar }} />
        <strong className="flex-1">{category.label}</strong>
        <span className={`font-mono text-xs tabular-nums ${overBudget ? 'font-semibold text-critical' : 'text-ink-soft'}`}>
          {formatAmount(spent)} / {formatAmount(limit)}
        </span>
      </div>
      <ProgressBar percent={percent} colorVar={category.colorVar} critical={overBudget} />
      {overBudget && (
        <span className="self-start rounded-full bg-critical-soft px-2.5 py-[3px] text-[11px] font-bold text-critical">
          Over budget
        </span>
      )}
    </div>
  )
}
