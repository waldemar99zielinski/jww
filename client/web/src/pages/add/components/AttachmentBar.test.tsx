import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import { AttachmentBar } from './AttachmentBar'
import type { TransactionAttachmentDraft } from '../types'

const attachment: TransactionAttachmentDraft = {
  id: 'row-1',
  name: 'photo.jpg',
  status: 'ready',
  attachmentId: 'a1',
  isImage: true,
}

function renderBar(overrides: Partial<Parameters<typeof AttachmentBar>[0]> = {}) {
  const props = {
    attachments: [] as TransactionAttachmentDraft[],
    status: 'idle' as const,
    canScan: false,
    canUndo: false,
    onRemove: vi.fn(),
    onPreview: vi.fn(),
    onScan: vi.fn(),
    onCancel: vi.fn(),
    onUndo: vi.fn(),
    ...overrides,
  }
  return { ...render(<AttachmentBar {...props} />), props }
}

describe('AttachmentBar', () => {
  it('renders nothing until an attachment exists', () => {
    const { container } = renderBar()

    expect(container).toBeEmptyDOMElement()
  })

  it('shows a compact ready state and starts a scan', async () => {
    const user = userEvent.setup()
    const { props } = renderBar({ attachments: [attachment], canScan: true })

    expect(screen.getByText('1 attachment')).toBeInTheDocument()
    expect(screen.getByText('Ready to scan.')).toBeInTheDocument()
    // The file list stays collapsed until details are requested.
    expect(screen.queryByText('photo.jpg')).not.toBeInTheDocument()

    await user.click(screen.getByRole('button', { name: 'Scan attachment' }))
    expect(props.onScan).toHaveBeenCalledTimes(1)
    // Clicking a button keeps its action and does not toggle the details.
    expect(screen.queryByText('photo.jpg')).not.toBeInTheDocument()

    // Anywhere else on the row expands the file list.
    await user.click(screen.getByText('1 attachment'))
    expect(screen.getByText('photo.jpg')).toBeInTheDocument()
    expect(screen.getByText('Ready to scan')).toBeInTheDocument()

    await user.click(screen.getByRole('button', { name: 'Preview photo.jpg' }))
    expect(props.onPreview).toHaveBeenCalledWith('row-1')

    await user.click(screen.getByRole('button', { name: 'Remove attachment photo.jpg' }))
    expect(props.onRemove).toHaveBeenCalledWith('row-1')
  })

  it('shows reading progress and lets the user cancel', async () => {
    const user = userEvent.setup()
    const { props } = renderBar({
      attachments: [{ ...attachment, status: 'processing' }],
      status: 'reading',
    })

    expect(screen.getByText('Reading attachment…')).toBeInTheDocument()
    await user.click(screen.getByRole('button', { name: 'Cancel' }))

    expect(props.onCancel).toHaveBeenCalledTimes(1)
  })

  it('summarizes a successful scan and keeps details behind the disclosure', async () => {
    const user = userEvent.setup()
    const onReviewItems = vi.fn()
    const { props } = renderBar({
      attachments: [attachment],
      status: 'success',
      canScan: true,
      canUndo: true,
      onReviewItems,
      summary: {
        attachmentIds: ['a1'],
        itemCount: 2,
        attachmentTotal: 1248,
        warnings: ['Discount line could not be read'],
        categoryHint: 'Groceries',
      },
    })

    expect(screen.getByText('2 items added')).toBeInTheDocument()
    expect(screen.getByText('Form filled — check the attachment details.')).toBeInTheDocument()
    expect(screen.queryByText('Discount line could not be read')).not.toBeInTheDocument()

    await user.click(screen.getByRole('button', { name: 'Review items' }))
    expect(onReviewItems).toHaveBeenCalledTimes(1)

    await user.click(screen.getByRole('button', { name: 'Show attachment details' }))
    expect(screen.getByText('Discount line could not be read')).toBeInTheDocument()
    expect(screen.getByText(/category .*groceries.* isn't in your list/i)).toBeInTheDocument()

    await user.click(screen.getByRole('button', { name: 'Undo scan' }))
    expect(props.onUndo).toHaveBeenCalledTimes(1)
  })

  it('offers a retry after a failed read', async () => {
    const user = userEvent.setup()
    const { props } = renderBar({
      attachments: [attachment],
      status: 'failed',
      canScan: true,
      canUndo: true,
    })

    expect(screen.getByText("Couldn't read attachment")).toBeInTheDocument()
    expect(screen.getByText('Files are still attached — try again.')).toBeInTheDocument()

    await user.click(screen.getByRole('button', { name: 'Try again' }))
    expect(props.onScan).toHaveBeenCalledTimes(1)

    await user.click(screen.getByRole('button', { name: 'Show attachment details' }))
    await user.click(screen.getByRole('button', { name: 'Undo scan' }))
    expect(props.onUndo).toHaveBeenCalledTimes(1)
  })
})
