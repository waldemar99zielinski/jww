import { useEffect, useState } from 'react'
import { isPlatformAuthenticatorAvailable } from '@/lib/webauthn'
import { isDesignMode } from '@/lib/designMode'

/**
 * Feature-detects a platform authenticator (Face ID/Touch ID/Windows Hello/...), for showing or
 * hiding the Face ID login button and the Settings enrollment row. Starts `false` and flips once
 * the async browser check resolves — never `true` on first render, so nothing flashes in only to
 * disappear. Always `false` in design mode: WebAuthn needs a real ceremony design mode's
 * "no backend" philosophy can't fake (see `api/webauthn.ts`).
 */
export function usePlatformAuthenticator(): boolean {
  const [available, setAvailable] = useState(false)

  useEffect(() => {
    if (isDesignMode()) return
    let cancelled = false
    void isPlatformAuthenticatorAvailable().then((result) => {
      if (!cancelled) setAvailable(result)
    })
    return () => {
      cancelled = true
    }
  }, [])

  return available
}
