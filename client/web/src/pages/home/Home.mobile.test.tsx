import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { describe, expect, it } from 'vitest'
import HomeMobile from '@/pages/home/Home.mobile'
import type { HomeViewProps } from '@/pages/home/types'
import { fixtureTransactions } from '@/test/fixtures'

const baseProps: HomeViewProps = {
  displayName: 'Waldemar',
  monthLabel: 'October 2026',
  balance: 2457.5,
  totalIncome: 2500,
  totalExpense: 42.5,
  breakdown: [{ categoryName: 'Groceries', amount: 42.5, percent: 100 }],
  recent: fixtureTransactions,
}

function renderMobile(props: Partial<HomeViewProps> = {}) {
  return render(
    <MemoryRouter>
      <HomeMobile {...baseProps} {...props} />
    </MemoryRouter>,
  )
}

describe('HomeMobile', () => {
  it('renders the display name, month label, and recent transactions', () => {
    renderMobile()

    expect(screen.getByRole('heading', { name: 'Waldemar' })).toBeInTheDocument()
    expect(screen.getByText('October 2026')).toBeInTheDocument()
    expect(screen.getByText('Weekly shop')).toBeInTheDocument()
    expect(screen.getByText('October salary')).toBeInTheDocument()
  })

  it('shows an empty state when there are no recent transactions', () => {
    renderMobile({ recent: [] })

    expect(screen.getByText('No transactions yet.')).toBeInTheDocument()
  })
})
