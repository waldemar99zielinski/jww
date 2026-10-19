import { lazy, useMemo, useState } from 'react'
import type { TransactionSortOrder } from '@/api/transactions'
import type { TransactionDto } from '@/api/types'
import { ResponsiveView } from '@/components/ResponsiveView'
import { useAuth } from '@/hooks/useAuth'
import { useCategories } from '@/hooks/useCategories'
import { useTransactions } from '@/hooks/useTransactions'
import { useUsers } from '@/hooks/useUsers'
import { formatDayBucket } from '@/lib/dates'
import type { TransactionFilterValue, TransactionsViewProps } from './types'

const Mobile = lazy(() => import('./Transactions.mobile'))
const Desktop = lazy(() => import('./Transactions.desktop'))

export default function TransactionsPage() {
  const [filter, setFilter] = useState<TransactionFilterValue>('all')
  const [userFilter, setUserFilter] = useState('')
  const [categoryFilter, setCategoryFilter] = useState('')
  const [nameFilter, setNameFilter] = useState('')
  const [dateFrom, setDateFrom] = useState('')
  const [dateTo, setDateTo] = useState('')
  const [sortOrder, setSortOrder] = useState<TransactionSortOrder>('desc')
  const { user } = useAuth()
  const usersQuery = useUsers()
  const incomeCategoriesQuery = useCategories('income')
  const expenseCategoriesQuery = useCategories('expense')
  const categories = useMemo(
    () => [...(incomeCategoriesQuery.data ?? []), ...(expenseCategoriesQuery.data ?? [])],
    [expenseCategoriesQuery.data, incomeCategoriesQuery.data]
  )
  const hasValidDateRange = !dateFrom || !dateTo || dateFrom <= dateTo
  const transactionsQuery = useTransactions({
    user_id: userFilter || undefined,
    category_id: categoryFilter || undefined,
    type: filter === 'all' ? undefined : filter,
    created_from: hasValidDateRange && dateFrom ? new Date(`${dateFrom}T00:00:00`).toISOString() : undefined,
    created_to: hasValidDateRange && dateTo ? new Date(`${dateTo}T23:59:59.999`).toISOString() : undefined,
    sort_key: 'date',
    sort_order: sortOrder,
  })
  const transactions = useMemo(() => {
    const query = nameFilter.trim().toLowerCase()
    const data = transactionsQuery.data ?? []
    if (!query) return data
    return data.filter((transaction) =>
      `${transaction.name} ${transaction.description} ${transaction.category_name}`.toLowerCase().includes(query)
    )
  }, [nameFilter, transactionsQuery.data])

  const grouped = useMemo(() => {
    const byDay = new Map<string, TransactionDto[]>()
    // Keep the server's order (newest-first): bucket rows in the order they arrive, no re-sorting.
    for (const transaction of transactions) {
      const bucket = formatDayBucket(transaction.created_at)
      const existing = byDay.get(bucket) ?? []
      existing.push(transaction)
      byDay.set(bucket, existing)
    }
    return [...byDay.entries()]
  }, [transactions])

  const props: TransactionsViewProps = {
    filter,
    onFilterChange: (nextFilter) => {
      setFilter(nextFilter)
      const selectedCategory = categories.find((category) => category.id === categoryFilter)
      if (selectedCategory && nextFilter !== 'all' && selectedCategory.type !== nextFilter) setCategoryFilter('')
    },
    userFilter,
    onUserFilterChange: setUserFilter,
    users: usersQuery.data ?? (user ? [user] : []),
    categoryFilter,
    onCategoryFilterChange: setCategoryFilter,
    categories,
    nameFilter,
    onNameFilterChange: setNameFilter,
    dateFrom,
    onDateFromChange: setDateFrom,
    dateTo,
    onDateToChange: setDateTo,
    sortOrder,
    onSortOrderChange: setSortOrder,
    grouped,
    isEmpty: transactions.length === 0,
  }

  return <ResponsiveView mobile={Mobile} desktop={Desktop} props={props} />
}
