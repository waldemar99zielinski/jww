import { apiFetch } from '@/api/client'
import type { WebAuthnCredentialDto } from '@/api/types'

/**
 * Like `auth.ts`, this file doesn't follow the mock-branch shape — WebAuthn needs a real browser
 * ceremony (Face ID/Touch ID/...) the design-mode "no backend" philosophy can't meaningfully fake,
 * so the Face ID UI is gated on `isPlatformAuthenticatorAvailable()` one layer up instead (see
 * `hooks/useAuth.ts` and `pages/settings/passkeys/`).
 */

export interface WebAuthnBeginResponse {
  // The ceremony's options object, forwarded verbatim to
  // PublicKeyCredential.parseCreationOptionsFromJSON()/parseRequestOptionsFromJSON() —
  // not a cryptographic key, named to match the browser API shape it feeds
  // (navigator.credentials.create({ publicKey: options })).
  options: unknown
  challenge_token: string
}

export function registerBegin(): Promise<WebAuthnBeginResponse> {
  return apiFetch<WebAuthnBeginResponse>('/webauthn/credentials/register/begin', { method: 'POST' })
}

export function registerFinish(challengeToken: string, name: string, credential: unknown): Promise<WebAuthnCredentialDto> {
  return apiFetch<WebAuthnCredentialDto>('/webauthn/credentials/register/finish', {
    method: 'POST',
    body: JSON.stringify({ challenge_token: challengeToken, name, credential }),
  })
}

export function listCredentials(): Promise<WebAuthnCredentialDto[]> {
  return apiFetch<WebAuthnCredentialDto[]>('/webauthn/credentials')
}

export function deleteCredential(id: string): Promise<void> {
  return apiFetch<void>(`/webauthn/credentials/${id}`, { method: 'DELETE' })
}

export function loginBegin(): Promise<WebAuthnBeginResponse> {
  return apiFetch<WebAuthnBeginResponse>('/webauthn/login/begin', { method: 'POST' })
}

export interface WebAuthnLoginFinishResponse {
  access_token: string
  token_type: string
}

export function loginFinish(challengeToken: string, credential: unknown): Promise<WebAuthnLoginFinishResponse> {
  return apiFetch<WebAuthnLoginFinishResponse>('/webauthn/login/finish', {
    method: 'POST',
    body: JSON.stringify({ challenge_token: challengeToken, credential }),
  })
}
