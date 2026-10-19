import { useState } from 'react'
import type { SubTransactionDto } from '@/api/types'
import type { SubTransactionListInput } from '@/components/SubTransactionList'
import { FieldLabel } from '@/components/ui/FieldLabel'
import { Modal } from '@/components/ui/Modal'
import { SignedAmountInput } from '@/components/ui/SignedAmountInput'
import { TextInput } from '@/components/ui/TextInput'
import { Button } from '@/components/ui/Button'
import { amountToCents, centsToAmountDisplay } from '@/lib/amount'

export interface SubTransactionModalProps {
  open: boolean
  /** The row being edited; omit to create a new one. */
  item?: SubTransactionDto
  disabled?: boolean
  onClose: () => void
  onSave: (input: SubTransactionListInput) => void
}

/**
 * Focused editor for one line item. Mounted with a key per row (or `new`) by the Add container, so
 * the initial values are seeded once and switching rows starts clean. Saving writes back to the
 * page's local draft; the transaction itself is only persisted from the main Save button.
 */
export function SubTransactionModal({ open, item, disabled = false, onClose, onSave }: SubTransactionModalProps) {
  const [name, setName] = useState(item?.name ?? '')
  const [label, setLabel] = useState(item?.label ?? '')
  const [amount, setAmount] = useState(item ? centsToAmountDisplay(item.amount) : '')

  const amountCents = amountToCents(amount)
  const canSave = name.trim() !== '' && amountCents !== 0

  function handleSave() {
    if (!canSave) return
    onSave({ name: name.trim(), label: label.trim(), amount: amountCents })
  }

  return (
    <Modal
      open={open}
      onClose={onClose}
      title={item ? 'Edit sub-transaction' : 'Add sub-transaction'}
      description="The transaction amount is the sum of its items."
      footer={
        <div className="flex justify-end gap-2">
          <Button variant="ghost" onClick={onClose}>
            Cancel
          </Button>
          <Button variant="primary" onClick={handleSave} disabled={!canSave || disabled}>
            {item ? 'Save' : 'Add item'}
          </Button>
        </div>
      }
    >
      <div className="flex flex-col gap-3">
        <FieldLabel label="Name">
          <TextInput
            type="text"
            value={name}
            onChange={(event) => setName(event.target.value)}
            placeholder="Item name"
            aria-label="Sub-transaction name"
            autoFocus
          />
        </FieldLabel>
        <FieldLabel label="Label (optional)">
          <TextInput
            type="text"
            value={label}
            onChange={(event) => setLabel(event.target.value)}
            placeholder="e.g. dairy"
            aria-label="Sub-transaction label"
          />
        </FieldLabel>
        <FieldLabel label="Amount">
          <SignedAmountInput
            value={amount}
            onChange={setAmount}
            placeholder="0,00 (negative for discounts)"
            label="Sub-transaction amount"
            className="font-mono tabular-nums"
          />
        </FieldLabel>
      </div>
    </Modal>
  )
}
