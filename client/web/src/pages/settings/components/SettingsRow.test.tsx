import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { describe, expect, it } from 'vitest'
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
})
