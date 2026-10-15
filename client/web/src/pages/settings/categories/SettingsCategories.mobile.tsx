import type { TransactionType } from '@/api/types'
import { IconClose } from '@/components/icons'
import { Button } from '@/components/ui/Button'
import { IconButton } from '@/components/ui/IconButton'
import { TextInput } from '@/components/ui/TextInput'
import { Loader } from '@/components/ui/Loader'
import { SegmentedControl } from '@/components/ui/SegmentedControl'
import type { SettingsCategoriesViewProps } from './types'

const tabs: { label: string; value: TransactionType }[] = [
  { label: 'Expense', value: 'expense' },
  { label: 'Income', value: 'income' },
]

export default function SettingsCategoriesMobile({
  type,
  onSelectType,
  categories,
  isLoading,
  onToggleActive,
  name,
  onNameChange,
  onCreate,
  creating,
  createError,
  onClose,
}: SettingsCategoriesViewProps) {
  return (
    <div className="flex flex-col gap-5 px-4 pb-4 pt-6">
      <div className="flex items-center justify-between">
        <h2 className="m-0 text-[19px] font-bold">Categories</h2>
        <IconButton
          aria-label="Close"
          onClick={onClose}
          icon={<IconClose className="h-4 w-4" />}
          className="flex h-[34px] w-[34px] items-center justify-center rounded-[10px] border border-line bg-surface-sunken"
        />
      </div>

      <SegmentedControl value={type} options={tabs} onChange={onSelectType} />

      {isLoading && (
        <div className="flex justify-center py-8">
          <Loader />
        </div>
      )}
      {!isLoading && categories?.length === 0 && (
        <p className="m-0 text-xs text-muted">No {type} categories yet — add one below.</p>
      )}

      <ul className="flex flex-col">
        {categories?.map((category) => (
          <li
            key={category.id}
            className="flex items-center justify-between gap-3 border-b border-line py-3 text-[13.5px] last:border-b-0"
          >
            <span className={category.active ? '' : 'text-muted line-through'}>{category.name}</span>
            <Button variant="link" onClick={() => onToggleActive(category)} className="text-xs font-semibold text-accent">
              {category.active ? 'Deactivate' : 'Activate'}
            </Button>
          </li>
        ))}
      </ul>

      <div className="flex flex-col gap-2 rounded-xl border border-line bg-surface p-3.5">
        <p className="m-0 text-xs font-semibold uppercase tracking-wide text-muted">Add {type} category</p>
        <div className="flex gap-2">
          <TextInput
            type="text"
            value={name}
            onChange={(e) => onNameChange(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && onCreate()}
            placeholder="e.g. Groceries"
            className="flex-1 rounded-[10px] bg-surface px-2.5 py-2 sm:text-[13px]"
          />
          <Button
            disabled={creating || !name.trim()}
            onClick={onCreate}
            variant="primary"
            className="rounded-[10px] bg-accent px-3.5 py-2 text-[13px] font-semibold text-on-accent disabled:opacity-40"
          >
            Add
          </Button>
        </div>
        {createError && <p className="m-0 text-[11.5px] text-critical">That didn't work — try again.</p>}
      </div>
    </div>
  )
}
