/**
 * Thin wrapper around the browser's WebAuthn API (Face ID/Touch ID/Windows Hello/... — the
 * browser/OS decide which, this code never knows or cares). No extra dependency: modern Safari
 * (17+) and Chrome both implement `PublicKeyCredential.parseCreationOptionsFromJSON()` /
 * `.parseRequestOptionsFromJSON()` and `credential.toJSON()`, which handle all the
 * base64url<->ArrayBuffer conversion the raw API needs, so no hand-rolled encoding is required.
 */

/** Feature probe used to show/hide the Face ID button and the Settings enrollment row. */
export async function isPlatformAuthenticatorAvailable(): Promise<boolean> {
  if (typeof window === 'undefined' || !window.PublicKeyCredential) return false
  try {
    return await PublicKeyCredential.isUserVerifyingPlatformAuthenticatorAvailable()
  } catch {
    return false
  }
}

/**
 * Runs the enrollment ceremony: prompts Face ID/Touch ID/etc. and returns the attestation
 * response as opaque JSON (whatever `.toJSON()` produces), forwarded to the server as-is.
 */
export async function registerCredential(optionsJSON: unknown): Promise<unknown> {
  const options = PublicKeyCredential.parseCreationOptionsFromJSON(
    optionsJSON as Parameters<typeof PublicKeyCredential.parseCreationOptionsFromJSON>[0],
  )
  const credential = await navigator.credentials.create({ publicKey: options })
  if (!(credential instanceof PublicKeyCredential)) {
    throw new Error('WebAuthn registration did not return a credential')
  }
  return credential.toJSON()
}

/** Runs the discoverable (usernameless) login ceremony: prompts Face ID/Touch ID/etc. with no `allowCredentials`. */
export async function authenticateDiscoverable(optionsJSON: unknown): Promise<unknown> {
  const options = PublicKeyCredential.parseRequestOptionsFromJSON(
    optionsJSON as Parameters<typeof PublicKeyCredential.parseRequestOptionsFromJSON>[0],
  )
  const credential = await navigator.credentials.get({ publicKey: options })
  if (!(credential instanceof PublicKeyCredential)) {
    throw new Error('WebAuthn login did not return a credential')
  }
  return credential.toJSON()
}
