import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import type { SubTransactionDto } from '@/api/types'
import { SubTransactionModal } from './SubTransactionModal'

const item: SubTransactionDto = { id: 'item-1', name: 'Milk 2%', label: 'dairy', amount: 649 }

function renderModal(overrides: Partial<Parameters<typeof SubTransactionModal>[0]> = {}) {
  const props = {
    open: true,
    onClose: vi.fn(),
    onSave: vi.fn(),
    ...overrides,
  }
  return { ...render(<SubTransactionModal {...props} />), props }
}

describe('SubTransactionModal', () => {
  it('creates a row with a name, optional label and amount in minor units', async () => {
    const user = userEvent.setup()
    const { props } = renderModal()

    expect(screen.getByRole('dialog', { name: 'Add sub-transaction' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Add item' })).toBeDisabled()

    await user.type(screen.getByLabelText('Sub-transaction name'), 'Coffee beans')
    await user.type(screen.getByLabelText('Sub-transaction label'), 'pantry')
    await user.type(screen.getByLabelText('Sub-transaction amount'), '8,50')
    await user.click(screen.getByRole('button', { name: 'Add item' }))

    expect(props.onSave).toHaveBeenCalledWith({ name: 'Coffee beans', label: 'pantry', amount: 850 })
  })

  it('accepts a negative discount amount', async () => {
    const user = userEvent.setup()
    const { props } = renderModal()

    await user.type(screen.getByLabelText('Sub-transaction name'), 'Coupon')
    await user.type(screen.getByLabelText('Sub-transaction label'), 'discount')
    await user.type(screen.getByLabelText('Sub-transaction amount'), '-1,00')
    await user.click(screen.getByRole('button', { name: 'Add item' }))

    expect(props.onSave).toHaveBeenCalledWith({ name: 'Coupon', label: 'discount', amount: -100 })
  })

  it('edits the row it was opened for', async () => {
    const user = userEvent.setup()
    const { props } = renderModal({ item })

    expect(screen.getByRole('dialog', { name: 'Edit sub-transaction' })).toBeInTheDocument()
    expect(screen.getByLabelText('Sub-transaction name')).toHaveValue('Milk 2%')
    expect(screen.getByLabelText('Sub-transaction label')).toHaveValue('dairy')
    expect(screen.getByLabelText('Sub-transaction amount')).toHaveValue('6,49')

    await user.clear(screen.getByLabelText('Sub-transaction amount'))
    await user.type(screen.getByLabelText('Sub-transaction amount'), '7,00')
    await user.click(screen.getByRole('button', { name: 'Save' }))

    expect(props.onSave).toHaveBeenCalledWith({ name: 'Milk 2%', label: 'dairy', amount: 700 })
  })

  it('cancels without saving', async () => {
    const user = userEvent.setup()
    const { props } = renderModal({ item })

    await user.click(screen.getByRole('button', { name: 'Cancel' }))

    expect(props.onClose).toHaveBeenCalledTimes(1)
    expect(props.onSave).not.toHaveBeenCalled()
  })
})
