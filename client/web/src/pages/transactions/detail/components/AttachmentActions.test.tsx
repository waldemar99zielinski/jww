import { fireEvent, render, screen } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import type { TransactionAttachmentItem } from '../types'
import { AttachmentActions } from './AttachmentActions'

const attachment: TransactionAttachmentItem = {
  id: 'attachment-1',
  filename: 'attachment.jpg',
  mediaType: 'image/jpeg',
  byteSize: 20480,
  previewUrl: '/api/v1/attachment-content?grant=grant-1',
}

function renderActions(overrides: Partial<Parameters<typeof AttachmentActions>[0]> = {}) {
  const props = {
    open: true,
    attachment,
    canPreview: true,
    isDownloading: false,
    isDeleting: false,
    deleteError: false,
    onPreview: vi.fn(),
    onDownload: vi.fn(),
    onDelete: vi.fn(),
    onClose: vi.fn(),
    ...overrides,
  }
  render(<AttachmentActions {...props} />)
  return props
}

describe('AttachmentActions', () => {
  it('offers preview, download and delete', () => {
    renderActions()

    expect(screen.getByRole('button', { name: 'Preview' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Download' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Delete' })).toBeInTheDocument()
  })

  it('hides preview for a type the browser cannot show', () => {
    renderActions({ canPreview: false })

    expect(screen.queryByRole('button', { name: 'Preview' })).not.toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Download' })).toBeInTheDocument()
  })

  it('fires the preview, download and delete callbacks', () => {
    const props = renderActions()

    fireEvent.click(screen.getByRole('button', { name: 'Preview' }))
    expect(props.onPreview).toHaveBeenCalledTimes(1)

    fireEvent.click(screen.getByRole('button', { name: 'Download' }))
    expect(props.onDownload).toHaveBeenCalledTimes(1)

    fireEvent.click(screen.getByRole('button', { name: 'Delete' }))
    fireEvent.click(screen.getByRole('button', { name: 'Delete attachment' }))
    expect(props.onDelete).toHaveBeenCalledTimes(1)
  })

  it('cancels a pending delete', () => {
    renderActions()

    fireEvent.click(screen.getByRole('button', { name: 'Delete' }))
    fireEvent.click(screen.getByRole('button', { name: 'Cancel' }))

    expect(screen.getByRole('button', { name: 'Delete' })).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Delete attachment' })).not.toBeInTheDocument()
  })

  it('shows a deletion error inside the confirm step', () => {
    renderActions({ deleteError: true })

    fireEvent.click(screen.getByRole('button', { name: 'Delete' }))

    expect(screen.getByText(/couldn't delete/i)).toBeInTheDocument()
  })

  it('disables the confirm button while deleting', () => {
    renderActions({ isDeleting: true })

    fireEvent.click(screen.getByRole('button', { name: 'Delete' }))

    expect(screen.getByRole('button', { name: /deleting/i })).toBeDisabled()
  })
})
