import { TransactionRow } from '@/components/TransactionRow'
import { formatTime } from '@/lib/dates'
import { TransactionFiltersDropdown } from './components/TransactionFiltersDropdown'
import { TransactionFiltersSummary } from './components/TransactionFiltersSummary'
import type { TransactionsViewProps } from './types'

export default function TransactionsMobile(props: TransactionsViewProps) {
  const { grouped, isEmpty } = props

  return (
    <div className="flex flex-col gap-4 px-4 pb-4">
      <div className="sticky top-0 z-20 -mx-4 flex items-center justify-between bg-paper px-4 pb-3 pt-6">
        <h2 className="m-0 text-[19px] font-bold">Transactions</h2>
        <TransactionFiltersDropdown {...props} align="left" compact />
      </div>

      <TransactionFiltersSummary {...props} />

      {isEmpty && <p className="m-0 text-xs text-muted">No transactions yet — add one from the + button below.</p>}

      {grouped.map(([day, dayTransactions]) => (
        <div key={day} className="flex flex-col gap-1">
          <p className="m-0 text-[11.5px] font-bold uppercase tracking-wide text-muted">{day}</p>
          <ul className="flex flex-col">
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
  )
}
