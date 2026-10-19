import { useState } from 'react'
import type { SubTransactionDto } from '@/api/types'
import { IconEdit, IconPlus, IconSpinner, IconTrash } from '@/components/icons'
import { Button } from '@/components/ui/Button'
import { IconButton } from '@/components/ui/IconButton'
import { SignedAmountInput } from '@/components/ui/SignedAmountInput'
import { TextInput } from '@/components/ui/TextInput'
import { cx } from '@/components/ui/classes'
import { amountToCents, centsToAmountDisplay } from '@/lib/amount'
import { formatAmount } from '@/lib/currency'

/** One item's editable fields. `amount` is whole minor units (cents), matching the create API. */
export interface SubTransactionListInput {
  name: string
  label: string
  amount: number
}

export interface SubTransactionListProps {
  items: SubTransactionDto[]
  onAdd: (input: SubTransactionListInput) => void
  onUpdate: (id: string, input: SubTransactionListInput) => void
  onRemove: (id: string) => void
  /** Locks every control (e.g. the parent form is saving or read-only). */
  disabled?: boolean
  /** Parent is persisting a change: controls lock and the affected row acknowledges before the item list refreshes. */
  busy?: boolean
  /** Currency of the parent transaction; defaults to the app's base-format fallback. */
  currencyCode?: string
}

/**
 * Editable line items: name / label / amount per row, inline edit + remove per row, a "+ Add item"
 * row that reveals the same small form, and the running total of what the rows add up to.
 *
 * Pure display: the committed rows and the running total come from `items`, and every change goes
 * out as a callback. Only the transient draft of the row currently being typed into is local
 * state — the parent owns what is saved (local on the Add screen, a CRUD call on the detail
 * screen). Amounts follow the transaction Amount field's convention exactly: the UI edits a comma
 * display string (`lib/amount.ts`), the callbacks receive whole minor units.
 */
export function SubTransactionList({
  items,
  onAdd,
  onUpdate,
  onRemove,
  disabled = false,
  busy = false,
  currencyCode = 'PLN',
}: SubTransactionListProps) {
  const [editingId, setEditingId] = useState<string | null>(null)
  const [editName, setEditName] = useState('')
  const [editLabel, setEditLabel] = useState('')
  const [editAmount, setEditAmount] = useState('')
  const [adding, setAdding] = useState(false)
  const [addName, setAddName] = useState('')
  const [addLabel, setAddLabel] = useState('')
  const [addAmount, setAddAmount] = useState('')

  const locked = disabled || busy
  const totalCents = items.reduce((sum, item) => sum + item.amount, 0)
  const canSave = editName.trim() !== '' && amountToCents(editAmount) !== 0
  const canAdd = addName.trim() !== '' && amountToCents(addAmount) !== 0

  function startEdit(item: SubTransactionDto) {
    setEditingId(item.id)
    setEditName(item.name)
    setEditLabel(item.label)
    setEditAmount(centsToAmountDisplay(item.amount))
  }

  function cancelEdit() {
    setEditingId(null)
  }

  function saveEdit() {
    if (!canSave || editingId === null) return
    onUpdate(editingId, { name: editName.trim(), label: editLabel.trim(), amount: amountToCents(editAmount) })
    setEditingId(null)
  }

  function startAdd() {
    setAddName('')
    setAddLabel('')
    setAddAmount('')
    setAdding(true)
  }

  function cancelAdd() {
    setAdding(false)
  }

  function saveAdd() {
    if (addName.trim() === '' || amountToCents(addAmount) === 0) return
    onAdd({ name: addName.trim(), label: addLabel.trim(), amount: amountToCents(addAmount) })
    setAdding(false)
  }

  function amountField(label: string, value: string, onChange: (next: string) => void) {
    return (
      <SignedAmountInput
        value={value}
        onChange={onChange}
        label={label}
        className="bg-surface font-mono tabular-nums sm:text-[13px]"
      />
    )
  }

  return (
    <section className="flex flex-col gap-2 text-ink" aria-label="Sub-transaction items">
      <ul className="m-0 flex list-none flex-col gap-2 p-0">
        {items.map((item) => {
          const editing = editingId === item.id
          const hasLabel = item.label !== ''

          if (editing) {
            return (
              <li
                key={item.id}
                role="group"
                aria-label={`Editing ${item.name}`}
                className={cx(rowClass(), 'flex-col items-stretch gap-2 sm:flex-row sm:items-end')}
              >
                <div className="flex min-w-0 flex-1 flex-col gap-1">
                  <TextInput
                    type="text"
                    value={editName}
                    onChange={(e) => setEditName(e.target.value)}
                    placeholder="Item name"
                    aria-label={`Item name for ${item.name}`}
                    className="bg-surface sm:text-[13px]"
                    autoFocus
                  />
                  <TextInput
                    type="text"
                    value={editLabel}
                    onChange={(e) => setEditLabel(e.target.value)}
                    placeholder="Label (optional)"
                    aria-label={`Item label for ${item.name}`}
                    className="bg-surface sm:text-[13px]"
                  />
                </div>
                <div className="flex w-full flex-col gap-2 sm:w-auto sm:flex-row sm:items-center">
                  <div className="w-full sm:w-32 sm:flex-none">
                    {amountField(`Item amount for ${item.name}`, editAmount, setEditAmount)}
                  </div>
                  <div className="flex items-center gap-2">
                    <Button
                      variant="primary"
                      onClick={saveEdit}
                      disabled={!canSave || busy}
                      className="h-9 flex-1 justify-center px-3 py-0 text-xs sm:h-8 sm:flex-none"
                    >
                      {busy ? 'Saving…' : 'Save'}
                    </Button>
                    <Button
                      variant="ghost"
                      onClick={cancelEdit}
                      disabled={busy}
                      className="h-9 flex-1 justify-center px-3 py-0 text-xs sm:h-8 sm:flex-none"
                    >
                      Cancel
                    </Button>
                  </div>
                </div>
              </li>
            )
          }

          return (
            <li key={item.id} className={rowClass()}>
              <div className="flex min-w-0 flex-1 flex-col sm:min-w-0">
                <span className="truncate text-[13px] font-semibold text-ink">{item.name}</span>
                {hasLabel && <span className="truncate text-xs text-muted">{item.label}</span>}
              </div>
              <div className="flex flex-none items-center gap-2">
                <span className="font-mono text-[13px] tabular-nums text-ink">
                  {centsToAmountDisplay(item.amount)}
                </span>
                <IconButton
                  icon={<IconEdit className="h-3.5 w-3.5" />}
                  aria-label={`Edit ${item.name}`}
                  onClick={() => startEdit(item)}
                  disabled={locked}
                  className="h-8 w-8 flex-none"
                />
                <IconButton
                  icon={<IconTrash className="h-3.5 w-3.5" />}
                  aria-label={`Remove ${item.name}`}
                  onClick={() => onRemove(item.id)}
                  disabled={locked}
                  className="h-8 w-8 flex-none text-critical"
                />
              </div>
            </li>
          )
        })}
      </ul>

      {adding ? (
        <div role="group" aria-label="New item" className={cx(rowClass(), 'flex-col items-stretch gap-2 sm:flex-row sm:items-end')}>
          <div className="flex min-w-0 flex-1 flex-col gap-1">
            <TextInput
              type="text"
              value={addName}
              onChange={(e) => setAddName(e.target.value)}
              placeholder="Item name"
              aria-label="New item name"
              className="bg-surface sm:text-[13px]"
              autoFocus
            />
            <TextInput
              type="text"
              value={addLabel}
              onChange={(e) => setAddLabel(e.target.value)}
              placeholder="Label (optional)"
              aria-label="New item label"
              className="bg-surface sm:text-[13px]"
            />
          </div>
          <div className="flex w-full flex-col gap-2 sm:w-auto sm:flex-row sm:items-center">
            <div className="w-full sm:w-32 sm:flex-none">
              {amountField('New item amount', addAmount, setAddAmount)}
            </div>
            <div className="flex items-center gap-2">
              <Button
                variant="primary"
                onClick={saveAdd}
                disabled={!canAdd || disabled}
                className="h-9 flex-1 justify-center px-3 py-0 text-xs sm:h-8 sm:flex-none"
              >
                Add item
              </Button>
              <Button
                variant="ghost"
                onClick={cancelAdd}
                className="h-9 flex-1 justify-center px-3 py-0 text-xs sm:h-8 sm:flex-none"
              >
                Cancel
              </Button>
            </div>
          </div>
        </div>
      ) : (
        <Button
          variant="secondary"
          onClick={startAdd}
          disabled={locked}
          className="flex items-center justify-center gap-1.5 py-2.5"
        >
          {busy ? <IconSpinner className="h-3.5 w-3.5 animate-spin" /> : <IconPlus className="h-3.5 w-3.5" />}
          Add item
        </Button>
      )}

      <div className="flex items-baseline justify-between gap-3 px-1 pt-1 text-[11.5px] font-bold uppercase tracking-wide text-muted">
        <span>Total</span>
        <span className="font-mono text-[13px] normal-case tabular-nums text-ink" aria-live="polite">
          {formatAmount(totalCents / 100, currencyCode)}
        </span>
      </div>
    </section>
  )
}

/** Two-line stacked (name+label / amount+controls) under the mobile breakpoint, one row at `sm:`. */
function rowClass() {
  return 'flex flex-col gap-2 rounded-xl border border-line bg-surface-sunken px-3.5 py-2.5 sm:flex-row sm:items-center sm:gap-3'
}
