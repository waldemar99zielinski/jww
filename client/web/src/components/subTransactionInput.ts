/** One item's editable fields. `amount` is whole minor units (cents), matching the create API. */
export interface SubTransactionInput {
  name: string
  label: string
  amount: number
}
