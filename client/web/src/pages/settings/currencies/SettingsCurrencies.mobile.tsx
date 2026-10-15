import { IconChevronRight, IconClose } from '@/components/icons'
import { Button } from '@/components/ui/Button'
import { IconButton } from '@/components/ui/IconButton'
import { TextInput } from '@/components/ui/TextInput'
import { Loader } from '@/components/ui/Loader'
import type { SettingsCurrenciesViewProps } from './types'

export default function SettingsCurrenciesMobile({
  currencies,
  isLoading,
  onMakeBase,
  code,
  onCodeChange,
  name,
  onNameChange,
  onCreate,
  creating,
  hasError,
  onClose,
}: SettingsCurrenciesViewProps) {
  return (
    <div className="flex flex-col gap-5 px-4 pb-4 pt-6">
      <div className="flex items-center justify-between">
        <h2 className="m-0 text-[19px] font-bold">Currencies</h2>
        <IconButton
          aria-label="Close"
          onClick={onClose}
          icon={<IconClose className="h-4 w-4" />}
          className="flex h-[34px] w-[34px] items-center justify-center rounded-[10px] border border-line bg-surface-sunken"
        />
      </div>

      {isLoading && (
        <div className="flex justify-center py-8">
          <Loader />
        </div>
      )}
      {!isLoading && currencies?.length === 0 && (
        <p className="m-0 text-xs text-muted">No currencies yet — add one below. The first one becomes the base currency.</p>
      )}

      <ul className="flex flex-col">
        {currencies?.map((currency) => (
          <li
            key={currency.code}
            className="flex items-center justify-between gap-3 border-b border-line py-3 text-[13.5px] last:border-b-0"
          >
            <div className="flex flex-col">
              <strong>{currency.code}</strong>
              <span className="text-xs text-muted">{currency.name}</span>
            </div>
            {currency.is_base ? (
              <span className="rounded-full bg-accent-soft px-2.5 py-1 text-[11px] font-bold text-accent-strong">Base</span>
            ) : (
              <Button
                variant="link"
                onClick={() => onMakeBase(currency)}
                className="flex items-center gap-1 text-xs font-semibold text-accent"
              >
                Make base
                <IconChevronRight className="h-3.5 w-3.5" />
              </Button>
            )}
          </li>
        ))}
      </ul>

      <div className="flex flex-col gap-2 rounded-xl border border-line bg-surface p-3.5">
        <p className="m-0 text-xs font-semibold uppercase tracking-wide text-muted">Add currency</p>
        <div className="flex gap-2">
          <TextInput
            type="text"
            value={code}
            onChange={(e) => onCodeChange(e.target.value.toUpperCase())}
            placeholder="PLN"
            maxLength={3}
            className="w-20 rounded-[10px] bg-surface px-2.5 py-2 uppercase sm:text-[13px]"
          />
          <TextInput
            type="text"
            value={name}
            onChange={(e) => onNameChange(e.target.value)}
            placeholder="Polish złoty"
            className="flex-1 rounded-[10px] bg-surface px-2.5 py-2 sm:text-[13px]"
          />
        </div>
        <Button
          disabled={creating || code.trim().length !== 3 || !name.trim()}
          onClick={onCreate}
          variant="primary"
          className="rounded-[10px] bg-accent py-2.5 text-[13px] font-semibold text-on-accent disabled:opacity-40"
        >
          Add
        </Button>
        {hasError && <p className="m-0 text-[11.5px] text-critical">That didn't work — check the code and try again.</p>}
      </div>
    </div>
  )
}
