import { screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it } from 'vitest'
import { fixtureUser } from '@/test/fixtures'
import { createTestStore, renderWithProviders } from '@/test/renderWithProviders'
import SettingsPage from './index'

function setDesktopViewport() {
  Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 1280 })
}

function authedStore() {
  return createTestStore({ status: 'authenticated', user: fixtureUser, error: null })
}

describe('SettingsPage', () => {
  it('renders the signed-in user, their base currency, and links to the sub-pages', async () => {
    setDesktopViewport()
    renderWithProviders(<SettingsPage />, { store: authedStore() })

    expect(await screen.findByText(fixtureUser.display_name)).toBeInTheDocument()
    expect(screen.getByText(`@${fixtureUser.username}`)).toBeInTheDocument()
    expect(screen.getByText('EUR')).toBeInTheDocument()
    expect(screen.getByRole('link', { name: /categories/i })).toHaveAttribute('href', '/settings/categories')
    expect(screen.getByRole('link', { name: /currency/i })).toHaveAttribute('href', '/settings/currencies')
  })

  it('logs the user out when "Log out" is clicked', async () => {
    setDesktopViewport()
    const user = userEvent.setup()
    const store = authedStore()
    renderWithProviders(<SettingsPage />, { store })

    await user.click(await screen.findByRole('button', { name: /log out/i }))

    expect(store.getState().auth.status).toBe('unauthenticated')
  })
})
