import { render, screen } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import type { CurrencyDto, TransactionCategoryDto, UserDto } from '@/api/types'
import AddTransactionDesktop from './AddTransaction.desktop'
import AddTransactionMobile from './AddTransaction.mobile'
import type { AddTransactionViewProps } from './types'

/**
 * `AddTransaction.desktop.tsx`/`.mobile.tsx` are documented (see pages/add/README.md) as the most
 * duplicated pair in the codebase — same props in, same fields/values out, differing only in
 * layout. This file renders both from ONE fixture and asserts they agree, so a silent content
 * drift between the two (not just a layout tweak) fails loudly.
 */
const categories: TransactionCategoryDto[] = [
  { id: 'cat-groceries', type: 'expense', name: 'Groceries', active: true, sort_order: 0 },
]
const currencies: CurrencyDto[] = [{ code: 'EUR', name: 'Euro', is_base: true }]
const users: UserDto[] = [{ id: 'user-1', username: 'a', display_name: 'Ada', created_at: '' }]

function buildProps(overrides: Partial<AddTransactionViewProps> = {}): AddTransactionViewProps {
  return {
    type: 'expense',
    onSelectType: vi.fn(),
    amount: '42.50',
    onAmountChange: vi.fn(),
    currencyCode: 'EUR',
    currencies,
    onCurrencyChange: vi.fn(),
    hasCurrency: true,
    categories,
    categoryId: 'cat-groceries',
    onSelectCategory: vi.fn(),
    name: 'Weekly shop',
    onNameChange: vi.fn(),
    description: 'Big shop',
    onDescriptionChange: vi.fn(),
    date: '2026-10-10',
    time: '09:30',
    onDateChange: vi.fn(),
    onTimeChange: vi.fn(),
    onUseToday: vi.fn(),
    dateLabel: 'Today',
    dateOpen: false,
    onToggleDate: vi.fn(),
    receipt: null,
    receiptPickerOpen: false,
    onOpenReceipt: vi.fn(),
    onCloseReceipt: vi.fn(),
    onPickReceipt: vi.fn(),
    onRemoveReceipt: vi.fn(),
    users,
    userId: 'user-1',
    onUserChange: vi.fn(),
    showUserPicker: false,
    moreOpen: false,
    onToggleMore: vi.fn(),
    canSave: true,
    saving: false,
    saveError: false,
    onSave: vi.fn(),
    onClose: vi.fn(),
    ...overrides,
  }
}

describe.each([
  ['desktop', AddTransactionDesktop],
  ['mobile', AddTransactionMobile],
])('AddTransaction.%s', (_label, Variant) => {
  it('renders the name, description, category, and amount from props', () => {
    render(<Variant {...buildProps()} />)

    expect(screen.getByLabelText('Transaction name')).toHaveValue('Weekly shop')
    expect(screen.getByLabelText('Transaction description')).toHaveValue('Big shop')
    expect(screen.getByLabelText('Amount')).toHaveValue('42.50')
    expect(screen.getByRole('button', { name: /groceries/i })).toBeInTheDocument()
  })

  it('shows the "no currency" hint when hasCurrency is false', () => {
    render(<Variant {...buildProps({ hasCurrency: false })} />)
    expect(screen.getByText(/no currency configured yet/i)).toBeInTheDocument()
  })

  it('disables the save button when canSave is false, and labels it "Saving…" while saving', () => {
    const { rerender } = render(<Variant {...buildProps({ canSave: false })} />)
    expect(screen.getByRole('button', { name: /save transaction/i })).toBeDisabled()

    rerender(<Variant {...buildProps({ saving: true })} />)
    expect(screen.getByRole('button', { name: /saving/i })).toBeDisabled()
  })

  it('shows a save error message when saveError is true', () => {
    render(<Variant {...buildProps({ saveError: true })} />)
    expect(screen.getByText(/couldn't save/i)).toBeInTheDocument()
  })
})
