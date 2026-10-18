import { lazy, useMemo, useState } from 'react'
import { ResponsiveView } from '@/components/ResponsiveView'
import { useAuth } from '@/hooks/useAuth'
import { useDefaultCurrency } from '@/hooks/useDefaultCurrency'
import { useSummary } from '@/hooks/useSummary'
import { useTransactions } from '@/hooks/useTransactions'
import { startOfMonth, toISODate } from '@/lib/dates'
import type { HomeViewProps } from './types'

const Mobile = lazy(() => import('./Home.mobile'))
const Desktop = lazy(() => import('./Home.desktop'))

export default function HomePage() {
  const { user } = useAuth()
  const { currency } = useDefaultCurrency()

  const [now] = useState(() => new Date())
  const periodStart = toISODate(startOfMonth(now))
  const periodEnd = toISODate(now)
  const monthLabel = now.toLocaleDateString('en-GB', { month: 'long', year: 'numeric' })

  const summaryQuery = useSummary({
    period: 'month',
    period_start: periodStart,
    period_end: periodEnd,
    currency_code: currency?.code ?? 'PLN',
  })

  const transactionsQuery = useTransactions({
    user_id: user?.id,
    created_from: `${periodStart}T00:00:00Z`,
    created_to: `${periodEnd}T23:59:59Z`,
  })
  const transactions = useMemo(() => transactionsQuery.data ?? [], [transactionsQuery.data])
  const recent = useMemo(() => [...transactions].reverse().slice(0, 3), [transactions])

  const breakdown = useMemo(() => {
    const totals = new Map<string, number>()
    for (const t of transactions) {
      if (t.type !== 'expense') continue
      totals.set(t.category_name, (totals.get(t.category_name) ?? 0) + Number(t.amount))
    }
    const max = Math.max(0, ...totals.values())
    return [...totals.entries()]
      .map(([categoryName, amount]) => ({
        categoryName,
        amount,
        percent: max > 0 ? (amount / max) * 100 : 0,
      }))
      .sort((a, b) => b.amount - a.amount)
  }, [transactions])

  const summary = summaryQuery.data
  const totalIncome = Number(summary?.total_income ?? 0)
  const totalExpense = Number(summary?.total_expense ?? 0)

  const props: HomeViewProps = {
    displayName: user?.display_name ?? 'there',
    monthLabel,
    balance: totalIncome - totalExpense,
    totalIncome,
    totalExpense,
    breakdown,
    recent,
  }

  return <ResponsiveView mobile={Mobile} desktop={Desktop} props={props} />
}
