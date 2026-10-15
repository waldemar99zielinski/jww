import { IconBell, IconChart, IconDownload, IconInfo, IconInstall, IconTag } from '@/components/icons'
import { Button } from '@/components/ui/Button'
import { InstallInstructions } from './components/InstallInstructions'
import { SettingsRow } from './components/SettingsRow'
import type { SettingsViewProps } from './types'

export default function SettingsMobile({ displayName, username, currencyCode, onLogout, install }: SettingsViewProps) {
  return (
    <div className="flex flex-col gap-5 px-4 pb-4 pt-6">
      <div className="flex items-center gap-3">
        <div className="flex h-11 w-11 flex-none items-center justify-center rounded-full bg-accent-soft text-base font-bold text-accent-strong">
          {displayName.charAt(0)}
        </div>
        <div className="flex flex-col">
          <strong className="text-[14.5px]">{displayName}</strong>
          <span className="text-xs text-muted">@{username}</span>
        </div>
      </div>

      <ul className="flex flex-col">
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

      <Button
        onClick={onLogout}
        variant="danger"
        className="rounded-xl border border-line bg-surface-sunken py-3 text-[13.5px] font-semibold text-critical"
      >
        Log out
      </Button>

      {install.platform && (
        <InstallInstructions
          platform={install.platform}
          compact
          open={install.instructionsOpen}
          onClose={install.onCloseInstructions}
        />
      )}
    </div>
  )
}
