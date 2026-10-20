import { screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { http, HttpResponse } from 'msw'
import { afterEach, describe, expect, it, vi } from 'vitest'
import LoginPage from '@/pages/login/LoginPage'
import * as webauthnLib from '@/lib/webauthn'
import { server } from '@/test/server'
import { renderWithProviders } from '@/test/renderWithProviders'

afterEach(() => vi.restoreAllMocks())

describe('LoginPage', () => {
  it('shows an error and stays unauthenticated for invalid credentials', async () => {
    server.use(
      http.post('*/api/v1/auth/login', () =>
        HttpResponse.json({ error: { message: 'unauthorized' } }, { status: 401 }),
      ),
    )

    const user = userEvent.setup()
    const { getByLabelText, findByText, store } = renderWithProviders(<LoginPage />, { route: '/login' })

    await user.type(getByLabelText('Username'), 'someone')
    await user.type(getByLabelText('Password'), 'wrong-password')
    await user.click(getByLabelText('Password'))
    await user.keyboard('{Enter}')

    expect(await findByText('Invalid username or password.')).toBeInTheDocument()
    expect(store.getState().auth.status).toBe('error')
  })

  it('authenticates and reaches the authenticated store state for valid credentials', async () => {
    const user = userEvent.setup()
    const { getByLabelText, getByRole, store } = renderWithProviders(<LoginPage />, { route: '/login' })

    await user.type(getByLabelText('Username'), 'waldemar')
    await user.type(getByLabelText('Password'), 'correct-password')
    await user.click(getByRole('button', { name: /sign in/i }))

    await waitFor(() => expect(store.getState().auth.status).toBe('authenticated'))
    expect(store.getState().auth.user?.username).toBe('waldemar')
  })

  it('hides the Face ID button when no platform authenticator is available (the default in tests/jsdom)', async () => {
    renderWithProviders(<LoginPage />, { route: '/login' })

    expect(await screen.findByLabelText('Username')).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: /face id/i })).not.toBeInTheDocument()
  })

  it('shows the Face ID button and signs in with it when a platform authenticator is available', async () => {
    vi.spyOn(webauthnLib, 'isPlatformAuthenticatorAvailable').mockResolvedValue(true)
    vi.spyOn(webauthnLib, 'authenticateDiscoverable').mockResolvedValue({ id: 'stub' })
    const user = userEvent.setup()
    const { store } = renderWithProviders(<LoginPage />, { route: '/login' })

    const passkeyButton = await screen.findByRole('button', { name: /sign in with face id/i })
    await user.click(passkeyButton)

    await waitFor(() => expect(store.getState().auth.status).toBe('authenticated'))
    expect(store.getState().auth.user?.username).toBe('waldemar')
  })
})
