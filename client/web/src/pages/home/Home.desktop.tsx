import { Section } from '@/components/Section'
import { TransactionRow } from '@/components/TransactionRow'
import { CategoryBreakdown } from './components/CategoryBreakdown'
import { HeroBalanceCard } from './components/HeroBalanceCard'
import type { HomeViewProps } from './types'

export default function HomeDesktop({
  displayName,
  monthLabel,
  currencyCode,
  balance,
  totalIncome,
  totalExpense,
  breakdown,
  recent,
}: HomeViewProps) {
  return (
    <div className="flex flex-col gap-8">
      <header className="flex items-center justify-between">
        <div>
          <p className="m-0 mb-1 text-sm text-muted">Good evening</p>
          <h1 className="m-0 text-3xl font-bold">{displayName}</h1>
        </div>
        <span className="rounded-full border border-line bg-surface-sunken px-4 py-2 text-sm font-semibold text-ink-soft">
          {monthLabel}
        </span>
      </header>

      <HeroBalanceCard
        balance={balance}
        income={totalIncome}
        expenses={totalExpense}
        currencyCode={currencyCode}
      />

      <div className="grid grid-cols-2 gap-8">
        <Section title="Spending by category" action={{ label: 'See all', to: '/budgets' }}>
          <div className="rounded-2xl border border-line bg-surface p-5">
            <CategoryBreakdown items={breakdown} currencyCode={currencyCode} />
          </div>
        </Section>

        <Section title="Recent transactions" action={{ label: 'See all', to: '/transactions' }}>
          {recent.length === 0 ? (
            <p className="m-0 text-xs text-muted">No transactions yet.</p>
          ) : (
            <ul className="flex flex-col divide-y divide-line rounded-2xl border border-line bg-surface px-4">
              {recent.map((transaction) => (
                <TransactionRow key={transaction.id} transaction={transaction} subtitle={transaction.category_name} />
              ))}
            </ul>
          )}
        </Section>
      </div>
    </div>
  )
}
