import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { http, HttpResponse } from 'msw'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { downloadTransactionAttachment } from '@/api/attachments'
import { server } from '@/test/server'
import { AttachmentPreview, type AttachmentPreviewItem } from './AttachmentPreview'

function attachment(overrides: Partial<AttachmentPreviewItem>): AttachmentPreviewItem {
  return {
    id: 'attachment-1',
    filename: 'attachment.jpg',
    mediaType: 'image/jpeg',
    byteSize: 20480,
    previewUrl: '/api/v1/attachment-content?grant=grant-1',
    ...overrides,
  }
}

function renderPreview(item: AttachmentPreviewItem, props: Partial<Parameters<typeof AttachmentPreview>[0]> = {}) {
  return render(
    <AttachmentPreview
      attachment={item}
      isDownloading={false}
      onDownload={vi.fn()}
      loadBlob={(attachmentId) => downloadTransactionAttachment(1, attachmentId)}
      onClose={vi.fn()}
      {...props}
    />,
  )
}

/** Drag on the (bubbling) dialog so the wrapper's touch handlers see it. */
function swipe(from: { x: number; y: number }, to: { x: number; y: number }) {
  const dialog = screen.getByRole('dialog')
  fireEvent.touchStart(dialog, { touches: [{ clientX: from.x, clientY: from.y }] })
  fireEvent.touchMove(dialog, { touches: [{ clientX: to.x, clientY: to.y }] })
  fireEvent.touchEnd(dialog, { changedTouches: [{ clientX: to.x, clientY: to.y }] })
}

describe('AttachmentPreview', () => {
  afterEach(() => vi.useRealTimers())
  beforeEach(() => {
    // jsdom doesn't implement object URLs; the blob-backed renderers create them.
    window.URL.createObjectURL = vi.fn(() => 'blob:fixture')
    window.URL.revokeObjectURL = vi.fn()
  })

  it('renders an image from its inline URL without fetching the blob', () => {
    renderPreview(attachment({}))

    expect(screen.getByAltText('Preview of attachment.jpg')).toHaveAttribute(
      'src',
      '/api/v1/attachment-content?grant=grant-1',
    )
  })

  it('keeps a loader up until the previewed media has actually loaded', () => {
    renderPreview(attachment({}))

    // The source URL is set immediately, but the image bytes are still in flight.
    expect(screen.getByRole('status', { name: 'Loading preview' })).toBeInTheDocument()

    fireEvent.load(screen.getByAltText('Preview of attachment.jpg'))

    expect(screen.queryByRole('status', { name: 'Loading preview' })).not.toBeInTheDocument()
  })

  it('embeds a PDF in a viewer frame', () => {
    renderPreview(attachment({ filename: 'statement.pdf', mediaType: 'application/pdf' }))

    expect(screen.getByTitle('Preview of statement.pdf').tagName).toBe('IFRAME')
  })

  it('fetches and renders plain text through the provided blob loader', async () => {
    server.use(
      http.get('*/api/v1/transactions/:id/attachments/:attachmentID/download', () =>
        new HttpResponse('amount,label\n12.50,coffee', { headers: { 'Content-Type': 'text/csv' } }),
      ),
    )
    renderPreview(attachment({ filename: 'invoice.csv', mediaType: 'text/csv', previewUrl: null }))

    expect(await screen.findByText(/amount,label/)).toBeInTheDocument()
  })

  it('renders video and audio with native controls', async () => {
    const { unmount } = renderPreview(attachment({ filename: 'clip.mp4', mediaType: 'video/mp4', previewUrl: null }))
    await waitFor(() => expect(document.querySelector('video')).toBeTruthy())
    expect(document.querySelector('video')).toHaveAttribute('controls')
    unmount()

    renderPreview(attachment({ filename: 'voice.mp3', mediaType: 'audio/mpeg', previewUrl: null }))
    await waitFor(() => expect(document.querySelector('audio')).toBeTruthy())
    expect(document.querySelector('audio')).toHaveAttribute('controls')
  })

  it('explains when a file type cannot be previewed', async () => {
    renderPreview(
      attachment({
        filename: 'report.docx',
        mediaType: 'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
        previewUrl: null,
      }),
    )

    expect(await screen.findByText(/can't be previewed/i)).toBeInTheDocument()
  })

  it('closes on Escape and on the close button', () => {
    const onClose = vi.fn()
    renderPreview(attachment({}), { onClose })

    fireEvent.keyDown(document, { key: 'Escape' })
    expect(onClose).toHaveBeenCalledTimes(1)

    fireEvent.click(screen.getByRole('button', { name: 'Close preview' }))
    expect(onClose).toHaveBeenCalledTimes(2)
  })

  it('delegates downloads to the provided handler', () => {
    const onDownload = vi.fn()
    const item = attachment({})
    renderPreview(item, { onDownload })

    fireEvent.click(screen.getByRole('button', { name: 'Download attachment.jpg' }))
    expect(onDownload).toHaveBeenCalledWith(item)
  })

  it('closes after a swipe down past the threshold', () => {
    vi.useFakeTimers()
    const onClose = vi.fn()
    renderPreview(attachment({}), { onClose })

    swipe({ x: 200, y: 100 }, { x: 202, y: 420 })

    // It animates out first, then unmounts.
    expect(onClose).not.toHaveBeenCalled()
    vi.advanceTimersByTime(300)
    expect(onClose).toHaveBeenCalledTimes(1)
  })

  it('springs back without closing on a short swipe down', () => {
    const onClose = vi.fn()
    renderPreview(attachment({}), { onClose })

    swipe({ x: 200, y: 100 }, { x: 200, y: 130 })

    expect(onClose).not.toHaveBeenCalled()
  })

  it('ignores a mostly-horizontal drag', () => {
    const onClose = vi.fn()
    renderPreview(attachment({}), { onClose })

    swipe({ x: 100, y: 100 }, { x: 380, y: 120 })

    expect(onClose).not.toHaveBeenCalled()
  })
})
