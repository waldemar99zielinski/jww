import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import type { TransactionDto } from '@/api/types'
import { TransactionRow } from '@/components/TransactionRow'

const baseTransaction: TransactionDto = {
  id: 1,
  type: 'expense',
  amount: '42.50',
  currency_code: 'EUR',
  category_id: 'cat-groceries',
  category_name: 'Groceries',
  user_id: 'user-1',
  source: 'manual',
  name: 'Weekly shop',
  created_at: '2026-10-10T09:30:00Z',
  description: '',
}

describe('TransactionRow', () => {
  it('renders an expense with the muted (non-bold) amount style', () => {
    const { container } = render(
      <ul>
        <TransactionRow transaction={baseTransaction} subtitle="Today" />
      </ul>,
    )

    expect(screen.getByText('Weekly shop')).toBeInTheDocument()
    expect(screen.getByText('Today')).toBeInTheDocument()
    expect(container.querySelector('.text-ink-soft')).toBeInTheDocument()
  })

  it('renders income with the bold amount style', () => {
    const { container } = render(
      <ul>
        <TransactionRow
          transaction={{ ...baseTransaction, type: 'income', name: 'October salary', category_name: 'Salary' }}
          subtitle="1 Oct"
        />
      </ul>,
    )

    expect(container.querySelector('.font-bold.text-ink')).toBeInTheDocument()
  })

  it('falls back to the generic tag icon/color for an unrecognized category name', () => {
    const { container } = render(
      <ul>
        <TransactionRow transaction={{ ...baseTransaction, category_name: 'Some Brand New Category' }} subtitle="Today" />
      </ul>,
    )

    expect(screen.getByText('Weekly shop')).toBeInTheDocument()
    expect(container.querySelector('svg')).toBeInTheDocument()
  })

  it('falls back to the category name as the title when name/description are empty', () => {
    render(
      <ul>
        <TransactionRow transaction={{ ...baseTransaction, name: '', description: '' }} subtitle="Today" />
      </ul>,
    )

    expect(screen.getByText('Groceries')).toBeInTheDocument()
  })
})
