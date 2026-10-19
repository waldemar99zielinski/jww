import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { HeroBalanceCard } from '@/pages/home/components/HeroBalanceCard'
import { formatAmount, formatSignedAmount } from '@/lib/currency'
import { normalizeText } from '@/test/normalizeText'

describe('HeroBalanceCard', () => {
  it('renders the formatted balance, income, and expenses', () => {
    render(<HeroBalanceCard balance={2457.5} income={2500} expenses={42.5} currencyCode="EUR" />)

    expect(screen.getByText(normalizeText(formatAmount(2457.5, 'EUR')))).toBeInTheDocument()
    expect(screen.getByText(normalizeText(formatSignedAmount(2500, 'income', 'EUR')))).toBeInTheDocument()
    expect(screen.getByText(normalizeText(formatSignedAmount(42.5, 'expense', 'EUR')))).toBeInTheDocument()
  })

  it('renders a negative balance using the plain (unsigned) formatter', () => {
    render(<HeroBalanceCard balance={-50} income={0} expenses={50} currencyCode="EUR" />)

    expect(screen.getByText(normalizeText(formatAmount(-50, 'EUR')))).toBeInTheDocument()
  })
})
