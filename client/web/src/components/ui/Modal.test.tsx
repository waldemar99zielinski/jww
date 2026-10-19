import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import { Modal } from './Modal'

describe('Modal', () => {
  it('renders nothing while closed', () => {
    const { container } = render(
      <Modal open={false} onClose={vi.fn()} title="Edit item">
        <p>Content</p>
      </Modal>,
    )

    expect(container).toBeEmptyDOMElement()
  })

  it('renders an accessible dialog with title, content and footer', () => {
    render(
      <Modal open onClose={vi.fn()} title="Edit item" description="Change the fields" footer={<button>Save</button>}>
        <p>Content</p>
      </Modal>,
    )

    expect(screen.getByRole('dialog', { name: 'Edit item' })).toBeInTheDocument()
    expect(screen.getByText('Change the fields')).toBeInTheDocument()
    expect(screen.getByText('Content')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Save' })).toBeInTheDocument()
  })

  it('closes on Escape and on the close button', async () => {
    const user = userEvent.setup()
    const onClose = vi.fn()
    render(
      <Modal open onClose={onClose} title="Edit item">
        <p>Content</p>
      </Modal>,
    )

    await user.keyboard('{Escape}')
    expect(onClose).toHaveBeenCalledTimes(1)

    await user.click(screen.getByRole('button', { name: 'Close' }))
    expect(onClose).toHaveBeenCalledTimes(2)
  })
})
