import { fireEvent, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import type { TransactionAttachmentItem, TransactionDetailEditProps, TransactionDetailViewProps } from './types'
import TransactionDetailDesktop from './TransactionDetail.desktop'
import TransactionDetailMobile from './TransactionDetail.mobile'

const imageAttachment: TransactionAttachmentItem = {
  id: 'attachment-1',
  filename: 'attachment.jpg',
  mediaType: 'image/jpeg',
  byteSize: 20480,
  previewUrl: '/api/v1/attachment-content?grant=grant-1',
}

const fileAttachment: TransactionAttachmentItem = {
  id: 'attachment-2',
  filename: 'note.txt',
  mediaType: 'text/plain',
  byteSize: 512,
  previewUrl: null,
}

const editProps: TransactionDetailEditProps = {
  type: 'expense',
  amount: '42,50',
  currencyCode: 'EUR',
  currencies: [{ code: 'EUR', name: 'Euro', is_base: true }],
  hasCurrency: true,
  categories: [{ id: 'cat-groceries', type: 'expense', name: 'Groceries', active: true, sort_order: 0 }],
  categoryId: 'cat-groceries',
  name: 'Weekly shop',
  description: 'Milk, bread and coffee',
  date: '2026-10-10',
  time: '09:30',
  users: [],
  userId: 'user-1',
  showUserPicker: false,
  moreOpen: false,
  canSave: true,
  saving: false,
  saveError: false,
  isDirty: false,
  onSelectType: () => {},
  onAmountChange: () => {},
  onCurrencyChange: () => {},
  onSelectCategory: () => {},
  onNameChange: () => {},
  onDescriptionChange: () => {},
  onDateChange: () => {},
  onTimeChange: () => {},
  onUseToday: () => {},
  onUserChange: () => {},
  onToggleMore: () => {},
  onSave: () => {},
  onReset: () => {},
  onCancel: () => {},
}

const baseProps: TransactionDetailViewProps = {
  title: 'Weekly shop',
  categoryName: 'Groceries',
  amount: '−42,50 zł',
  isIncome: false,
  items: [
    { label: 'Type', value: 'Expense' },
    { label: 'Category', value: 'Groceries' },
    { label: 'Date', value: '10 October 2026 · 09:30' },
  ],
  subTransactions: [
    { id: 'sub-1', name: 'Milk 2%', label: 'dairy', amount: '6,49 zł' },
    { id: 'sub-2', name: 'Sourdough bread', label: 'bakery', amount: '5,99 zł' },
  ],
  attachments: [imageAttachment, fileAttachment],
  attachmentsUploading: false,
  attachmentUploadError: false,
  downloadError: false,
  isLoading: false,
  isError: false,
  isEditing: false,
  onStartEdit: () => {},
  onAddAttachment: () => {},
  onRequestDelete: () => {},
  edit: editProps,
  onPreview: () => {},
  onOpenActions: () => {},
  onClose: () => {},
}

const variants = [
  { name: 'desktop', View: TransactionDetailDesktop },
  { name: 'mobile', View: TransactionDetailMobile },
]

describe.each(variants)('TransactionDetail $name', ({ View }) => {
  it('renders the hero, every field, and the attachments', () => {
    render(<View {...baseProps} />)

    expect(screen.getByText('Weekly shop')).toBeInTheDocument()
    expect(screen.getByText('−42,50 zł')).toBeInTheDocument()
    expect(screen.getByText('Expense')).toBeInTheDocument()
    expect(screen.getByText('attachment.jpg')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Close' })).toBeInTheDocument()
  })

  it('lists the line items and asks to remove the transaction', async () => {
    const user = userEvent.setup()
    const onRequestDelete = vi.fn()
    render(<View {...baseProps} onRequestDelete={onRequestDelete} />)

    expect(screen.getByRole('heading', { name: 'Items' })).toBeInTheDocument()
    expect(screen.getByText('Milk 2%')).toBeInTheDocument()
    expect(screen.getByText('dairy')).toBeInTheDocument()
    expect(screen.getByText('6,49 zł')).toBeInTheDocument()
    expect(screen.getByText('Sourdough bread')).toBeInTheDocument()

    await user.click(screen.getByRole('button', { name: 'Remove transaction' }))

    expect(onRequestDelete).toHaveBeenCalledTimes(1)
  })

  it('offers an Add attachment action, including the empty and error states', async () => {
    const user = userEvent.setup()
    const onAddAttachment = vi.fn()
    render(
      <View
        {...baseProps}
        attachments={[]}
        attachmentUploadError
        onAddAttachment={onAddAttachment}
      />,
    )

    expect(screen.getByText('No attachments yet.')).toBeInTheDocument()
    expect(screen.getByText(/couldn't upload that file/i)).toBeInTheDocument()

    await user.click(screen.getByRole('button', { name: 'Add attachment' }))

    expect(onAddAttachment).toHaveBeenCalledTimes(1)
  })

  it('disables Add attachment while an upload is in flight', () => {
    render(<View {...baseProps} attachments={[]} attachmentsUploading />)

    expect(screen.getByRole('button', { name: 'Uploading…' })).toBeDisabled()
  })

  it('offers an Edit button next to Close and asks to start editing', async () => {
    const user = userEvent.setup()
    const onStartEdit = vi.fn()
    render(<View {...baseProps} onStartEdit={onStartEdit} />)

    await user.click(screen.getByRole('button', { name: 'Edit' }))

    expect(onStartEdit).toHaveBeenCalledTimes(1)
  })

  it('hides the Edit button while already editing', () => {
    render(<View {...baseProps} isEditing />)

    expect(screen.queryByRole('button', { name: 'Edit' })).not.toBeInTheDocument()
    expect(screen.getByText('Edit transaction')).toBeInTheDocument()
  })

  it('replaces the read-only fields with inputs and the Save/cancel buttons in edit mode', () => {
    render(<View {...baseProps} isEditing />)

    // Every editable field is a form control…
    expect(screen.getByLabelText('Transaction name')).toHaveValue('Weekly shop')
    expect(screen.getByLabelText('Transaction description')).toHaveValue('Milk, bread and coffee')
    expect(screen.getByLabelText('Amount')).toHaveValue('42,50')
    expect(screen.getByRole('button', { name: /groceries/i })).toBeInTheDocument()
    expect(screen.getByLabelText('Date')).toHaveValue('2026-10-10')
    // …and the definition list is gone.
    expect(screen.queryByText('10 October 2026 · 09:30')).not.toBeInTheDocument()
    expect(screen.queryByRole('heading', { name: 'Items' })).not.toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Remove transaction' })).not.toBeInTheDocument()

    expect(screen.getByRole('button', { name: 'Save changes' })).toBeInTheDocument()
    // Untouched form: the secondary button cancels rather than resets.
    expect(screen.getByRole('button', { name: 'Cancel' })).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Reset' })).not.toBeInTheDocument()
  })

  it('asks to cancel (leave edit mode) while nothing has changed', async () => {
    const user = userEvent.setup()
    const onCancel = vi.fn()
    render(<View {...baseProps} isEditing edit={{ ...editProps, isDirty: false, onCancel }} />)

    await user.click(screen.getByRole('button', { name: 'Cancel' }))

    expect(onCancel).toHaveBeenCalledTimes(1)
  })

  it('swaps Cancel for Reset once the form is dirty, and asks to reset', async () => {
    const user = userEvent.setup()
    const onReset = vi.fn()
    render(<View {...baseProps} isEditing edit={{ ...editProps, isDirty: true, onReset }} />)

    expect(screen.queryByRole('button', { name: 'Cancel' })).not.toBeInTheDocument()
    await user.click(screen.getByRole('button', { name: 'Reset' }))

    expect(onReset).toHaveBeenCalledTimes(1)
  })

  it('disables Save while the form is not saveable', () => {
    render(<View {...baseProps} isEditing edit={{ ...editProps, canSave: false }} />)

    expect(screen.getByRole('button', { name: 'Save changes' })).toBeDisabled()
  })

  it('opens a preview from the row for every previewable type', () => {
    render(<View {...baseProps} />)

    // The row itself is the preview trigger, for the image and the plain-text file alike.
    expect(screen.getByRole('button', { name: 'Preview attachment.jpg' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Preview note.txt' })).toBeInTheDocument()
  })

  it('opens the actions drawer from the row button', async () => {
    const user = userEvent.setup()
    const onOpenActions = vi.fn()
    render(<View {...baseProps} onOpenActions={onOpenActions} />)

    await user.click(screen.getByRole('button', { name: 'Actions for note.txt' }))

    expect(onOpenActions).toHaveBeenCalledWith(fileAttachment)
  })

  it('asks to preview the clicked attachment', async () => {
    const user = userEvent.setup()
    const onPreview = vi.fn()
    render(<View {...baseProps} onPreview={onPreview} />)

    await user.click(screen.getByRole('button', { name: 'Preview attachment.jpg' }))

    expect(onPreview).toHaveBeenCalledWith(imageAttachment)
  })

  it('indicates an attachment preview that is still loading, then reveals it', () => {
    render(<View {...baseProps} attachments={[imageAttachment]} />)

    // The metadata is present, but the image bytes are still in flight → spinner in the slot.
    expect(screen.getByRole('status', { name: 'Loading preview of attachment.jpg' })).toBeInTheDocument()
    expect(screen.getByAltText('Preview of attachment.jpg')).toHaveClass('invisible')

    fireEvent.load(screen.getByAltText('Preview of attachment.jpg'))

    expect(screen.queryByRole('status', { name: 'Loading preview of attachment.jpg' })).not.toBeInTheDocument()
    expect(screen.getByAltText('Preview of attachment.jpg')).not.toHaveClass('invisible')
  })

  it('falls back to the paperclip when an inline image cannot be rendered', () => {
    // Browsers cannot decode DNG; the thumbnail must degrade, not show a broken image.
    const dng: TransactionAttachmentItem = {
      id: 'attachment-3',
      filename: 'IMG_0001.dng',
      mediaType: 'image/dng',
      byteSize: 2_000_000,
      previewUrl: '/api/v1/attachment-content?grant=grant-dng',
    }
    render(<View {...baseProps} attachments={[dng]} />)

    fireEvent.error(screen.getByAltText('Preview of IMG_0001.dng'))

    expect(screen.queryByAltText('Preview of IMG_0001.dng')).not.toBeInTheDocument()
    // The row still previews (the lightbox will explain DNG can't be shown) and still has actions.
    expect(screen.getByRole('button', { name: 'Preview IMG_0001.dng' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Actions for IMG_0001.dng' })).toBeInTheDocument()
  })

  it('shows the loader while loading', () => {
    render(<View {...baseProps} isLoading />)

    expect(screen.getByRole('status', { name: 'Loading' })).toBeInTheDocument()
    expect(screen.queryByText('Weekly shop')).not.toBeInTheDocument()
  })

  it('shows a not-found message on error', () => {
    render(<View {...baseProps} isError />)

    expect(screen.getByText(/couldn't load this transaction/i)).toBeInTheDocument()
    expect(screen.queryByText('Weekly shop')).not.toBeInTheDocument()
  })
})
