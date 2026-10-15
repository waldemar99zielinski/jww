import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { describe, expect, it, vi } from 'vitest'
import { IconTag } from '@/components/icons'
import { SettingsRow } from './SettingsRow'

describe('SettingsRow', () => {
  it('renders a link with a chevron when `to` is given', () => {
    render(
      <MemoryRouter>
        <ul>
          <SettingsRow Icon={IconTag} label="Categories" to="/settings/categories" />
        </ul>
      </MemoryRouter>,
    )

    const link = screen.getByRole('link', { name: /categories/i })
    expect(link).toHaveAttribute('href', '/settings/categories')
  })

  it('renders trailing content instead of the chevron when provided', () => {
    render(
      <MemoryRouter>
        <ul>
          <SettingsRow Icon={IconTag} label="Currency" to="/settings/currencies" trailing={<small>EUR</small>} />
        </ul>
      </MemoryRouter>,
    )

    expect(screen.getByText('EUR')).toBeInTheDocument()
  })

  it('renders as plain non-interactive content when no `to` is given', () => {
    render(
      <MemoryRouter>
        <ul>
          <SettingsRow Icon={IconTag} label="Notifications" />
        </ul>
      </MemoryRouter>,
    )

    expect(screen.getByText('Notifications')).toBeInTheDocument()
    expect(screen.queryByRole('link')).not.toBeInTheDocument()
  })

  it('renders a real button when `onClick` is given and fires it on click', async () => {
    const onClick = vi.fn()
    render(
      <MemoryRouter>
        <ul>
          <SettingsRow Icon={IconTag} label="Install app" onClick={onClick} />
        </ul>
      </MemoryRouter>,
    )

    const button = screen.getByRole('button', { name: /install app/i })
    await userEvent.click(button)

    expect(onClick).toHaveBeenCalledTimes(1)
    expect(screen.queryByRole('link')).not.toBeInTheDocument()
  })
})
