import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import type { ReceiptDraft } from '../types'
import { ReceiptBanner } from './ReceiptBanner'

describe('ReceiptBanner', () => {
  it('shows a "parsing" state while the receipt is being read', () => {
    const receipt: ReceiptDraft = { name: 'receipt.jpg', status: 'parsing' }
    render(<ReceiptBanner receipt={receipt} onRemove={vi.fn()} />)

    expect(screen.getByText('Reading file…')).toBeInTheDocument()
  })

  it('shows the file name once ready and calls onRemove', async () => {
    const user = userEvent.setup()
    const onRemove = vi.fn()
    const receipt: ReceiptDraft = { name: 'receipt.jpg', status: 'ready' }
    render(<ReceiptBanner receipt={receipt} onRemove={onRemove} />)

    expect(screen.getByText(/receipt\.jpg/)).toBeInTheDocument()

    await user.click(screen.getByRole('button', { name: 'Remove attachment' }))
    expect(onRemove).toHaveBeenCalledTimes(1)
  })
})
