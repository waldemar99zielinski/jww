import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { CategoryBreakdown } from '@/pages/home/components/CategoryBreakdown'
import { formatAmount } from '@/lib/currency'
import { normalizeText } from '@/test/normalizeText'

describe('CategoryBreakdown', () => {
  it('renders an empty state when there are no items', () => {
    render(<CategoryBreakdown items={[]} />)

    expect(screen.getByText('No expenses yet this month.')).toBeInTheDocument()
  })

  it('renders each item with its category name and formatted amount', () => {
    render(
      <CategoryBreakdown
        items={[
          { categoryName: 'Groceries', amount: 42.5, percent: 100 },
          { categoryName: 'Transport', amount: 10, percent: 23.5 },
        ]}
      />,
    )

    expect(screen.getByText('Groceries')).toBeInTheDocument()
    expect(screen.getByText(normalizeText(formatAmount(42.5)))).toBeInTheDocument()
    expect(screen.getByText('Transport')).toBeInTheDocument()
    expect(screen.getByText(normalizeText(formatAmount(10)))).toBeInTheDocument()
  })
})
