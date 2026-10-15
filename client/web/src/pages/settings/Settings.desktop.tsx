import { IconBell, IconChart, IconDownload, IconInfo, IconInstall, IconTag } from '@/components/icons'
import { Button } from '@/components/ui/Button'
import { InstallInstructions } from './components/InstallInstructions'
import { SettingsRow } from './components/SettingsRow'
import type { SettingsViewProps } from './types'

export default function SettingsDesktop({ displayName, username, currencyCode, onLogout, install }: SettingsViewProps) {
  return (
    <div className="flex flex-col gap-6">
      <h1 className="m-0 text-2xl font-bold">Settings</h1>

      <div className="grid grid-cols-[280px_1fr] gap-8">
        <div className="flex flex-col items-center gap-3 rounded-2xl border border-line bg-surface p-6 text-center">
          <div className="flex h-16 w-16 items-center justify-center rounded-full bg-accent-soft text-2xl font-bold text-accent-strong">
            {displayName.charAt(0)}
          </div>
          <div className="flex flex-col">
            <strong className="text-[15px]">{displayName}</strong>
            <span className="text-xs text-muted">@{username}</span>
          </div>
          <Button
            onClick={onLogout}
            variant="danger"
            className="mt-2 w-full rounded-xl border border-line bg-surface-sunken py-2.5 text-[13px] font-semibold text-critical"
          >
            Log out
          </Button>
        </div>

        <div className="flex flex-col gap-3">
          <ul className="flex flex-col rounded-2xl border border-line bg-surface px-5">
            <SettingsRow Icon={IconTag} label="Categories" to="/settings/categories" />
            <SettingsRow
              Icon={IconChart}
              label="Currency"
              to="/settings/currencies"
              trailing={<small className="font-medium text-muted">{currencyCode ?? 'Not set'}</small>}
            />
            <SettingsRow Icon={IconBell} label="Notifications" />
            <SettingsRow Icon={IconDownload} label="Export data" />
            {install.show && (
              <SettingsRow Icon={IconInstall} label="Install app" trailing={install.trailing} onClick={install.onClick} />
            )}
            <SettingsRow Icon={IconInfo} label="About Joint Wallet Watch" />
          </ul>

          {install.platform && (
            <InstallInstructions
              platform={install.platform}
              compact={false}
              open={install.instructionsOpen}
              onClose={install.onCloseInstructions}
            />
          )}
        </div>
      </div>
    </div>
  )
}
