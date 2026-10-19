import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import { DateQuickField } from './DateQuickField'

describe('DateQuickField', () => {
  it('shows the date label and toggles the date/time fields', async () => {
    const user = userEvent.setup()
    const onToggleDate = vi.fn()
    render(<DateQuickField dateLabel="Today" dateOpen={false} onToggleDate={onToggleDate} />)

    expect(screen.getByText('Today')).toBeInTheDocument()
    await user.click(screen.getByRole('button', { name: 'Today' }))

    expect(onToggleDate).toHaveBeenCalledTimes(1)
  })
})
