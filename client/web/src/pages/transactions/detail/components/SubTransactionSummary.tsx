import type { TransactionDetailSubTransactionItem } from '../types'

export interface SubTransactionSummaryProps {
  items: TransactionDetailSubTransactionItem[]
}

/**
 * Read-only line items shown on the detail screen when the transaction was split into products
 * or parts. Editing, adding, and removing rows stays in the edit form's `SubTransactionList`.
 */
export function SubTransactionSummary({ items }: SubTransactionSummaryProps) {
  return (
    <section className="flex flex-col gap-2" aria-label="Sub-transaction items">
      <h2 className="m-0 text-xs font-semibold uppercase tracking-wide text-muted">Items</h2>
      <ul className="m-0 flex list-none flex-col gap-2 p-0">
        {items.map((item) => (
          <li
            key={item.id}
            className="flex items-center justify-between gap-3 rounded-xl border border-line bg-surface-sunken px-3.5 py-2.5"
          >
            <span className="flex min-w-0 flex-col">
              <span className="truncate text-[13px] font-semibold text-ink">{item.name}</span>
              {item.label !== '' && <span className="truncate text-xs text-muted">{item.label}</span>}
            </span>
            <span className="shrink-0 whitespace-nowrap font-mono text-[13px] tabular-nums text-ink">
              {item.amount}
            </span>
          </li>
        ))}
      </ul>
    </section>
  )
}
