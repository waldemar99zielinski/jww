import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import type { SubTransactionDto } from '@/api/types'
import { SubTransactionsPanel } from './SubTransactionsPanel'

const items: SubTransactionDto[] = [
  { id: 'item-1', name: 'Milk 2%', label: 'dairy', amount: 649 },
  { id: 'item-2', name: 'Sourdough bread', label: 'bakery', amount: 599 },
]

function renderPanel(overrides: Partial<Parameters<typeof SubTransactionsPanel>[0]> = {}) {
  const props = {
    items,
    attachmentTotal: 1248,
    expanded: false,
    onToggle: vi.fn(),
    onAdd: vi.fn(),
    onEdit: vi.fn(),
    onRemove: vi.fn(),
    ...overrides,
  }
  return { ...render(<SubTransactionsPanel {...props} />), props }
}

describe('SubTransactionsPanel', () => {
  it('shows only a count and total summary while collapsed', async () => {
    const user = userEvent.setup()
    const { props } = renderPanel()

    expect(screen.getByText('2 items · 12,48')).toBeInTheDocument()
    expect(screen.queryByText('Milk 2%')).not.toBeInTheDocument()

    await user.click(screen.getByRole('button', { name: 'Expand sub-transactions' }))
    expect(props.onToggle).toHaveBeenCalledTimes(1)
  })

  it('surfaces an attachment-total difference in the collapsed summary', () => {
    renderPanel({ attachmentTotal: 1400 })

    expect(screen.getByText(/attachment total 14,00/i)).toBeInTheDocument()
    expect(screen.getByText(/difference 1,52/i)).toBeInTheDocument()
  })

  it('expands to show editable rows, removal controls and the add action', async () => {
    const user = userEvent.setup()
    const { props } = renderPanel({ expanded: true })

    expect(screen.getByText('Milk 2%')).toBeInTheDocument()
    expect(screen.getByText('Sourdough bread')).toBeInTheDocument()

    await user.click(screen.getByRole('button', { name: 'Edit Milk 2%' }))
    expect(props.onEdit).toHaveBeenCalledWith('item-1')

    await user.click(screen.getByRole('button', { name: 'Remove Sourdough bread' }))
    expect(props.onRemove).toHaveBeenCalledWith('item-2')

    await user.click(screen.getByRole('button', { name: 'Add sub-transaction' }))
    expect(props.onAdd).toHaveBeenCalledTimes(1)
  })

  it('offers Add directly when there are no items', async () => {
    const user = userEvent.setup()
    const { props } = renderPanel({ items: [], attachmentTotal: undefined })

    expect(screen.getByText('No sub-transactions')).toBeInTheDocument()
    await user.click(screen.getByRole('button', { name: 'Add sub-transaction' }))

    expect(props.onAdd).toHaveBeenCalledTimes(1)
  })

  it('renders a read-only list when no action callbacks are supplied', () => {
    render(<SubTransactionsPanel items={items} expanded onToggle={vi.fn()} currencyCode="EUR" />)

    expect(screen.getByText('Milk 2%')).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: /^Edit / })).not.toBeInTheDocument()
    expect(screen.queryByRole('button', { name: /^Remove / })).not.toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Add sub-transaction' })).not.toBeInTheDocument()
    expect(screen.getAllByText(/12,48/).length).toBeGreaterThan(0)
  })
})
