import { act, renderHook, waitFor } from '@testing-library/react'
import type { ReactNode } from 'react'
import { Provider as ReduxProvider } from 'react-redux'
import { HttpResponse, http } from 'msw'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { useAuth } from '@/hooks/useAuth'
import * as webauthnLib from '@/lib/webauthn'
import { createTestStore } from '@/test/renderWithProviders'
import { server } from '@/test/server'
import { fixtureUser } from '@/test/fixtures'

afterEach(() => vi.restoreAllMocks())

function wrapperFor(store: ReturnType<typeof createTestStore>) {
  return function Wrapper({ children }: { children: ReactNode }) {
    return <ReduxProvider store={store}>{children}</ReduxProvider>
  }
}

describe('useAuth', () => {
  it('exposes the preloaded unauthenticated state', () => {
    const store = createTestStore({ status: 'unauthenticated', user: null, error: null })
    const { result } = renderHook(() => useAuth(), { wrapper: wrapperFor(store) })

    expect(result.current.status).toBe('unauthenticated')
    expect(result.current.user).toBeNull()
    expect(result.current.error).toBeNull()
  })

  it('exposes the preloaded authenticated user', () => {
    const store = createTestStore({ status: 'authenticated', user: fixtureUser, error: null })
    const { result } = renderHook(() => useAuth(), { wrapper: wrapperFor(store) })

    expect(result.current.status).toBe('authenticated')
    expect(result.current.user).toEqual(fixtureUser)
  })

  it('login() resolves and updates status to authenticated on success', async () => {
    const store = createTestStore({ status: 'unauthenticated', user: null, error: null })
    const { result } = renderHook(() => useAuth(), { wrapper: wrapperFor(store) })

    await act(async () => {
      await result.current.login(fixtureUser.username, 'correct-password')
    })

    await waitFor(() => expect(result.current.status).toBe('authenticated'))
    expect(result.current.user).toEqual(fixtureUser)
  })

  it('login() rejects and sets a friendly error on invalid credentials (401)', async () => {
    server.use(
      http.post('*/api/v1/auth/login', () =>
        HttpResponse.json({ error: { message: 'bad credentials' } }, { status: 401 }),
      ),
    )
    const store = createTestStore({ status: 'unauthenticated', user: null, error: null })
    const { result } = renderHook(() => useAuth(), { wrapper: wrapperFor(store) })

    await act(async () => {
      await expect(result.current.login(fixtureUser.username, 'wrong-password')).rejects.toBeTruthy()
    })

    await waitFor(() => expect(result.current.status).toBe('error'))
    expect(result.current.error).toBe('Invalid username or password.')
  })

  it('loginWithPasskey() resolves and updates status to authenticated on success', async () => {
    vi.spyOn(webauthnLib, 'authenticateDiscoverable').mockResolvedValue({ id: 'stub' })
    const store = createTestStore({ status: 'unauthenticated', user: null, error: null })
    const { result } = renderHook(() => useAuth(), { wrapper: wrapperFor(store) })

    await act(async () => {
      await result.current.loginWithPasskey()
    })

    await waitFor(() => expect(result.current.status).toBe('authenticated'))
    expect(result.current.user).toEqual(fixtureUser)
  })

  it('loginWithPasskey() rejects and sets a Face-ID-specific error on a 401', async () => {
    vi.spyOn(webauthnLib, 'authenticateDiscoverable').mockResolvedValue({ id: 'stub' })
    server.use(
      http.post('*/api/v1/webauthn/login/finish', () =>
        HttpResponse.json({ error: { message: 'bad credentials' } }, { status: 401 }),
      ),
    )
    const store = createTestStore({ status: 'unauthenticated', user: null, error: null })
    const { result } = renderHook(() => useAuth(), { wrapper: wrapperFor(store) })

    await act(async () => {
      await expect(result.current.loginWithPasskey()).rejects.toBeTruthy()
    })

    await waitFor(() => expect(result.current.status).toBe('error'))
    expect(result.current.error).toBe('Face ID sign-in failed — try again or use your password.')
  })

  it('logout() clears the session', async () => {
    const store = createTestStore({ status: 'authenticated', user: fixtureUser, error: null })
    const { result } = renderHook(() => useAuth(), { wrapper: wrapperFor(store) })

    await act(async () => {
      await result.current.logout()
    })

    await waitFor(() => expect(result.current.status).toBe('unauthenticated'))
    expect(result.current.user).toBeNull()
  })
})
