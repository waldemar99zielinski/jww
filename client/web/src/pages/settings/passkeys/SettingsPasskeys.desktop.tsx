import { IconClose, IconFaceId, IconTrash } from '@/components/icons'
import { Button } from '@/components/ui/Button'
import { IconButton } from '@/components/ui/IconButton'
import { Loader } from '@/components/ui/Loader'
import { TextInput } from '@/components/ui/TextInput'
import type { SettingsPasskeysViewProps } from './types'

function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString('en-GB', { day: 'numeric', month: 'short', year: 'numeric' })
}

export default function SettingsPasskeysDesktop({
  credentials,
  isLoading,
  name,
  onNameChange,
  onEnroll,
  enrolling,
  enrollError,
  onDelete,
  onClose,
}: SettingsPasskeysViewProps) {
  return (
    <div className="mx-auto flex max-w-2xl flex-col gap-6 rounded-2xl border border-line bg-surface p-8">
      <div className="flex items-center justify-between">
        <h1 className="m-0 text-2xl font-bold">Face ID &amp; biometrics</h1>
        <IconButton
          aria-label="Close"
          onClick={onClose}
          icon={<IconClose className="h-4 w-4" />}
          className="flex h-9 w-9 items-center justify-center rounded-[10px] border border-line bg-surface-sunken"
        />
      </div>

      <p className="m-0 text-xs text-muted">
        Sign in without your password using Face ID, Touch ID, or your device's own biometric unlock. Each
        device you enroll gets its own entry below.
      </p>

      {isLoading && (
        <div className="flex justify-center py-8">
          <Loader />
        </div>
      )}
      {!isLoading && credentials?.length === 0 && (
        <p className="m-0 text-xs text-muted">No devices enrolled yet — add this one below.</p>
      )}

      <ul className="flex flex-col gap-2">
        {credentials?.map((credential) => (
          <li
            key={credential.id}
            className="flex items-center justify-between gap-3 rounded-xl border border-line bg-surface-sunken px-3.5 py-3"
          >
            <div className="flex items-center gap-3">
              <IconFaceId className="h-5 w-5 flex-none text-accent-strong" />
              <div className="flex flex-col">
                <span className="text-[13.5px] font-semibold">{credential.name}</span>
                <small className="text-muted">
                  Added {formatDate(credential.created_at)}
                  {credential.last_used_at ? ` · last used ${formatDate(credential.last_used_at)}` : ''}
                </small>
              </div>
            </div>
            <IconButton
              aria-label={`Remove ${credential.name}`}
              onClick={() => onDelete(credential)}
              icon={<IconTrash className="h-4 w-4" />}
              className="flex h-9 w-9 items-center justify-center rounded-[10px] border border-line bg-surface text-critical"
            />
          </li>
        ))}
      </ul>

      <div className="flex flex-col gap-2 rounded-xl border border-line bg-surface-sunken p-3.5">
        <p className="m-0 text-xs font-semibold uppercase tracking-wide text-muted">Add this device</p>
        <div className="flex gap-2">
          <TextInput
            type="text"
            value={name}
            onChange={(e) => onNameChange(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && onEnroll()}
            placeholder="e.g. My iPhone"
            className="flex-1 rounded-[10px] bg-surface px-2.5 py-2 sm:text-[13px]"
          />
          <Button
            disabled={enrolling || !name.trim()}
            onClick={onEnroll}
            variant="primary"
            className="rounded-[10px] bg-accent px-3.5 py-2 text-[13px] font-semibold text-on-accent disabled:opacity-40"
          >
            {enrolling ? 'Confirming…' : 'Add'}
          </Button>
        </div>
        {enrollError && <p className="m-0 text-[11.5px] text-critical">{enrollError}</p>}
      </div>
    </div>
  )
}
