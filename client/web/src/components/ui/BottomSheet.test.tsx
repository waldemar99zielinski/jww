import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import { BottomSheet } from '@/components/ui/BottomSheet'

describe('BottomSheet', () => {
  it('marks itself aria-hidden and inert when closed', () => {
    const { container } = render(
      <BottomSheet open={false} onClose={() => {}} title="Filters">
        <p>Content</p>
      </BottomSheet>,
    )

    const root = container.firstElementChild as HTMLElement
    expect(root).toHaveAttribute('aria-hidden', 'true')
    expect(root).toHaveAttribute('inert')
  })

  it('is not aria-hidden/inert when open', () => {
    const { container } = render(
      <BottomSheet open={true} onClose={() => {}} title="Filters">
        <p>Content</p>
      </BottomSheet>,
    )

    const root = container.firstElementChild as HTMLElement
    expect(root).toHaveAttribute('aria-hidden', 'false')
    expect(root).not.toHaveAttribute('inert')
  })

  it('calls onClose when the close button is clicked', async () => {
    const onClose = vi.fn()
    render(
      <BottomSheet open={true} onClose={onClose} title="Filters">
        <p>Content</p>
      </BottomSheet>,
    )

    await userEvent.click(screen.getByRole('button', { name: 'Close' }))

    expect(onClose).toHaveBeenCalledTimes(1)
  })

  it('calls onClose when Escape is pressed while open', async () => {
    const onClose = vi.fn()
    render(
      <BottomSheet open={true} onClose={onClose} title="Filters">
        <p>Content</p>
      </BottomSheet>,
    )

    await userEvent.keyboard('{Escape}')

    expect(onClose).toHaveBeenCalledTimes(1)
  })

  it('renders the title, description, and footer', () => {
    render(
      <BottomSheet
        open={true}
        onClose={() => {}}
        title="Filters"
        description="Narrow the list"
        footer={<button>Apply</button>}
      >
        <p>Content</p>
      </BottomSheet>,
    )

    expect(screen.getByText('Filters')).toBeInTheDocument()
    expect(screen.getByText('Narrow the list')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Apply' })).toBeInTheDocument()
  })
})
