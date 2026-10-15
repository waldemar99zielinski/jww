import { filterLabels } from '../filterLabels'
import type { TransactionsViewProps } from '../types'

export function TransactionFiltersSummary({
  filter,
  userFilter,
  users,
  categoryFilter,
  categories,
  nameFilter,
  dateFrom,
  dateTo,
}: Pick<TransactionsViewProps, 'filter' | 'userFilter' | 'users' | 'categoryFilter' | 'categories' | 'nameFilter' | 'dateFrom' | 'dateTo'>) {
  const selectedUser = users.find((option) => option.id === userFilter)
  const selectedCategory = categories.find((category) => category.id === categoryFilter)
  const activeFilters = [filter !== 'all', userFilter !== '', categoryFilter !== '', nameFilter.trim() !== '', dateFrom !== '', dateTo !== ''].filter(Boolean).length

  if (activeFilters === 0) return null

  return (
    <div className="flex flex-wrap items-center gap-1.5 rounded-2xl border border-line bg-surface-sunken px-3 py-2 text-xs text-muted md:px-4 md:py-3">
      <span className="font-bold text-ink-soft">Showing:</span>
      <span>{filterLabels[filter]}</span>
      {selectedUser && <span>· {selectedUser.display_name}</span>}
      {selectedCategory && <span>· {selectedCategory.name}</span>}
      {nameFilter.trim() && <span>· {nameFilter.trim()}</span>}
      {dateFrom && <span>· from {dateFrom}</span>}
      {dateTo && <span>· to {dateTo}</span>}
    </div>
  )
}
