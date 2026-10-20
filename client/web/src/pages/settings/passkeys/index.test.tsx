import { screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { http, HttpResponse } from 'msw'
import { Route, Routes } from 'react-router-dom'
import { afterEach, describe, expect, it, vi } from 'vitest'
import * as webauthnLib from '@/lib/webauthn'
import { fixtureUser, fixtureWebAuthnCredential } from '@/test/fixtures'
import { createTestStore, renderWithProviders } from '@/test/renderWithProviders'
import { server } from '@/test/server'
import SettingsPasskeysPage from './index'

function setDesktopViewport() {
  Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 1280 })
}

function authedStore() {
  return createTestStore({ status: 'authenticated', user: fixtureUser, error: null })
}

afterEach(() => {
  vi.restoreAllMocks()
  vi.unstubAllEnvs()
})

describe('SettingsPasskeysPage', () => {
  it('redirects to /settings in design mode without calling the backend', async () => {
    setDesktopViewport()
    vi.stubEnv('VITE_DESIGN_MODE', 'true')
    let credentialsRequested = false
    server.use(
      http.get('*/api/v1/webauthn/credentials', () => {
        credentialsRequested = true
        return HttpResponse.json({ data: [] })
      }),
    )
    renderWithProviders(
      <Routes>
        <Route path="/settings/passkeys" element={<SettingsPasskeysPage />} />
        <Route path="/settings" element={<div>Settings home</div>} />
      </Routes>,
      { store: authedStore(), route: '/settings/passkeys' },
    )

    expect(await screen.findByText('Settings home')).toBeInTheDocument()
    expect(credentialsRequested).toBe(false)
  })

  it('lists enrolled credentials', async () => {
    setDesktopViewport()
    renderWithProviders(<SettingsPasskeysPage />, { store: authedStore() })

    expect(await screen.findByText('iPhone')).toBeInTheDocument()
  })

  it('shows an empty state when no credentials are enrolled', async () => {
    setDesktopViewport()
    server.use(http.get('*/api/v1/webauthn/credentials', () => HttpResponse.json({ data: [] })))
    renderWithProviders(<SettingsPasskeysPage />, { store: authedStore() })

    expect(await screen.findByText(/no devices enrolled yet/i)).toBeInTheDocument()
  })

  it('enrolls a new device: begin -> browser ceremony -> finish, then clears the name field and relists', async () => {
    setDesktopViewport()
    vi.spyOn(webauthnLib, 'registerCredential').mockResolvedValue({ id: 'stub' })
    let finishBody: unknown
    const credentials = [fixtureWebAuthnCredential]
    server.use(
      http.get('*/api/v1/webauthn/credentials', () => HttpResponse.json({ data: credentials })),
      http.post('*/api/v1/webauthn/credentials/register/finish', async ({ request }) => {
        finishBody = await request.json()
        const created = { ...fixtureWebAuthnCredential, id: 'new-id', name: (finishBody as { name: string }).name }
        credentials.push(created)
        return HttpResponse.json({ data: created }, { status: 201 })
      }),
    )
    const user = userEvent.setup()
    renderWithProviders(<SettingsPasskeysPage />, { store: authedStore() })

    await screen.findByText('iPhone')
    const input = screen.getByPlaceholderText('e.g. My iPhone')
    await user.type(input, 'My iPad')
    await user.click(screen.getByRole('button', { name: 'Add' }))

    await waitFor(() => expect(finishBody).toMatchObject({ challenge_token: 'fixture-challenge-token', name: 'My iPad' }))
    await waitFor(() => expect(input).toHaveValue(''))
    expect(await screen.findByText('My iPad')).toBeInTheDocument()
  })

  it('shows an error and keeps the name field when the ceremony or server rejects enrollment', async () => {
    setDesktopViewport()
    vi.spyOn(webauthnLib, 'registerCredential').mockRejectedValue(new Error('user cancelled'))
    const user = userEvent.setup()
    renderWithProviders(<SettingsPasskeysPage />, { store: authedStore() })

    await screen.findByText('iPhone')
    const input = screen.getByPlaceholderText('e.g. My iPhone')
    await user.type(input, 'My iPad')
    await user.click(screen.getByRole('button', { name: 'Add' }))

    expect(await screen.findByText(/didn't work/i)).toBeInTheDocument()
    expect(input).toHaveValue('My iPad')
  })

  it('opens a confirmation drawer and removes the credential after confirming', async () => {
    setDesktopViewport()
    let deleteCalled = false
    server.use(
      http.delete('*/api/v1/webauthn/credentials/:id', () => {
        deleteCalled = true
        return new HttpResponse(null, { status: 204 })
      }),
    )
    const user = userEvent.setup()
    renderWithProviders(<SettingsPasskeysPage />, { store: authedStore() })

    await screen.findByText('iPhone')
    await user.click(screen.getByRole('button', { name: /remove iphone/i }))

    expect(await screen.findByText('Remove this device?')).toBeInTheDocument()
    await user.click(screen.getByRole('button', { name: 'Remove permanently' }))

    await waitFor(() => expect(deleteCalled).toBe(true))
  })

  it('does not delete when the confirmation drawer is cancelled', async () => {
    setDesktopViewport()
    let deleteCalled = false
    server.use(
      http.delete('*/api/v1/webauthn/credentials/:id', () => {
        deleteCalled = true
        return new HttpResponse(null, { status: 204 })
      }),
    )
    const user = userEvent.setup()
    renderWithProviders(<SettingsPasskeysPage />, { store: authedStore() })

    await screen.findByText('iPhone')
    await user.click(screen.getByRole('button', { name: /remove iphone/i }))
    expect(await screen.findByText('Remove this device?')).toBeInTheDocument()

    await user.click(screen.getByRole('button', { name: 'Cancel' }))

    expect(deleteCalled).toBe(false)
  })

  it('keeps the confirmation open and shows an error when removal fails', async () => {
    setDesktopViewport()
    server.use(http.delete('*/api/v1/webauthn/credentials/:id', () => HttpResponse.json({ error: { message: 'boom' } }, { status: 500 })))
    const user = userEvent.setup()
    renderWithProviders(<SettingsPasskeysPage />, { store: authedStore() })

    await screen.findByText('iPhone')
    await user.click(screen.getByRole('button', { name: /remove iphone/i }))
    await user.click(await screen.findByRole('button', { name: 'Remove permanently' }))

    expect(await screen.findByText(/couldn't remove the device/i)).toBeInTheDocument()
    expect(screen.getByText('Remove this device?')).toBeInTheDocument()
  })
})
