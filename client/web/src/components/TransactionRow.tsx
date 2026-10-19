import { Link } from 'react-router-dom'
import type { TransactionDto } from '@/api/types'
import { IconTag } from '@/components/icons'
import { getCategoryVisual } from '@/lib/categoryVisuals'
import { cssVars } from '@/lib/cssVars'
import { formatSignedAmount } from '@/lib/currency'

interface TransactionRowProps {
  transaction: TransactionDto
  subtitle: string
  /** When set, the whole row becomes a link to that route (e.g. the transaction detail screen). */
  to?: string
}

export function TransactionRow({ transaction, subtitle, to }: TransactionRowProps) {
  const { Icon, colorVar } = transaction.category_name
    ? getCategoryVisual(transaction.category_name)
    : { Icon: IconTag, colorVar: 'var(--accent)' }
  const title = transaction.name || transaction.description || transaction.category_name

  const content = (
    <>
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
        {formatSignedAmount(transaction.amount / 100, transaction.type, transaction.currency_code)}
      </span>
    </>
  )

  if (to) {
    return (
      <li>
        <Link
          to={to}
          className="-mx-2 flex items-center gap-3 rounded-lg px-2 py-2 transition-colors hover:bg-surface-sunken"
        >
          {content}
        </Link>
      </li>
    )
  }

  return <li className="flex items-center gap-3 py-2">{content}</li>
}
