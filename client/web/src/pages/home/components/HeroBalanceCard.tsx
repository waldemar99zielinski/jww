import { formatAmount, formatSignedAmount } from '@/lib/currency'

interface HeroBalanceCardProps {
  balance: number
  income: number
  expenses: number
}

export function HeroBalanceCard({ balance, income, expenses }: HeroBalanceCardProps) {
  return (
    <div className="hero-card relative overflow-hidden rounded-[22px] px-5 pb-[18px] pt-5 text-on-accent shadow-[var(--shadow-sm)]">
      <svg
        className="pointer-events-none absolute -right-1.5 -top-1 h-[92px] w-[170px] opacity-55"
        viewBox="0 0 170 92"
        preserveAspectRatio="none"
        aria-hidden="true"
      >
        <path
          d="M0,70 L15,64 L30,68 L45,50 L60,55 L75,38 L90,42 L105,26 L120,30 L135,16 L150,20 L170,6"
          fill="none"
          stroke="currentColor"
          strokeWidth={2}
          strokeLinecap="round"
          strokeLinejoin="round"
        />
        <circle cx="170" cy="6" r="3.5" fill="currentColor" />
      </svg>

      <p className="relative m-0 text-xs font-semibold opacity-80">Balance</p>
      <p className="relative m-0 mb-3.5 mt-1 font-mono text-[32px] font-semibold tracking-tight tabular-nums">
        {formatAmount(balance)}
      </p>
      <div className="relative flex gap-5 text-[13px]">
        <span className="flex items-center gap-1.5 opacity-90">
          <span className="h-[7px] w-[7px] rounded-full bg-current" />
          Income <strong className="font-mono font-semibold tabular-nums">{formatSignedAmount(income, 'income')}</strong>
        </span>
        <span className="flex items-center gap-1.5 opacity-90">
          <span className="h-[7px] w-[7px] rounded-full bg-current" />
          Expenses <strong className="font-mono font-semibold tabular-nums">{formatSignedAmount(expenses, 'expense')}</strong>
        </span>
      </div>
    </div>
  )
}
