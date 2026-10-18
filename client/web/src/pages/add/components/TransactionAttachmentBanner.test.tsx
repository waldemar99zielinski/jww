import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import type { TransactionAttachmentDraft } from '../types'
import { TransactionAttachmentBanner } from './TransactionAttachmentBanner'

describe('TransactionAttachmentBanner', () => {
  it('shows an uploading state while the file is being sent', () => {
    const attachment: TransactionAttachmentDraft = { id: 'row-1', name: 'receipt.jpg', status: 'uploading', flow: 'attach' }
    render(<TransactionAttachmentBanner attachment={attachment} onRemove={vi.fn()} />)

    expect(screen.getByText('Uploading…')).toBeInTheDocument()
  })

  it('shows a spinner instead of the remove button while uploading', () => {
    const attachment: TransactionAttachmentDraft = { id: 'row-1', name: 'receipt.jpg', status: 'uploading', flow: 'attach' }
    render(<TransactionAttachmentBanner attachment={attachment} onRemove={vi.fn()} />)

    expect(screen.getByRole('status', { name: 'Uploading attachment' })).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: /remove attachment/i })).not.toBeInTheDocument()
  })

  it('shows the file name once ready and calls onRemove', async () => {
    const user = userEvent.setup()
    const onRemove = vi.fn()
    const attachment: TransactionAttachmentDraft = {
      id: 'row-1',
      name: 'receipt.jpg',
      status: 'ready',
      flow: 'attach',
      attachmentId: 'attachment-1',
    }
    render(<TransactionAttachmentBanner attachment={attachment} onRemove={onRemove} />)

    expect(screen.getByText(/receipt\.jpg/)).toBeInTheDocument()

    await user.click(screen.getByRole('button', { name: /remove attachment/i }))
    expect(onRemove).toHaveBeenCalledTimes(1)
  })

  it('shows an error state when the upload failed', () => {
    const attachment: TransactionAttachmentDraft = { id: 'row-1', name: 'receipt.jpg', status: 'error', flow: 'attach' }
    render(<TransactionAttachmentBanner attachment={attachment} onRemove={vi.fn()} />)

    expect(screen.getByText('Upload failed')).toBeInTheDocument()
    expect(screen.getByText(/couldn't attach this file/i)).toBeInTheDocument()
  })

  it('summarises a prefilled read and keeps the attachment', () => {
    const attachment: TransactionAttachmentDraft = {
      id: 'row-1',
      name: 'receipt.jpg',
      status: 'ready',
      flow: 'prefill',
      prefilledCount: 3,
    }
    render(<TransactionAttachmentBanner attachment={attachment} onRemove={vi.fn()} />)

    expect(screen.getByText(/3 fields filled from receipt\.jpg/i)).toBeInTheDocument()
    expect(screen.getByText(/check the form/i)).toBeInTheDocument()
  })

  it('names a failed read while keeping the attachment', () => {
    const attachment: TransactionAttachmentDraft = {
      id: 'row-1',
      name: 'receipt.jpg',
      status: 'ready',
      flow: 'prefill',
      readFailed: true,
    }
    render(<TransactionAttachmentBanner attachment={attachment} onRemove={vi.fn()} />)

    expect(screen.getByText(/couldn't read the file/i)).toBeInTheDocument()
    expect(screen.getByText(/attached when you save/i)).toBeInTheDocument()
  })

  it('disambiguates the remove button by file name so multiple rows are addressable', () => {
    const attachment: TransactionAttachmentDraft = {
      id: 'row-1',
      name: 'lidl.pdf',
      status: 'ready',
      flow: 'attach',
      attachmentId: 'attachment-1',
    }
    render(<TransactionAttachmentBanner attachment={attachment} onRemove={vi.fn()} />)

    expect(screen.getByRole('button', { name: 'Remove attachment lidl.pdf' })).toBeInTheDocument()
  })
})
