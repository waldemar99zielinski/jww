import { Section } from '@/components/Section'
import { StatusBanner } from '@/components/StatusBanner'
import { sampleBudgets, sampleGoal } from '@/data/sampleBudgets'
import { BudgetCard } from './components/BudgetCard'
import { GoalCard } from './components/GoalCard'

export default function BudgetsDesktop() {
  return (
    <div className="flex flex-col gap-6">
      <StatusBanner />
      <h1 className="m-0 text-2xl font-bold">Budgets &amp; goals</h1>

      <Section title="Monthly budgets">
        <div className="grid grid-cols-2 gap-4">
          {sampleBudgets.map((budget) => (
            <BudgetCard key={budget.categoryId} categoryId={budget.categoryId} spent={budget.spent} limit={budget.limit} />
          ))}
        </div>
      </Section>

      <Section title="Savings goals">
        <div className="grid grid-cols-2 gap-4">
          <GoalCard label={sampleGoal.label} saved={sampleGoal.saved} target={sampleGoal.target} />
        </div>
      </Section>
    </div>
  )
}
