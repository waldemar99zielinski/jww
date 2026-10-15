import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { BudgetCard } from './BudgetCard'

describe('BudgetCard', () => {
  it('renders the category label without an over-budget flag when under the limit', () => {
    render(<BudgetCard categoryId="groceries" spent={860} limit={1000} />)

    expect(screen.getByText('Groceries')).toBeInTheDocument()
    expect(screen.queryByText('Over budget')).not.toBeInTheDocument()
  })

  it('flags an over-budget category', () => {
    render(<BudgetCard categoryId="entertainment" spent={410} limit={350} />)

    expect(screen.getByText('Fun')).toBeInTheDocument()
    expect(screen.getByText('Over budget')).toBeInTheDocument()
  })
})
