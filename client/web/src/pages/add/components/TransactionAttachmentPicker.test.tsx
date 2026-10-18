import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import { TransactionAttachmentPicker } from './TransactionAttachmentPicker'

describe('TransactionAttachmentPicker', () => {
  it('forwards the picked file with the "prefill" flow chosen in the first step', async () => {
    const user = userEvent.setup()
    const onClose = vi.fn()
    const onFile = vi.fn()
    const { container } = render(<TransactionAttachmentPicker open onClose={onClose} onFile={onFile} />)

    await user.click(screen.getByRole('button', { name: /import & prefill/i }))
    await user.click(screen.getByRole('button', { name: 'Choose files' }))
    expect(onClose).toHaveBeenCalledTimes(1)

    const fileInput = container.querySelectorAll('input[type="file"]')[1] as HTMLInputElement
    const file = new File(['content'], 'receipt.jpg', { type: 'image/jpeg' })
    await user.upload(fileInput, file)

    expect(onFile).toHaveBeenCalledWith(file, 'prefill')
  })

  it('forwards the "attach" flow when "Just attach" is chosen', async () => {
    const user = userEvent.setup()
    const onFile = vi.fn()
    const { container } = render(<TransactionAttachmentPicker open onClose={vi.fn()} onFile={onFile} />)

    await user.click(screen.getByRole('button', { name: /just attach/i }))
    await user.click(screen.getByRole('button', { name: 'Choose files' }))

    const fileInput = container.querySelectorAll('input[type="file"]')[1] as HTMLInputElement
    await user.upload(fileInput, new File(['content'], 'receipt.pdf', { type: 'application/pdf' }))

    expect(onFile).toHaveBeenCalledWith(expect.any(File), 'attach')
  })

  it('forwards every file when several are selected at once', async () => {
    const user = userEvent.setup()
    const onFile = vi.fn()
    const { container } = render(<TransactionAttachmentPicker open onClose={vi.fn()} onFile={onFile} />)

    await user.click(screen.getByRole('button', { name: /just attach/i }))
    await user.click(screen.getByRole('button', { name: 'Choose files' }))

    const fileInput = container.querySelectorAll('input[type="file"]')[1] as HTMLInputElement
    const files = [
      new File(['a'], 'one.jpg', { type: 'image/jpeg' }),
      new File(['b'], 'two.pdf', { type: 'application/pdf' }),
    ]
    await user.upload(fileInput, files)

    expect(onFile).toHaveBeenCalledTimes(2)
    expect(onFile).toHaveBeenNthCalledWith(1, files[0], 'attach')
    expect(onFile).toHaveBeenNthCalledWith(2, files[1], 'attach')
  })

  it('returns to the mode choice from the source step', async () => {
    const user = userEvent.setup()
    render(<TransactionAttachmentPicker open onClose={vi.fn()} onFile={vi.fn()} />)

    await user.click(screen.getByRole('button', { name: /just attach/i }))
    expect(screen.queryByRole('button', { name: /just attach/i })).not.toBeInTheDocument()

    await user.click(screen.getByRole('button', { name: /back/i }))
    expect(screen.getByRole('button', { name: /import & prefill/i })).toBeInTheDocument()
  })

  it('stays mounted but aria-hidden when closed', () => {
    render(<TransactionAttachmentPicker open={false} onClose={vi.fn()} onFile={vi.fn()} />)

    const title = screen.getByText('Add attachment')
    expect(title.closest('[aria-hidden="true"]')).toBeTruthy()
  })
})
