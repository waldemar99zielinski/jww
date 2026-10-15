import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { TransactionFiltersSummary } from '@/pages/transactions/components/TransactionFiltersSummary'
import type { TransactionCategoryDto, UserDto } from '@/api/types'

const users: UserDto[] = [{ id: 'user-1', username: 'waldemar', display_name: 'Waldemar', created_at: '2026-01-01T00:00:00Z' }]
const categories: TransactionCategoryDto[] = [
  { id: 'cat-groceries', type: 'expense', name: 'Groceries', active: true, sort_order: 0 },
]

describe('TransactionFiltersSummary', () => {
  it('renders nothing when no filters are active', () => {
    const { container } = render(
      <TransactionFiltersSummary
        filter="all"
        userFilter=""
        users={users}
        categoryFilter=""
        categories={categories}
        nameFilter=""
        dateFrom=""
        dateTo=""
      />,
    )

    expect(container).toBeEmptyDOMElement()
  })

  it('summarizes the active filters', () => {
    render(
      <TransactionFiltersSummary
        filter="expense"
        userFilter="user-1"
        users={users}
        categoryFilter="cat-groceries"
        categories={categories}
        nameFilter="coffee"
        dateFrom="2026-10-01"
        dateTo="2026-10-31"
      />,
    )

    expect(screen.getByText('Showing:')).toBeInTheDocument()
    expect(screen.getByText('Expenses')).toBeInTheDocument()
    expect(screen.getByText('· Waldemar')).toBeInTheDocument()
    expect(screen.getByText('· Groceries')).toBeInTheDocument()
    expect(screen.getByText('· coffee')).toBeInTheDocument()
    expect(screen.getByText('· from 2026-10-01')).toBeInTheDocument()
    expect(screen.getByText('· to 2026-10-31')).toBeInTheDocument()
  })
})
