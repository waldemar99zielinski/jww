import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import type { TransactionSortOrder } from '@/api/transactions'
import { IconChevronDown } from '@/components/icons'
import { BottomSheet } from '@/components/ui/BottomSheet'
import { Button } from '@/components/ui/Button'
import { DateInput } from '@/components/ui/DateInput'
import { FieldLabel } from '@/components/ui/FieldLabel'
import { SearchableSelect, type SearchableSelectOption } from '@/components/ui/SearchableSelect'
import { SelectInput } from '@/components/ui/SelectInput'
import { TextInput } from '@/components/ui/TextInput'
import { getCategoryVisual } from '@/lib/categoryVisuals'
import { cssVars } from '@/lib/cssVars'
import { filterLabels, sortOrderLabels } from '../filterLabels'
import type { TransactionFilterValue, TransactionsViewProps } from '../types'
import { FilterButton } from './FilterButton'

type TransactionFiltersDropdownProps = Pick<
  TransactionsViewProps,
  | 'filter'
  | 'onFilterChange'
  | 'userFilter'
  | 'onUserFilterChange'
  | 'users'
  | 'categoryFilter'
  | 'onCategoryFilterChange'
  | 'categories'
  | 'nameFilter'
  | 'onNameFilterChange'
  | 'dateFrom'
  | 'onDateFromChange'
  | 'dateTo'
  | 'onDateToChange'
  | 'sortOrder'
  | 'onSortOrderChange'
> & {
  align?: 'left' | 'right'
  compact?: boolean
}

export function TransactionFiltersDropdown({
  filter,
  onFilterChange,
  userFilter,
  onUserFilterChange,
  users,
  categoryFilter,
  onCategoryFilterChange,
  categories,
  nameFilter,
  onNameFilterChange,
  dateFrom,
  onDateFromChange,
  dateTo,
  onDateToChange,
  sortOrder,
  onSortOrderChange,
  align = 'right',
  compact = false,
}: TransactionFiltersDropdownProps) {
  const [filtersOpen, setFiltersOpen] = useState(false)
  const [draftFilter, setDraftFilter] = useState<TransactionFilterValue>(filter)
  const [draftUserFilter, setDraftUserFilter] = useState(userFilter)
  const [draftCategoryFilter, setDraftCategoryFilter] = useState(categoryFilter)
  const [draftNameFilter, setDraftNameFilter] = useState(nameFilter)
  const [draftDateFrom, setDraftDateFrom] = useState(dateFrom)
  const [draftDateTo, setDraftDateTo] = useState(dateTo)
  const [draftSortOrder, setDraftSortOrder] = useState<TransactionSortOrder>(sortOrder)
  const rootRef = useRef<HTMLDivElement>(null)
  const hasInvalidDateRange = draftDateFrom !== '' && draftDateTo !== '' && draftDateFrom > draftDateTo
  const activeFilters = [
    filter !== 'all',
    userFilter !== '',
    categoryFilter !== '',
    nameFilter.trim() !== '',
    dateFrom !== '',
    dateTo !== '',
  ].filter(Boolean).length
  const activeDraftFilters = [
    draftFilter !== 'all',
    draftUserFilter !== '',
    draftCategoryFilter !== '',
    draftNameFilter.trim() !== '',
    draftDateFrom !== '',
    draftDateTo !== '',
  ].filter(Boolean).length
  const draftHasChanges = activeDraftFilters > 0 || draftSortOrder !== 'desc'
  const userOptions = useMemo<SearchableSelectOption[]>(
    () =>
      users.map((member) => ({
        value: member.id,
        label: member.display_name,
        trailing: <span className="text-xs font-medium text-muted">@{member.username}</span>,
      })),
    [users]
  )
  const categoryOptions = useMemo<SearchableSelectOption[]>(
    () =>
      categories
        .filter((category) => draftFilter === 'all' || category.type === draftFilter)
        .map((category) => {
          const { Icon, colorVar } = getCategoryVisual(category.name)
          return {
            value: category.id,
            label: category.name,
            icon: (
              <span
                className="category-tint flex h-6 w-6 flex-none items-center justify-center rounded-md"
                style={cssVars({ '--tint': colorVar })}
              >
                <Icon className="h-3.5 w-3.5" />
              </span>
            ),
            trailing: <span className="text-[11px] font-bold uppercase text-muted">{category.type}</span>,
          }
        }),
    [categories, draftFilter]
  )

  function syncDraftFromApplied() {
    setDraftFilter(filter)
    setDraftUserFilter(userFilter)
    setDraftCategoryFilter(categoryFilter)
    setDraftNameFilter(nameFilter)
    setDraftDateFrom(dateFrom)
    setDraftDateTo(dateTo)
    setDraftSortOrder(sortOrder)
  }

  const applyDraft = useCallback(() => {
    if (hasInvalidDateRange) return false
    onFilterChange(draftFilter)
    onUserFilterChange(draftUserFilter)
    onCategoryFilterChange(draftCategoryFilter)
    onNameFilterChange(draftNameFilter)
    onDateFromChange(draftDateFrom)
    onDateToChange(draftDateTo)
    onSortOrderChange(draftSortOrder)
    return true
  }, [
    draftCategoryFilter,
    draftDateFrom,
    draftDateTo,
    draftFilter,
    draftNameFilter,
    draftSortOrder,
    draftUserFilter,
    hasInvalidDateRange,
    onCategoryFilterChange,
    onDateFromChange,
    onDateToChange,
    onFilterChange,
    onNameFilterChange,
    onSortOrderChange,
    onUserFilterChange,
  ])

  const closeWithApply = useCallback(() => {
    if (!applyDraft()) return false
    setFiltersOpen(false)
    return true
  }, [applyDraft])

  useEffect(() => {
    if (compact || !filtersOpen) return
    function handlePointerDown(event: PointerEvent) {
      if (!rootRef.current?.contains(event.target as Node)) closeWithApply()
    }
    document.addEventListener('pointerdown', handlePointerDown)
    return () => document.removeEventListener('pointerdown', handlePointerDown)
  }, [closeWithApply, compact, filtersOpen])

  function clearFilters() {
    setDraftFilter('all')
    setDraftUserFilter('')
    setDraftCategoryFilter('')
    setDraftNameFilter('')
    setDraftDateFrom('')
    setDraftDateTo('')
    setDraftSortOrder('desc')
  }

  function toggleFilters() {
    if (filtersOpen) {
      closeWithApply()
      return
    }
    syncDraftFromApplied()
    setFiltersOpen(true)
  }

  const fields = (
    <div className="grid gap-3">
      <FieldLabel label="User">
        <SearchableSelect
          value={draftUserFilter}
          onChange={setDraftUserFilter}
          options={userOptions}
          placeholder="All users"
          emptyOptionLabel="All users"
          searchPlaceholder="Search users"
          emptyMessage="No users found"
        />
      </FieldLabel>

      <FieldLabel label="Type">
        <SelectInput
          value={draftFilter}
          onChange={(e) => {
            const nextFilter = e.target.value as TransactionFilterValue
            setDraftFilter(nextFilter)
            const selectedDraftCategory = categories.find((category) => category.id === draftCategoryFilter)
            if (selectedDraftCategory && nextFilter !== 'all' && selectedDraftCategory.type !== nextFilter) setDraftCategoryFilter('')
          }}
          icon={<IconChevronDown className="h-4 w-4" />}
        >
          {Object.entries(filterLabels).map(([value, label]) => (
            <option key={value} value={value}>
              {label}
            </option>
          ))}
        </SelectInput>
      </FieldLabel>

      <div className="flex flex-col gap-1.5 text-[11.5px] font-bold uppercase tracking-wide text-muted">
        Category
        <SearchableSelect
          value={draftCategoryFilter}
          onChange={setDraftCategoryFilter}
          options={categoryOptions}
          placeholder="All categories"
          emptyOptionLabel="All categories"
          searchPlaceholder="Search categories"
          emptyMessage="No categories found"
        />
      </div>

      <FieldLabel label="Name">
        <TextInput
          type="search"
          value={draftNameFilter}
          onChange={(e) => setDraftNameFilter(e.target.value)}
          placeholder="Search name or category"
          className="normal-case tracking-normal"
        />
      </FieldLabel>

      <div className={compact ? 'grid min-w-0 gap-3' : 'grid min-w-0 grid-cols-2 gap-3'}>
        <DateInput label="From" value={draftDateFrom} max={draftDateTo || undefined} invalid={hasInvalidDateRange} onChange={setDraftDateFrom} />
        <DateInput label="To" value={draftDateTo} min={draftDateFrom || undefined} invalid={hasInvalidDateRange} onChange={setDraftDateTo} />
      </div>
      {hasInvalidDateRange && <p className="m-0 text-[11.5px] text-critical">From date must be before To date.</p>}

      <FieldLabel label="Order">
        <SelectInput
          value={draftSortOrder}
          onChange={(e) => setDraftSortOrder(e.target.value as TransactionSortOrder)}
          icon={<IconChevronDown className="h-4 w-4" />}
        >
          {Object.entries(sortOrderLabels).map(([value, label]) => (
            <option key={value} value={value}>
              {label}
            </option>
          ))}
        </SelectInput>
      </FieldLabel>
    </div>
  )

  const actions = (
    <div className="grid grid-cols-2 gap-2">
      <Button variant="secondary" disabled={!draftHasChanges} onClick={clearFilters} className="justify-center">
        Reset
      </Button>
      <Button variant="primary" disabled={hasInvalidDateRange} onClick={closeWithApply} className="justify-center">
        Apply
      </Button>
    </div>
  )

  return (
    <div ref={rootRef} className="relative">
      <FilterButton compact={compact} activeCount={activeFilters} open={filtersOpen} onClick={toggleFilters} />

      {compact ? (
        <BottomSheet
          open={filtersOpen}
          onClose={closeWithApply}
          title="Filter transactions"
          description="Filter and sort the list by user, type, category, date, name, or amount."
          footer={actions}
        >
          {fields}
        </BottomSheet>
      ) : (
        filtersOpen && (
          <div
            className={`absolute top-[calc(100%+10px)] z-10 rounded-3xl border border-line bg-surface p-4 shadow-[var(--shadow-md)] ${
              align === 'right' ? 'right-0 w-[460px]' : 'left-0 w-[460px]'
            }`}
          >
            <div className="mb-3">
              <p className="m-0 text-sm font-bold text-ink">Filter transactions</p>
              <p className="m-0 text-xs text-muted">Filter and sort the list by user, type, category, date, name, or amount.</p>
            </div>
            {fields}
            <div className="mt-4">{actions}</div>
          </div>
        )
      )}
    </div>
  )
}
