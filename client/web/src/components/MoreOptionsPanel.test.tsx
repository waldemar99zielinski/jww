import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import type { UserDto } from '@/api/types'
import { MoreOptionsPanel } from './MoreOptionsPanel'

const users: UserDto[] = [
  { id: 'user-1', username: 'a', display_name: 'A', created_at: '' },
  { id: 'user-2', username: 'b', display_name: 'B', created_at: '' },
]

describe('MoreOptionsPanel', () => {
  it('is collapsed by default and expands on toggle', async () => {
    const user = userEvent.setup()
    const onToggle = vi.fn()
    render(
      <MoreOptionsPanel
        open={false}
        onToggle={onToggle}
        users={users}
        userId="user-1"
        onUserChange={vi.fn()}
        showUserPicker
      />,
    )

    expect(screen.queryByLabelText(/paid by/i)).not.toBeInTheDocument()
    await user.click(screen.getByRole('button', { name: /more options/i }))
    expect(onToggle).toHaveBeenCalledTimes(1)
  })

  it('shows the payer picker once open', () => {
    render(
      <MoreOptionsPanel open onToggle={vi.fn()} users={users} userId="user-1" onUserChange={vi.fn()} showUserPicker />,
    )

    expect(screen.getByLabelText(/paid by/i)).toBeInTheDocument()
  })

  it('renders nothing when showUserPicker is false', () => {
    render(
      <MoreOptionsPanel open onToggle={vi.fn()} users={users} userId="user-1" onUserChange={vi.fn()} showUserPicker={false} />,
    )

    expect(screen.queryByLabelText(/paid by/i)).not.toBeInTheDocument()
    expect(screen.queryByRole('button', { name: /more options/i })).not.toBeInTheDocument()
  })
})
