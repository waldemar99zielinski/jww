import type {
  CurrencyDto,
  SubTransactionDto,
  SummaryDto,
  TransactionAttachmentDto,
  TransactionCategoryDto,
  TransactionDto,
  TransactionPrefillDto,
  UserDto,
} from '@/api/types'
import type { TokenResponse } from '@/api/auth'

export const fixtureUser: UserDto = {
  id: 'user-1',
  username: 'waldemar',
  display_name: 'Waldemar',
  created_at: '2026-01-01T00:00:00Z',
}

export const fixtureUsers: UserDto[] = [
  fixtureUser,
  { id: 'user-2', username: 'partner', display_name: 'Partner', created_at: '2026-01-01T00:00:00Z' },
]

export const fixtureCurrencies: CurrencyDto[] = [
  { code: 'EUR', name: 'Euro', is_base: true },
  { code: 'USD', name: 'US Dollar', is_base: false },
]

export const fixtureExpenseCategories: TransactionCategoryDto[] = [
  { id: 'cat-groceries', type: 'expense', name: 'Groceries', active: true, sort_order: 0 },
  { id: 'cat-transport', type: 'expense', name: 'Transport', active: true, sort_order: 1 },
  { id: 'cat-rent', type: 'expense', name: 'Rent', active: true, sort_order: 4 },
]

export const fixtureIncomeCategories: TransactionCategoryDto[] = [
  { id: 'cat-salary', type: 'income', name: 'Salary', active: true, sort_order: 0 },
  { id: 'cat-freelance', type: 'income', name: 'Freelance', active: true, sort_order: 1 },
]

export const fixtureCategories: TransactionCategoryDto[] = [...fixtureExpenseCategories, ...fixtureIncomeCategories]

export const fixtureAttachment: TransactionAttachmentDto = {
  id: 'attachment-1',
  transaction_id: null,
  original_filename: 'attachment.jpg',
  media_type: 'image/jpeg',
  byte_size: 7,
  created_at: '2026-10-10T09:00:00Z',
}

export const fixturePrefill: TransactionPrefillDto = {
  attachment_id: fixtureAttachment.id,
  attachment_ids: [fixtureAttachment.id],
  type: { value: 'expense', confidence: 90 },
  amount: { value: 2490, confidence: 95, source: 'total line' },
  name: { value: 'Lidl groceries', confidence: 80, source: 'merchant header' },
  category_hint: { value: 'Groceries', confidence: 85 },
  occurred_at: { value: '14:35', confidence: 70, source: 'footer' },
  duration_ms: 900,
}

/** One persisted line item under a transaction; `amount` is whole cents, like `TransactionDto.amount`. */
export const fixtureSubTransaction: SubTransactionDto = {
  id: 'sub-transaction-1',
  name: 'Milk 2%',
  label: 'dairy',
  amount: 649,
}

/** Two items, so the derived Σ(amount) is distinguishable from any single row. */
export const fixtureSubTransactions: SubTransactionDto[] = [
  fixtureSubTransaction,
  { id: 'sub-transaction-2', name: 'Sourdough bread', label: 'bakery', amount: 599 },
]

/** Prefill variant proposing line items, for the auto-fill extraction flow. */
export const fixturePrefillWithSubTransactions: TransactionPrefillDto = {
  ...fixturePrefill,
  sub_transactions: fixtureSubTransactions.map(({ name, label, amount }, index) => ({
    name,
    label,
    amount,
    confidence: 90 - index * 5,
  })),
}

/**
 * A transaction whose amount is the derived Σ of its items — mirrors the server's rule that a
 * transaction with line items always reports their sum (and its amount is then read-only).
 * Deliberately standalone, *not* part of `fixtureTransactions`: the list fixtures are asserted
 * as-is by the Home/transactions tests, and a second itemised expense in that list would shift
 * the derived category breakdown. Tests and handlers that need it reference it explicitly.
 */
export const fixtureItemisedTransaction: TransactionDto = {
  id: 3,
  type: 'expense',
  amount: 1248,
  currency_code: 'EUR',
  category_id: 'cat-groceries',
  category_name: 'Groceries',
  user_id: fixtureUser.id,
  source: 'manual',
  name: 'Itemised shop',
  created_at: '2026-10-09T09:30:00Z',
  description: 'Attachment with line items',
  attachments: [],
  sub_transactions: fixtureSubTransactions,
}

export const fixtureTransactions: TransactionDto[] = [
  {
    id: 1,
    type: 'expense',
    amount: 4250,
    currency_code: 'EUR',
    category_id: 'cat-groceries',
    category_name: 'Groceries',
    user_id: fixtureUser.id,
    source: 'manual',
    name: 'Weekly shop',
    created_at: '2026-10-10T09:30:00Z',
    description: 'Milk, bread and coffee',
    attachments: [
      {
        id: 'attachment-1',
        transaction_id: 1,
        original_filename: 'attachment.jpg',
        media_type: 'image/jpeg',
        byte_size: 20480,
        created_at: '2026-10-10T09:31:00Z',
      },
      {
        id: 'attachment-2',
        transaction_id: 1,
        original_filename: 'note.txt',
        media_type: 'text/plain',
        byte_size: 512,
        created_at: '2026-10-10T09:32:00Z',
      },
    ],
    sub_transactions: [],
  },
  {
    id: 2,
    type: 'income',
    amount: 250000,
    currency_code: 'EUR',
    category_id: 'cat-salary',
    category_name: 'Salary',
    user_id: fixtureUser.id,
    source: 'manual',
    name: 'October salary',
    created_at: '2026-10-01T08:00:00Z',
    description: '',
    attachments: [],
    sub_transactions: [],
  },
]

export const fixtureSummary: SummaryDto = {
  period: 'month',
  period_start: '2026-10-01T00:00:00Z',
  period_end: '2026-10-31T23:59:59Z',
  currency_code: 'EUR',
  total_income: 250000,
  total_expense: 4250,
  net: 245750,
  by_currency: {
    EUR: { currency_code: 'EUR', total_income: 250000, total_expense: 4250, net: 245750 },
  },
  transaction_count: 2,
}

export const fixtureTokenResponse: TokenResponse = {
  access_token: 'fixture-access-token',
  token_type: 'Bearer',
}
