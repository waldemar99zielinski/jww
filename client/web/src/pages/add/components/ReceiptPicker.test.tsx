import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import { ReceiptPicker } from './ReceiptPicker'

describe('ReceiptPicker', () => {
  it('closes and forwards the picked file when choosing a file', async () => {
    const user = userEvent.setup()
    const onClose = vi.fn()
    const onFile = vi.fn()
    const { container } = render(<ReceiptPicker open onClose={onClose} onFile={onFile} />)

    await user.click(screen.getByRole('button', { name: 'Choose a file' }))
    expect(onClose).toHaveBeenCalledTimes(1)

    const fileInput = container.querySelectorAll('input[type="file"]')[1] as HTMLInputElement
    const file = new File(['content'], 'receipt.jpg', { type: 'image/jpeg' })
    await user.upload(fileInput, file)

    expect(onFile).toHaveBeenCalledWith(file)
  })

  it('stays mounted but aria-hidden when closed', () => {
    render(<ReceiptPicker open={false} onClose={vi.fn()} onFile={vi.fn()} />)

    const title = screen.getByText('Add attachment')
    expect(title.closest('[aria-hidden="true"]')).toBeTruthy()
  })
})
