import { afterEach, describe, expect, it, vi } from 'vitest'
import { authenticateDiscoverable, isPlatformAuthenticatorAvailable, registerCredential } from '@/lib/webauthn'

/** Minimal stand-in for the browser's PublicKeyCredential — jsdom implements none of this. */
class StubPublicKeyCredential {
  static isUserVerifyingPlatformAuthenticatorAvailable = vi.fn()
  static parseCreationOptionsFromJSON = vi.fn((json: unknown) => json)
  static parseRequestOptionsFromJSON = vi.fn((json: unknown) => json)
  toJSON() {
    return { id: 'stub-id', rawId: 'stub-raw-id', type: 'public-key' }
  }
}

afterEach(() => {
  vi.unstubAllGlobals()
  vi.restoreAllMocks()
})

describe('isPlatformAuthenticatorAvailable', () => {
  it('is false when the browser has no PublicKeyCredential global at all', async () => {
    vi.stubGlobal('PublicKeyCredential', undefined)
    expect(await isPlatformAuthenticatorAvailable()).toBe(false)
  })

  it('reflects the browser check when PublicKeyCredential exists', async () => {
    StubPublicKeyCredential.isUserVerifyingPlatformAuthenticatorAvailable.mockResolvedValue(true)
    vi.stubGlobal('PublicKeyCredential', StubPublicKeyCredential)
    expect(await isPlatformAuthenticatorAvailable()).toBe(true)
  })

  it('is false (not a thrown error) when the browser check itself rejects', async () => {
    StubPublicKeyCredential.isUserVerifyingPlatformAuthenticatorAvailable.mockRejectedValue(new Error('nope'))
    vi.stubGlobal('PublicKeyCredential', StubPublicKeyCredential)
    expect(await isPlatformAuthenticatorAvailable()).toBe(false)
  })
})

describe('registerCredential', () => {
  it('parses the options, runs navigator.credentials.create, and returns the serialized credential', async () => {
    vi.stubGlobal('PublicKeyCredential', StubPublicKeyCredential)
    const create = vi.fn().mockResolvedValue(new StubPublicKeyCredential())
    vi.stubGlobal('navigator', { credentials: { create, get: vi.fn() } })

    const result = await registerCredential({ challenge: 'abc' })

    expect(StubPublicKeyCredential.parseCreationOptionsFromJSON).toHaveBeenCalledWith({ challenge: 'abc' })
    expect(create).toHaveBeenCalledWith({ publicKey: { challenge: 'abc' } })
    expect(result).toEqual({ id: 'stub-id', rawId: 'stub-raw-id', type: 'public-key' })
  })

  it('throws when navigator.credentials.create does not return a PublicKeyCredential', async () => {
    vi.stubGlobal('PublicKeyCredential', StubPublicKeyCredential)
    vi.stubGlobal('navigator', { credentials: { create: vi.fn().mockResolvedValue(null), get: vi.fn() } })

    await expect(registerCredential({})).rejects.toThrow()
  })
})

describe('authenticateDiscoverable', () => {
  it('parses the options, runs navigator.credentials.get with no allowCredentials, and returns the serialized credential', async () => {
    vi.stubGlobal('PublicKeyCredential', StubPublicKeyCredential)
    const get = vi.fn().mockResolvedValue(new StubPublicKeyCredential())
    vi.stubGlobal('navigator', { credentials: { create: vi.fn(), get } })

    const result = await authenticateDiscoverable({ challenge: 'xyz' })

    expect(StubPublicKeyCredential.parseRequestOptionsFromJSON).toHaveBeenCalledWith({ challenge: 'xyz' })
    expect(get).toHaveBeenCalledWith({ publicKey: { challenge: 'xyz' } })
    expect(result).toEqual({ id: 'stub-id', rawId: 'stub-raw-id', type: 'public-key' })
  })

  it('throws when navigator.credentials.get does not return a PublicKeyCredential', async () => {
    vi.stubGlobal('PublicKeyCredential', StubPublicKeyCredential)
    vi.stubGlobal('navigator', { credentials: { create: vi.fn(), get: vi.fn().mockResolvedValue(undefined) } })

    await expect(authenticateDiscoverable({})).rejects.toThrow()
  })
})
