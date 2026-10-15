import { ProgressBar } from '@/components/ProgressBar'
import { getCategoryVisual } from '@/lib/categoryVisuals'
import { formatAmount } from '@/lib/currency'

interface CategoryBreakdownProps {
  items: readonly { categoryName: string; amount: number; percent: number }[]
}

export function CategoryBreakdown({ items }: CategoryBreakdownProps) {
  if (items.length === 0) {
    return <p className="m-0 text-xs text-muted">No expenses yet this month.</p>
  }

  return (
    <ul className="flex flex-col gap-3">
      {items.map((item) => {
        const { colorVar } = getCategoryVisual(item.categoryName)
        return (
          <li
            key={item.categoryName}
            className="grid grid-cols-[auto_74px_1fr_auto] items-center gap-2.5 text-[13px]"
          >
            <span className="h-2 w-2 rounded-full" style={{ background: colorVar }} />
            <span className="truncate text-ink-soft">{item.categoryName}</span>
            <ProgressBar percent={item.percent} colorVar={colorVar} />
            <span className="font-mono text-xs tabular-nums text-ink-soft">{formatAmount(item.amount)}</span>
          </li>
        )
      })}
    </ul>
  )
}
