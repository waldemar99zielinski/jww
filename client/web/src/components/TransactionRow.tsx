import type { TransactionDto } from '@/api/types'
import { IconTag } from '@/components/icons'
import { getCategoryVisual } from '@/lib/categoryVisuals'
import { cssVars } from '@/lib/cssVars'
import { formatSignedAmount } from '@/lib/currency'

interface TransactionRowProps {
  transaction: TransactionDto
  subtitle: string
}

export function TransactionRow({ transaction, subtitle }: TransactionRowProps) {
  const { Icon, colorVar } = transaction.category_name
    ? getCategoryVisual(transaction.category_name)
    : { Icon: IconTag, colorVar: 'var(--accent)' }
  const title = transaction.name || transaction.description || transaction.category_name

  return (
    <li className="flex items-center gap-3 py-2">
      <span
        className="flex h-9 w-9 flex-none items-center justify-center rounded-xl category-tint"
        style={cssVars({ '--tint': colorVar })}
      >
        <Icon className="h-[18px] w-[18px]" />
      </span>
      <div className="flex min-w-0 flex-1 flex-col">
        <strong className="truncate text-[13.5px] font-semibold">{title}</strong>
        <span className="truncate text-xs text-muted">{subtitle}</span>
      </div>
      <span
        className={`shrink-0 whitespace-nowrap font-mono text-[13.5px] tabular-nums ${transaction.type === 'income' ? 'font-bold text-ink' : 'text-ink-soft'}`}
      >
        {formatSignedAmount(Number(transaction.amount), transaction.type)}
      </span>
    </li>
  )
}
