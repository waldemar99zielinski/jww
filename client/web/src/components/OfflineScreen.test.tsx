import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import { OfflineScreen } from '@/components/OfflineScreen'

describe('OfflineScreen', () => {
  it('exposes an alert dialog with the offline accessible name and description', () => {
    render(<OfflineScreen onRetry={() => {}} stillOffline={false} />)

    const dialog = screen.getByRole('alertdialog', { name: /offline/i })
    expect(dialog).toHaveAccessibleDescription(/needs a connection/i)
  })

  it('invokes onRetry when Retry is clicked', async () => {
    const onRetry = vi.fn()
    render(<OfflineScreen onRetry={onRetry} stillOffline={false} />)

    await userEvent.click(screen.getByRole('button', { name: /retry/i }))

    expect(onRetry).toHaveBeenCalledTimes(1)
  })

  it('shows the failure line only when stillOffline', () => {
    const { rerender } = render(<OfflineScreen onRetry={() => {}} stillOffline={false} />)
    expect(screen.queryByText(/still offline/i)).not.toBeInTheDocument()

    rerender(<OfflineScreen onRetry={() => {}} stillOffline />)
    expect(screen.getByText(/still offline — check your connection/i)).toBeInTheDocument()
  })
})
