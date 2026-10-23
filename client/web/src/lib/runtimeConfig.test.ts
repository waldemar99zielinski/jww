import { afterEach, describe, expect, it } from 'vitest'
import { runtimeConfig } from './runtimeConfig'

describe('runtimeConfig', () => {
  afterEach(() => {
    delete window.__JWW_CONFIG__
  })

  it('returns an empty config when no server config was loaded', () => {
    delete window.__JWW_CONFIG__
    expect(runtimeConfig()).toEqual({})
  })

  it('returns the config loaded before the application bundle', () => {
    window.__JWW_CONFIG__ = { apiBaseUrl: 'https://api.example.com' }
    expect(runtimeConfig()).toEqual({ apiBaseUrl: 'https://api.example.com' })
  })
})
