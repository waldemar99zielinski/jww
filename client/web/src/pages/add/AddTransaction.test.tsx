import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import type { CurrencyDto, SubTransactionDto, TransactionCategoryDto, UserDto } from '@/api/types'
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
const subTransactionItems: SubTransactionDto[] = [
  { id: 'item-1', name: 'Milk 2%', label: 'dairy', amount: 649 },
  { id: 'item-2', name: 'Sourdough bread', label: 'bakery', amount: 599 },
]

function buildProps(overrides: Partial<AddTransactionViewProps> = {}): AddTransactionViewProps {
  return {
    type: 'expense',
    onSelectType: vi.fn(),
    amount: '42,50',
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
    attachments: [],
    extractionStatus: 'idle',
    extractionSummary: undefined,
    canScanAttachments: false,
    canUndoExtraction: false,
    attachmentPickerOpen: false,
    onOpenAttachment: vi.fn(),
    onCloseAttachment: vi.fn(),
    onPickAttachment: vi.fn(),
    onRemoveAttachment: vi.fn(),
    onPreviewAttachment: vi.fn(),
    onScanAttachments: vi.fn(),
    onCancelExtraction: vi.fn(),
    onUndoExtraction: vi.fn(),
    subTransactionItems: [],
    subTransactionsExpanded: false,
    onToggleSubTransactions: vi.fn(),
    onReviewSubTransactions: vi.fn(),
    onAddSubTransaction: vi.fn(),
    onEditSubTransaction: vi.fn(),
    onRemoveSubTransaction: vi.fn(),
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
    expect(screen.getByLabelText('Amount')).toHaveValue('42,50')
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

  it('renders every attachment row from props once attachment details are expanded', async () => {
    const user = userEvent.setup()
    render(
      <Variant
        {...buildProps({
          attachments: [
            { id: 'row-1', name: 'attachment.jpg', status: 'ready', attachmentId: 'a1', isImage: true },
            { id: 'row-2', name: 'note.txt', status: 'ready', attachmentId: 'a2' },
          ],
        })}
      />,
    )

    await user.click(screen.getByRole('button', { name: 'Show attachment details' }))

    expect(screen.getByText(/attachment\.jpg/)).toBeInTheDocument()
    expect(screen.getByText(/note\.txt/)).toBeInTheDocument()
    expect(screen.getAllByText('Ready to scan')).toHaveLength(1)
  })

  it('shows the extracted summary collapsed, then reveals the rows when expanded', () => {
    const props = buildProps({
      attachments: [{ id: 'row-1', name: 'attachment.jpg', status: 'ready', attachmentId: 'a1', isImage: true }],
      extractionStatus: 'success',
      extractionSummary: {
        attachmentIds: ['a1'],
        itemCount: 2,
        attachmentTotal: 1248,
        warnings: [],
      },
      subTransactionItems,
      amount: '12,48',
      amountReadOnly: true,
      amountNote: '= sum of 2 items',
    })
    const { rerender } = render(<Variant {...props} />)

    expect(screen.getByText('2 items added')).toBeInTheDocument()
    let itemsCard = screen.getByRole('region', { name: 'Sub-transactions' })
    expect(within(itemsCard).getByText('2 items · 12,48')).toBeInTheDocument()
    expect(within(itemsCard).queryByText('Milk 2%')).not.toBeInTheDocument()

    rerender(<Variant {...props} subTransactionsExpanded />)

    itemsCard = screen.getByRole('region', { name: 'Sub-transactions' })
    expect(within(itemsCard).getByText('Milk 2%')).toBeInTheDocument()
    expect(within(itemsCard).getByText('Sourdough bread')).toBeInTheDocument()
    expect(screen.getByText('= sum of 2 items')).toBeInTheDocument()
  })

  it('keeps sub-transactions collapsed by default and expands them from the summary row', () => {
    const { rerender } = render(<Variant {...buildProps({ subTransactionItems })} />)

    expect(screen.getByText('2 items · 12,48')).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Edit Milk 2%' })).not.toBeInTheDocument()

    rerender(<Variant {...buildProps({ subTransactionItems, subTransactionsExpanded: true })} />)

    expect(screen.getByRole('button', { name: 'Edit Milk 2%' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Remove Sourdough bread' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Add sub-transaction' })).toBeInTheDocument()
  })
})
