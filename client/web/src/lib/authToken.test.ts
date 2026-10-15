import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

const STORAGE_KEY = 'auth_token'

describe('authToken', () => {
  beforeEach(() => {
    localStorage.clear()
    vi.resetModules()
  })

  afterEach(() => {
    localStorage.clear()
  })

  it('getAuthToken returns null when nothing is stored', async () => {
    const { getAuthToken } = await import('@/lib/authToken')
    expect(getAuthToken()).toBeNull()
  })

  it('getAuthToken reflects a token already in localStorage at module load time', async () => {
    localStorage.setItem(STORAGE_KEY, 'preexisting-token')
    const { getAuthToken } = await import('@/lib/authToken')
    expect(getAuthToken()).toBe('preexisting-token')
  })

  it('setAuthToken stores a token and getAuthToken returns it', async () => {
    const { getAuthToken, setAuthToken } = await import('@/lib/authToken')
    setAuthToken('new-token')
    expect(getAuthToken()).toBe('new-token')
    expect(localStorage.getItem(STORAGE_KEY)).toBe('new-token')
  })

  it('setAuthToken(null) clears the stored token', async () => {
    const { getAuthToken, setAuthToken } = await import('@/lib/authToken')
    setAuthToken('new-token')
    setAuthToken(null)
    expect(getAuthToken()).toBeNull()
    expect(localStorage.getItem(STORAGE_KEY)).toBeNull()
  })
})
