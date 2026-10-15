import { configureStore } from '@reduxjs/toolkit'
import { http, HttpResponse } from 'msw'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { MOCK_USER } from '@/api/mock/data'
import { getAuthToken, setAuthToken } from '@/lib/authToken'
import * as designModeModule from '@/lib/designMode'
import { authReducer, login, logout, restoreSession } from '@/store/authSlice'
import { fixtureUser } from '@/test/fixtures'
import { server } from '@/test/server'

function createStore() {
  return configureStore({ reducer: { auth: authReducer } })
}

beforeEach(() => {
  setAuthToken(null)
})

afterEach(() => {
  vi.restoreAllMocks()
  setAuthToken(null)
})

describe('restoreSession', () => {
  it('skips the request (condition guard) when there is no stored token', async () => {
    const meSpy = vi.fn()
    server.use(
      http.get('*/api/v1/auth/me', () => {
        meSpy()
        return HttpResponse.json({ data: fixtureUser })
      }),
    )
    setAuthToken(null)
    const store = createStore()
    const statusBefore = store.getState().auth.status

    await store.dispatch(restoreSession())

    expect(meSpy).not.toHaveBeenCalled()
    expect(store.getState().auth.status).toBe(statusBefore)
  })

  it('authenticates with the restored user when a token is present and /auth/me succeeds', async () => {
    setAuthToken('stored-token')
    const store = createStore()

    await store.dispatch(restoreSession())

    expect(store.getState().auth).toEqual({ status: 'authenticated', user: fixtureUser, error: null })
  })

  it('clears the token and becomes unauthenticated when /auth/me fails', async () => {
    server.use(http.get('*/api/v1/auth/me', () => HttpResponse.json({ error: { message: 'unauthorized' } }, { status: 401 })))
    setAuthToken('stale-token')
    const store = createStore()

    await store.dispatch(restoreSession())

    const state = store.getState().auth
    expect(state.status).toBe('unauthenticated')
    expect(state.user).toBeNull()
    expect(getAuthToken()).toBeNull()
  })
})

describe('login', () => {
  it('authenticates on success', async () => {
    const store = createStore()

    await store.dispatch(login({ username: 'waldemar', password: 'correct-password' }))

    const state = store.getState().auth
    expect(state.status).toBe('authenticated')
    expect(state.user).toEqual(fixtureUser)
    expect(state.error).toBeNull()
    expect(getAuthToken()).toBe('fixture-access-token')
  })

  it('surfaces a fixed message and does not authenticate on invalid credentials (401)', async () => {
    server.use(http.post('*/api/v1/auth/login', () => HttpResponse.json({ error: { message: 'bad creds' } }, { status: 401 })))
    const store = createStore()

    await store.dispatch(login({ username: 'waldemar', password: 'wrong-password' }))

    const state = store.getState().auth
    expect(state.status).toBe('error')
    expect(state.user).toBeNull()
    expect(state.error).toBe('Invalid username or password.')
  })

  it('short-circuits to MOCK_USER in design mode without hitting the network', async () => {
    const loginSpy = vi.fn()
    server.use(
      http.post('*/api/v1/auth/login', () => {
        loginSpy()
        return HttpResponse.json({ error: { message: 'should not be called' } }, { status: 500 })
      }),
    )
    vi.spyOn(designModeModule, 'isDesignMode').mockReturnValue(true)
    const store = createStore()

    await store.dispatch(login({ username: 'anyone', password: 'anything' }))

    expect(loginSpy).not.toHaveBeenCalled()
    const state = store.getState().auth
    expect(state.status).toBe('authenticated')
    expect(state.user).toEqual(MOCK_USER)
  })
})

describe('logout', () => {
  it('clears auth state and the stored token', async () => {
    const store = createStore()
    await store.dispatch(login({ username: 'waldemar', password: 'correct-password' }))
    expect(store.getState().auth.status).toBe('authenticated')

    await store.dispatch(logout())

    expect(store.getState().auth).toEqual({ status: 'unauthenticated', user: null, error: null })
    expect(getAuthToken()).toBeNull()
  })
})
