import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import { Switch } from '@/components/ui/Switch'

describe('Switch', () => {
  it('exposes role="switch" with aria-checked matching the checked prop', () => {
    render(<Switch checked={true} onChange={() => {}} label="Notifications" />)
    expect(screen.getByRole('switch', { name: 'Notifications' })).toHaveAttribute('aria-checked', 'true')
  })

  it('calls onChange with the toggled value on click', async () => {
    const onChange = vi.fn()
    render(<Switch checked={false} onChange={onChange} label="Notifications" />)

    await userEvent.click(screen.getByRole('switch', { name: 'Notifications' }))

    expect(onChange).toHaveBeenCalledWith(true)
  })
})
