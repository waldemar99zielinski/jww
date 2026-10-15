import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import { IconButton } from '@/components/ui/IconButton'

describe('IconButton', () => {
  it('renders as an accessible button via its aria-label', async () => {
    const onClick = vi.fn()
    render(<IconButton aria-label="Close" icon={<svg data-icon />} onClick={onClick} />)

    const button = screen.getByRole('button', { name: 'Close' })
    await userEvent.click(button)

    expect(onClick).toHaveBeenCalledTimes(1)
  })

  it('renders the given icon content inside the button', () => {
    render(<IconButton aria-label="Close" icon={<svg data-testid="close-icon" />} />)
    expect(screen.getByRole('button', { name: 'Close' }).querySelector('[data-testid="close-icon"]')).toBeInTheDocument()
  })
})
