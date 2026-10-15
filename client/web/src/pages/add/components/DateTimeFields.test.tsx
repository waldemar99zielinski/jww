import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import { DateTimeFields } from './DateTimeFields'

describe('DateTimeFields', () => {
  it('renders the given date and time and calls onUseToday', async () => {
    const user = userEvent.setup()
    const onUseToday = vi.fn()
    render(
      <DateTimeFields date="2026-10-10" time="09:30" onDateChange={vi.fn()} onTimeChange={vi.fn()} onUseToday={onUseToday} />,
    )

    expect(screen.getByLabelText('Date')).toHaveValue('2026-10-10')
    expect(screen.getByLabelText('Time')).toHaveValue('09:30')

    await user.click(screen.getByRole('button', { name: 'Use today' }))
    expect(onUseToday).toHaveBeenCalledTimes(1)
  })

  it('calls onDateChange/onTimeChange when edited', async () => {
    const user = userEvent.setup()
    const onDateChange = vi.fn()
    render(<DateTimeFields date="2026-10-10" time="09:30" onDateChange={onDateChange} onTimeChange={vi.fn()} onUseToday={vi.fn()} />)

    await user.clear(screen.getByLabelText('Date'))
    await user.type(screen.getByLabelText('Date'), '2026-01-02')

    expect(onDateChange).toHaveBeenCalled()
  })
})
