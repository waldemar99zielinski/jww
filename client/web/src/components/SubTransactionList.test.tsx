import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import type { SubTransactionDto } from '@/api/types'
import { formatAmount } from '@/lib/currency'
import { normalizeText } from '@/test/normalizeText'
import { SubTransactionList } from './SubTransactionList'

const items: SubTransactionDto[] = [
  { id: 'item-1', name: 'Milk 2%', label: 'dairy', amount: 649 },
  { id: 'item-2', name: 'Sourdough bread', label: 'bakery', amount: 599 },
]

describe('SubTransactionList', () => {
  it('renders each row with its name and label', () => {
    render(<SubTransactionList items={items} onAdd={vi.fn()} onUpdate={vi.fn()} onRemove={vi.fn()} />)

    expect(screen.getByText('Milk 2%')).toBeInTheDocument()
    expect(screen.getByText('dairy')).toBeInTheDocument()
    expect(screen.getByText('Sourdough bread')).toBeInTheDocument()
    expect(screen.getByText('bakery')).toBeInTheDocument()
  })

  it('renders the running total of what the rows add up to', () => {
    render(<SubTransactionList items={items} onAdd={vi.fn()} onUpdate={vi.fn()} onRemove={vi.fn()} />)

    expect(screen.getByText('Total')).toBeInTheDocument()
    expect(screen.getByText(normalizeText(formatAmount(12.48)))).toBeInTheDocument()
  })

  it('is fully controlled — with no items it shows only the add affordance', () => {
    render(<SubTransactionList items={[]} onAdd={vi.fn()} onUpdate={vi.fn()} onRemove={vi.fn()} />)

    expect(screen.getByRole('button', { name: 'Add item' })).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: /^Edit / })).not.toBeInTheDocument()
    // The empty list totals to zero.
    expect(screen.getByText(normalizeText(formatAmount(0)))).toBeInTheDocument()
  })

  it('adds a row with its amount converted from the display string to whole minor units', async () => {
    const user = userEvent.setup()
    const onAdd = vi.fn()
    render(<SubTransactionList items={items} onAdd={onAdd} onUpdate={vi.fn()} onRemove={vi.fn()} />)

    await user.click(screen.getByRole('button', { name: 'Add item' }))
    await user.type(screen.getByLabelText('New item name'), 'Coffee beans')
    await user.type(screen.getByLabelText('New item label'), 'pantry')
    await user.type(screen.getByLabelText('New item amount'), '8,50')
    await user.click(screen.getByRole('button', { name: 'Add item' }))

    expect(onAdd).toHaveBeenCalledWith({ name: 'Coffee beans', label: 'pantry', amount: 850 })
  })

  it('keeps inline item inputs at 16px on mobile so Safari does not zoom on focus', async () => {
    const user = userEvent.setup()
    render(<SubTransactionList items={items} onAdd={vi.fn()} onUpdate={vi.fn()} onRemove={vi.fn()} />)

    await user.click(screen.getByRole('button', { name: 'Add item' }))

    for (const label of ['New item name', 'New item label', 'New item amount']) {
      const input = screen.getByLabelText(label)
      expect(input).toHaveClass('text-base')
      expect(input).not.toHaveClass('text-[13px]')
    }
  })

  it('edits a row, passing the changed fields with the amount in whole minor units', async () => {
    const user = userEvent.setup()
    const onUpdate = vi.fn()
    render(<SubTransactionList items={items} onAdd={vi.fn()} onUpdate={onUpdate} onRemove={vi.fn()} />)

    await user.click(screen.getByRole('button', { name: 'Edit Milk 2%' }))

    const nameInput = screen.getByLabelText('Item name for Milk 2%')
    const amountInput = screen.getByLabelText('Item amount for Milk 2%')
    await user.clear(nameInput)
    await user.type(nameInput, 'Milk 3%')
    await user.clear(amountInput)
    await user.type(amountInput, '7,00')
    await user.click(screen.getByRole('button', { name: 'Save' }))

    expect(onUpdate).toHaveBeenCalledWith('item-1', { name: 'Milk 3%', label: 'dairy', amount: 700 })
  })

  it('removes a row by its id', async () => {
    const user = userEvent.setup()
    const onRemove = vi.fn()
    render(<SubTransactionList items={items} onAdd={vi.fn()} onUpdate={vi.fn()} onRemove={onRemove} />)

    await user.click(screen.getByRole('button', { name: 'Remove Sourdough bread' }))

    expect(onRemove).toHaveBeenCalledWith('item-2')
  })

  it('does not add a row until it has a name and a non-zero amount', async () => {
    const user = userEvent.setup()
    const onAdd = vi.fn()
    render(<SubTransactionList items={items} onAdd={onAdd} onUpdate={vi.fn()} onRemove={vi.fn()} />)

    await user.click(screen.getByRole('button', { name: 'Add item' }))
    // Amount but no name.
    await user.type(screen.getByLabelText('New item amount'), '8,50')
    expect(screen.getByRole('button', { name: 'Add item' })).toBeDisabled()

    await user.type(screen.getByLabelText('New item name'), 'Coffee beans')
    await user.click(screen.getByRole('button', { name: 'Add item' }))

    expect(onAdd).toHaveBeenCalledTimes(1)
  })

  it('accepts a negative discount amount and includes it in the total', async () => {
    const user = userEvent.setup()
    const onAdd = vi.fn()
    render(<SubTransactionList items={items} onAdd={onAdd} onUpdate={vi.fn()} onRemove={vi.fn()} />)

    await user.click(screen.getByRole('button', { name: 'Add item' }))
    await user.type(screen.getByLabelText('New item name'), 'Coupon')
    await user.type(screen.getByLabelText('New item label'), 'discount')
    await user.type(screen.getByLabelText('New item amount'), '-1,00')
    await user.click(screen.getByRole('button', { name: 'Add item' }))

    expect(onAdd).toHaveBeenCalledWith({ name: 'Coupon', label: 'discount', amount: -100 })
  })

  it('renders a negative line item and subtracts it from the running total', () => {
    render(
      <SubTransactionList
        items={[...items, { id: 'item-3', name: 'Coupon', label: 'discount', amount: -100 }]}
        onAdd={vi.fn()}
        onUpdate={vi.fn()}
        onRemove={vi.fn()}
      />,
    )

    expect(screen.getByText('-1,00')).toBeInTheDocument()
    expect(screen.getByText(normalizeText(formatAmount(11.48)))).toBeInTheDocument()
  })

  it('locks the row controls while the parent is busy persisting a change', () => {
    render(<SubTransactionList items={items} onAdd={vi.fn()} onUpdate={vi.fn()} onRemove={vi.fn()} busy />)

    expect(screen.getByRole('button', { name: 'Edit Milk 2%' })).toBeDisabled()
    expect(screen.getByRole('button', { name: 'Remove Milk 2%' })).toBeDisabled()
    expect(screen.getByRole('button', { name: /Add item/ })).toBeDisabled()
  })
})
