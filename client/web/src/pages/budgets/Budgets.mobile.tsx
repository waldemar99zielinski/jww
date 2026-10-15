import { Section } from '@/components/Section'
import { StatusBanner } from '@/components/StatusBanner'
import { sampleBudgets, sampleGoal } from '@/data/sampleBudgets'
import { BudgetCard } from './components/BudgetCard'
import { GoalCard } from './components/GoalCard'

export default function BudgetsMobile() {
  return (
    <div className="flex flex-col gap-5 px-4 pb-4 pt-6">
      <StatusBanner />
      <h2 className="m-0 text-[19px] font-bold">Budgets &amp; goals</h2>

      <Section title="Monthly budgets">
        <div className="flex flex-col gap-2.5">
          {sampleBudgets.map((budget) => (
            <BudgetCard key={budget.categoryId} categoryId={budget.categoryId} spent={budget.spent} limit={budget.limit} />
          ))}
        </div>
      </Section>

      <Section title="Savings goals">
        <GoalCard label={sampleGoal.label} saved={sampleGoal.saved} target={sampleGoal.target} />
      </Section>
    </div>
  )
}
