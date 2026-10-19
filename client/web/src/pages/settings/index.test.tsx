import { screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it } from 'vitest'
import { APP_VERSION } from '@/lib/version'
import { fixtureUser } from '@/test/fixtures'
import { createTestStore, renderWithProviders } from '@/test/renderWithProviders'
import SettingsPage from './index'

const IOS_SAFARI_UA =
  'Mozilla/5.0 (iPhone; CPU iPhone OS 17_5 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.5 Mobile/15E148 Safari/604.1'
const ANDROID_CHROME_UA =
  'Mozilla/5.0 (Linux; Android 14; Pixel 8) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Mobile Safari/537.36'

const originalUserAgent = window.navigator.userAgent

function setUserAgent(userAgent: string) {
  Object.defineProperty(window.navigator, 'userAgent', { configurable: true, get: () => userAgent })
}

function setDesktopViewport() {
  Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 1280 })
}

function authedStore() {
  return createTestStore({ status: 'authenticated', user: fixtureUser, error: null })
}

afterEach(() => setUserAgent(originalUserAgent))

describe('SettingsPage', () => {
  it('renders the signed-in user, their base currency, and links to the sub-pages', async () => {
    setDesktopViewport()
    renderWithProviders(<SettingsPage />, { store: authedStore() })

    expect(await screen.findByText(fixtureUser.display_name)).toBeInTheDocument()
    expect(screen.getByText(`@${fixtureUser.username}`)).toBeInTheDocument()
    expect(screen.getByText('EUR')).toBeInTheDocument()
    expect(screen.getByText(`v${APP_VERSION}`)).toBeInTheDocument()
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

  it('renders a passive, non-actionable install row on an unsupported browser', async () => {
    setDesktopViewport()
    setUserAgent('SomeCustomBrowser/1.0')
    renderWithProviders(<SettingsPage />, { store: authedStore() })

    expect(await screen.findByText('Install app')).toBeInTheDocument()
    expect(screen.getByText('Not available in this browser')).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: /install app/i })).not.toBeInTheDocument()
  })

  it('hides the install row on Chromium before a prompt is captured', async () => {
    setDesktopViewport()
    setUserAgent(ANDROID_CHROME_UA)
    renderWithProviders(<SettingsPage />, { store: authedStore() })

    expect(await screen.findByText(fixtureUser.display_name)).toBeInTheDocument()
    expect(screen.queryByText('Install app')).not.toBeInTheDocument()
  })

  it('opens the install instructions from the actionable row on iOS Safari', async () => {
    setDesktopViewport()
    setUserAgent(IOS_SAFARI_UA)
    const user = userEvent.setup()
    renderWithProviders(<SettingsPage />, { store: authedStore() })

    await user.click(await screen.findByRole('button', { name: /install app/i }))

    expect(screen.getByText('Install Joint Wallet Watch')).toBeInTheDocument()
    expect(screen.getByText(/add to home screen/i)).toBeInTheDocument()
  })
})
