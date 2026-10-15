import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { HeroBalanceCard } from '@/pages/home/components/HeroBalanceCard'
import { formatAmount, formatSignedAmount } from '@/lib/currency'
import { normalizeText } from '@/test/normalizeText'

describe('HeroBalanceCard', () => {
  it('renders the formatted balance, income, and expenses', () => {
    render(<HeroBalanceCard balance={2457.5} income={2500} expenses={42.5} />)

    expect(screen.getByText(normalizeText(formatAmount(2457.5)))).toBeInTheDocument()
    expect(screen.getByText(normalizeText(formatSignedAmount(2500, 'income')))).toBeInTheDocument()
    expect(screen.getByText(normalizeText(formatSignedAmount(42.5, 'expense')))).toBeInTheDocument()
  })

  it('renders a negative balance using the plain (unsigned) formatter', () => {
    render(<HeroBalanceCard balance={-50} income={0} expenses={50} />)

    expect(screen.getByText(normalizeText(formatAmount(-50)))).toBeInTheDocument()
  })
})
