import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import { QuickFieldsRow } from './QuickFieldsRow'

describe('QuickFieldsRow', () => {
  it('shows the date label and "Attach" when there is no attachment, and fires the right callbacks', async () => {
    const user = userEvent.setup()
    const onToggleDate = vi.fn()
    const onOpenAttachment = vi.fn()
    render(
      <QuickFieldsRow dateLabel="Today" dateOpen={false} onToggleDate={onToggleDate} hasAttachment={false} onOpenAttachment={onOpenAttachment} />,
    )

    expect(screen.getByText('Today')).toBeInTheDocument()
    expect(screen.getByText('Attach')).toBeInTheDocument()

    await user.click(screen.getByText('Today'))
    expect(onToggleDate).toHaveBeenCalledTimes(1)

    await user.click(screen.getByText('Attach'))
    expect(onOpenAttachment).toHaveBeenCalledTimes(1)
  })

  it('shows "Attached" once an attachment exists', () => {
    render(<QuickFieldsRow dateLabel="Today" dateOpen={false} onToggleDate={vi.fn()} hasAttachment onOpenAttachment={vi.fn()} />)

    expect(screen.getByText('Attached')).toBeInTheDocument()
  })
})
