import type { SubTransactionDto } from '@/api/types'
import { IconChevronDown, IconList, IconPlus, IconTrash } from '@/components/icons'
import { Button } from '@/components/ui/Button'
import { IconButton } from '@/components/ui/IconButton'
import { cx } from '@/components/ui/classes'
import { centsToAmountDisplay } from '@/lib/amount'

export interface SubTransactionsPanelProps {
  items: SubTransactionDto[]
  /** The reader's own total in cents, used as a checksum. */
  attachmentTotal?: number
  expanded: boolean
  /** Extracted rows scored below the confidence threshold, outlined for review. */
  lowConfidenceIds?: ReadonlySet<string>
  disabled?: boolean
  onToggle: () => void
  onAdd: () => void
  onEdit: (id: string) => void
  onRemove: (id: string) => void
}

/**
 * Collapsible line-item panel. Collapsed, it is a one-line summary (count, total and an attachment-
 * total mismatch warning); expanding reveals the rows. Each row opens the focused edit modal, and
 * its trash button removes it. Keeping the default collapsed means a scanned attachment can fill a
 * form without pushing every line item into the main layout.
 */
export function SubTransactionsPanel({
  items,
  attachmentTotal,
  expanded,
  lowConfidenceIds,
  disabled = false,
  onToggle,
  onAdd,
  onEdit,
  onRemove,
}: SubTransactionsPanelProps) {
  const totalCents = items.reduce((sum, item) => sum + item.amount, 0)
  const attachmentTotalCents = attachmentTotal
  const differenceCents =
    attachmentTotalCents === undefined || items.length === 0 ? 0 : attachmentTotalCents - totalCents
  const hasMismatch = differenceCents !== 0
  const count = items.length

  return (
    <section className="flex flex-col rounded-2xl border border-line bg-surface" aria-label="Sub-transactions">
      <div className="flex items-center gap-2 p-3 pr-4">
        <button
          type="button"
          onClick={onToggle}
          aria-expanded={expanded}
          aria-label={expanded ? 'Collapse sub-transactions' : 'Expand sub-transactions'}
          className="flex min-w-0 flex-1 cursor-pointer items-start gap-3 rounded-xl px-1 py-1 text-left"
        >
          <span className="flex h-9 w-9 flex-none items-center justify-center rounded-xl bg-surface-sunken text-muted">
            <IconList className="h-4 w-4" />
          </span>
          <span className="flex min-w-0 flex-1 flex-col gap-0.5">
            <span className="text-sm font-bold text-ink">Sub-transactions</span>
            <span className={cx('text-xs', hasMismatch ? 'text-critical' : 'text-muted')}>
              {count === 0 ? 'No sub-transactions' : `${count} ${count === 1 ? 'item' : 'items'} · ${displayTotal(totalCents)}`}
            </span>
            {hasMismatch && (
              <span className="text-[11px] text-critical">
                Attachment total {centsToAmountDisplay(attachmentTotalCents!)} · difference{' '}
                {centsToAmountDisplay(Math.abs(differenceCents))}
              </span>
            )}
          </span>
          <IconChevronDown
            className={cx(
              'mt-1 h-4 w-4 flex-none text-muted transition-transform',
              expanded && 'rotate-180',
            )}
          />
        </button>

        {count === 0 && !expanded && (
          <Button
            variant="secondary"
            onClick={onAdd}
            disabled={disabled}
            aria-label="Add sub-transaction"
            className="inline-flex h-8 flex-none items-center justify-center px-3 py-0 text-xs"
          >
            Add
          </Button>
        )}
      </div>

      {expanded && (
        <div className="flex flex-col gap-2 border-t border-line px-3 pb-3 pt-2">
          {count === 0 ? (
            <p className="m-0 px-1 py-2 text-xs text-muted">No items yet. Add one to split this transaction.</p>
          ) : (
            <ul className="m-0 flex list-none flex-col p-0">
              {items.map((item) => {
                const lowConfidence = lowConfidenceIds?.has(item.id) ?? false
                return (
                <li key={item.id} className="flex items-center gap-1 border-t border-line/60 first:border-t-0">
                  <button
                    type="button"
                    onClick={() => onEdit(item.id)}
                    disabled={disabled}
                    aria-label={`Edit ${item.name}`}
                    aria-invalid={lowConfidence || undefined}
                    className={cx(
                      'flex min-w-0 flex-1 cursor-pointer items-center gap-3 rounded-lg px-1 py-2.5 text-left disabled:cursor-default',
                      lowConfidence && 'ring-2 ring-critical',
                    )}
                  >
                    <span className="flex min-w-0 flex-1 flex-col">
                      <span className="truncate text-[13px] font-semibold text-ink">{item.name}</span>
                      {item.label && <span className="truncate text-[11px] text-muted">{item.label}</span>}
                    </span>
                    <span className="flex-none font-mono text-[12.5px] tabular-nums text-ink-soft">
                      {centsToAmountDisplay(item.amount)}
                    </span>
                  </button>
                  <IconButton
                    aria-label={`Remove ${item.name}`}
                    onClick={() => onRemove(item.id)}
                    disabled={disabled}
                    icon={<IconTrash className="h-3.5 w-3.5" />}
                    className="h-8 w-8 flex-none text-critical"
                  />
                </li>
                )
              })}
            </ul>
          )}

          <div className="flex items-baseline justify-between gap-3 border-t border-line px-1 pt-2">
            <span className="text-[11px] font-bold uppercase tracking-wide text-muted">Total</span>
            <span className="font-mono text-[13px] font-semibold tabular-nums text-ink">
              {displayTotal(totalCents)}
            </span>
          </div>

          <Button
            variant="secondary"
            onClick={onAdd}
            disabled={disabled}
            aria-label="Add sub-transaction"
            className="flex items-center justify-center gap-1.5 py-2.5"
          >
            <IconPlus className="h-3.5 w-3.5" />
            Add item
          </Button>
        </div>
      )}
    </section>
  )
}

function displayTotal(cents: number): string {
  return centsToAmountDisplay(cents)
}
