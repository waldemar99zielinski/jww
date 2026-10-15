import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import BudgetsPage from '@/pages/budgets'

describe('BudgetsPage', () => {
  it('shows the "not built yet" status banner and renders sample budgets/goals', async () => {
    render(<BudgetsPage />)

    expect(await screen.findByText('Proposed — not built yet')).toBeInTheDocument()
    expect(screen.getByText('Groceries')).toBeInTheDocument()
    expect(screen.getByText('Emergency fund')).toBeInTheDocument()
  })
})
