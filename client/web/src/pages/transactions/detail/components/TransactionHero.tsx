import { IconTag } from '@/components/icons'
import { getCategoryVisual } from '@/lib/categoryVisuals'
import { cssVars } from '@/lib/cssVars'

export interface TransactionHeroProps {
  title: string
  categoryName: string
  amount: string
  isIncome: boolean
}

/** The icon + name + signed amount block at the top of the transaction detail screen. */
export function TransactionHero({ title, categoryName, amount, isIncome }: TransactionHeroProps) {
  const { Icon, colorVar } = categoryName
    ? getCategoryVisual(categoryName)
    : { Icon: IconTag, colorVar: 'var(--accent)' }

  return (
    <div className="flex flex-col items-center gap-2 py-2 text-center">
      <span
        className="flex h-14 w-14 items-center justify-center rounded-2xl category-tint"
        style={cssVars({ '--tint': colorVar })}
      >
        <Icon className="h-7 w-7" />
      </span>
      <strong className="text-[15px] font-semibold">{title}</strong>
      <span className={`font-mono text-2xl tabular-nums ${isIncome ? 'font-bold text-ink' : 'text-ink-soft'}`}>
        {amount}
      </span>
    </div>
  )
}
