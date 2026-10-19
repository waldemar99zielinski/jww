import { IconChevronDown } from '@/components/icons'
import { LogoMark } from '@/components/LogoMark'
import { Section } from '@/components/Section'
import { TransactionRow } from '@/components/TransactionRow'
import { Button } from '@/components/ui/Button'
import { CategoryBreakdown } from './components/CategoryBreakdown'
import { HeroBalanceCard } from './components/HeroBalanceCard'
import type { HomeViewProps } from './types'

export default function HomeMobile({
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
    <div className="flex flex-col gap-5 px-4 pb-4 pt-6">
      <header className="flex items-start justify-between">
        <div className="flex items-center gap-2.5">
          <LogoMark size="sm" />
          <div>
            <p className="m-0 mb-0.5 text-xs text-muted">Good evening</p>
            <h2 className="m-0 text-xl font-bold">{displayName}</h2>
          </div>
        </div>
        <Button
          variant="custom"
          className="flex items-center gap-1.5 rounded-full border border-line bg-surface-sunken px-3 py-[7px] text-xs font-semibold text-ink-soft"
        >
          {monthLabel}
          <IconChevronDown className="h-3.5 w-3.5" />
        </Button>
      </header>

      <HeroBalanceCard
        balance={balance}
        income={totalIncome}
        expenses={totalExpense}
        currencyCode={currencyCode}
      />

      <Section title="Spending by category" action={{ label: 'See all', to: '/budgets' }}>
        <CategoryBreakdown items={breakdown} currencyCode={currencyCode} />
      </Section>

      <Section title="Recent transactions" action={{ label: 'See all', to: '/transactions' }}>
        {recent.length === 0 ? (
          <p className="m-0 text-xs text-muted">No transactions yet.</p>
        ) : (
          <ul className="flex flex-col">
            {recent.map((transaction) => (
              <TransactionRow key={transaction.id} transaction={transaction} subtitle={transaction.category_name} />
            ))}
          </ul>
        )}
      </Section>
    </div>
  )
}
