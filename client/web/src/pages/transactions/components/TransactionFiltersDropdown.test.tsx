import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import { TransactionFiltersDropdown } from '@/pages/transactions/components/TransactionFiltersDropdown'
import type { TransactionCategoryDto, UserDto } from '@/api/types'

const users: UserDto[] = [{ id: 'user-1', username: 'waldemar', display_name: 'Waldemar', created_at: '2026-01-01T00:00:00Z' }]
const categories: TransactionCategoryDto[] = [
  { id: 'cat-groceries', type: 'expense', name: 'Groceries', active: true, sort_order: 0 },
  { id: 'cat-salary', type: 'income', name: 'Salary', active: true, sort_order: 0 },
]

function baseProps() {
  return {
    filter: 'all' as const,
    onFilterChange: vi.fn(),
    userFilter: '',
    onUserFilterChange: vi.fn(),
    users,
    categoryFilter: '',
    onCategoryFilterChange: vi.fn(),
    categories,
    nameFilter: '',
    onNameFilterChange: vi.fn(),
    dateFrom: '',
    onDateFromChange: vi.fn(),
    dateTo: '',
    onDateToChange: vi.fn(),
  }
}

describe('TransactionFiltersDropdown', () => {
  it('does not commit a draft edit until Apply is pressed', async () => {
    const user = userEvent.setup()
    const props = baseProps()
    render(<TransactionFiltersDropdown {...props} />)

    await user.click(screen.getByRole('button', { name: /filters/i }))
    await user.selectOptions(screen.getByLabelText('Type'), 'Income')

    expect(props.onFilterChange).not.toHaveBeenCalled()

    await user.click(screen.getByRole('button', { name: /^apply$/i }))

    expect(props.onFilterChange).toHaveBeenCalledWith('income')
  })

  it('closes the panel after applying', async () => {
    const user = userEvent.setup()
    render(<TransactionFiltersDropdown {...baseProps()} />)

    await user.click(screen.getByRole('button', { name: /filters/i }))
    expect(screen.getByLabelText('Type')).toBeInTheDocument()

    await user.click(screen.getByRole('button', { name: /^apply$/i }))

    expect(screen.queryByLabelText('Type')).not.toBeInTheDocument()
  })

  it('"Reset" clears the draft fields without committing them', async () => {
    const user = userEvent.setup()
    const props = baseProps()
    render(<TransactionFiltersDropdown {...props} />)

    await user.click(screen.getByRole('button', { name: /filters/i }))
    await user.type(screen.getByLabelText('Name'), 'coffee')
    expect(screen.getByLabelText('Name')).toHaveValue('coffee')

    await user.click(screen.getByRole('button', { name: /^reset$/i }))

    expect(screen.getByLabelText('Name')).toHaveValue('')
    expect(props.onNameFilterChange).not.toHaveBeenCalled()
  })

  it('shows the active filter count on the toggle button', () => {
    render(<TransactionFiltersDropdown {...baseProps()} filter="income" nameFilter="coffee" />)

    expect(screen.getByRole('button', { name: /filters/i })).toHaveTextContent('2')
  })
})
