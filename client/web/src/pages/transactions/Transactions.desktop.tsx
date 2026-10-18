import { TransactionRow } from '@/components/TransactionRow'
import { formatTime } from '@/lib/dates'
import { TransactionFiltersDropdown } from './components/TransactionFiltersDropdown'
import { TransactionFiltersSummary } from './components/TransactionFiltersSummary'
import type { TransactionsViewProps } from './types'

export default function TransactionsDesktop(props: TransactionsViewProps) {
  const { grouped, isEmpty } = props

  return (
    <div className="flex flex-col gap-6">
      <div className="sticky top-0 z-20 flex items-center justify-between gap-4 bg-paper py-4">
        <h1 className="m-0 text-2xl font-bold">Transactions</h1>
        <TransactionFiltersDropdown {...props} />
      </div>

      <TransactionFiltersSummary {...props} />

      {isEmpty && <p className="m-0 text-xs text-muted">No transactions yet — add one from the sidebar.</p>}

      <div className="rounded-2xl border border-line bg-surface">
        {grouped.map(([day, dayTransactions], index) => (
          <div key={day} className={index > 0 ? 'border-t border-line' : ''}>
            <p className="m-0 px-5 pt-4 text-[11.5px] font-bold uppercase tracking-wide text-muted">{day}</p>
            <ul className="flex flex-col divide-y divide-line px-5">
              {dayTransactions.map((transaction) => (
                <TransactionRow
                  key={transaction.id}
                  transaction={transaction}
                  subtitle={`${transaction.category_name} · ${formatTime(transaction.created_at)}`}
                  to={`/transactions/${transaction.id}`}
                />
              ))}
            </ul>
          </div>
        ))}
      </div>
    </div>
  )
}
