import type { TransactionDetailItem } from '../types'

export interface DetailListProps {
  items: TransactionDetailItem[]
}

/** Renders a transaction's fields as a label/value definition list. */
export function DetailList({ items }: DetailListProps) {
  return (
    <dl className="m-0 flex flex-col divide-y divide-line">
      {items.map((item) => (
        <div key={item.label} className="flex items-baseline justify-between gap-4 py-2.5">
          <dt className="flex-none text-xs text-muted">{item.label}</dt>
          <dd className="m-0 min-w-0 break-words text-right text-[13.5px] font-medium">{item.value}</dd>
        </div>
      ))}
    </dl>
  )
}
