import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import { InstallInstructions } from '@/pages/settings/components/InstallInstructions'

describe('InstallInstructions', () => {
  it('renders the iOS steps inside a dismissible sheet when compact', async () => {
    const onClose = vi.fn()
    const { container } = render(<InstallInstructions platform="ios-safari" compact open onClose={onClose} />)

    expect(container.querySelector('.sheet-panel')).not.toBeNull()
    expect(screen.getByText('Install Joint Wallet Watch')).toBeInTheDocument()
    expect(screen.getByText(/add to home screen/i)).toBeInTheDocument()

    await userEvent.click(screen.getByRole('button', { name: 'Close' }))
    expect(onClose).toHaveBeenCalledTimes(1)
  })

  it('renders the same iOS steps inline (no sheet) when non-compact', () => {
    const { container } = render(<InstallInstructions platform="ios-safari" compact={false} open onClose={() => {}} />)

    expect(container.querySelector('.sheet-panel')).toBeNull()
    expect(screen.getByText('Install Joint Wallet Watch')).toBeInTheDocument()
    expect(screen.getByText(/add to home screen/i)).toBeInTheDocument()
  })

  it('renders the macOS steps for the macOS platform', () => {
    render(<InstallInstructions platform="macos-safari" compact={false} open onClose={() => {}} />)

    expect(screen.getByText(/add to dock/i)).toBeInTheDocument()
    expect(screen.queryByText(/add to home screen/i)).not.toBeInTheDocument()
  })

  it('renders nothing inline when closed', () => {
    const { container } = render(
      <InstallInstructions platform="macos-safari" compact={false} open={false} onClose={() => {}} />,
    )

    expect(container).toBeEmptyDOMElement()
  })
})
