import { lazy, useState } from 'react'
import { Navigate, useNavigate } from 'react-router-dom'
import type { WebAuthnCredentialDto } from '@/api/types'
import { ResponsiveView } from '@/components/ResponsiveView'
import { useDeleteWebAuthnCredential, useEnrollWebAuthnCredential, useWebAuthnCredentials } from '@/hooks/useWebAuthnCredentials'
import { isDesignMode } from '@/lib/designMode'
import { RemovePasskeyAction } from './components/RemovePasskeyAction'
import type { SettingsPasskeysViewProps } from './types'

const Mobile = lazy(() => import('./SettingsPasskeys.mobile'))
const Desktop = lazy(() => import('./SettingsPasskeys.desktop'))

export default function SettingsPasskeysPage() {
  const navigate = useNavigate()
  const [name, setName] = useState('')
  const [removeTarget, setRemoveTarget] = useState<WebAuthnCredentialDto | null>(null)
  const [removeFailed, setRemoveFailed] = useState(false)
  const { data: credentials, isLoading } = useWebAuthnCredentials(!isDesignMode())
  const enroll = useEnrollWebAuthnCredential()
  const deleteCredential = useDeleteWebAuthnCredential()

  // The Settings row that links here is already hidden in design mode (see
  // `usePlatformAuthenticator`), but the route itself is reachable directly
  // (typed URL, bookmark, back/forward) — guard here too, since this page's
  // hooks call the real backend and have no mock-branch to fall back to
  // (see `api/webauthn.ts`).
  if (isDesignMode()) return <Navigate to="/settings" replace />

  function handleEnroll() {
    const trimmed = name.trim()
    if (!trimmed) return
    enroll.mutate(trimmed, { onSuccess: () => setName('') })
  }

  function handleRequestRemove(credential: WebAuthnCredentialDto) {
    setRemoveFailed(false)
    setRemoveTarget(credential)
  }

  function handleCloseRemove() {
    if (deleteCredential.isPending) return
    setRemoveTarget(null)
    setRemoveFailed(false)
  }

  async function handleConfirmRemove() {
    if (!removeTarget) return
    setRemoveFailed(false)
    try {
      await deleteCredential.mutateAsync(removeTarget.id)
      setRemoveTarget(null)
    } catch {
      setRemoveFailed(true)
    }
  }

  const props: SettingsPasskeysViewProps = {
    credentials,
    isLoading,
    name,
    onNameChange: setName,
    onEnroll: handleEnroll,
    enrolling: enroll.isPending,
    enrollError: enroll.isError ? "That didn't work — try again." : null,
    onDelete: handleRequestRemove,
    onClose: () => navigate(-1),
  }

  return (
    <>
      <ResponsiveView mobile={Mobile} desktop={Desktop} props={props} />
      <RemovePasskeyAction
        open={removeTarget !== null}
        credentialName={removeTarget?.name ?? ''}
        isRemoving={deleteCredential.isPending}
        error={removeFailed}
        onConfirm={() => void handleConfirmRemove()}
        onClose={handleCloseRemove}
      />
    </>
  )
}
