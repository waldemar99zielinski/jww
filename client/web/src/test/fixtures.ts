import type {
  CurrencyDto,
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
  original_filename: 'receipt.jpg',
  media_type: 'image/jpeg',
  byte_size: 7,
  created_at: '2026-10-10T09:00:00Z',
}

export const fixturePrefill: TransactionPrefillDto = {
  attachment_id: fixtureAttachment.id,
  type: { value: 'expense', confidence: 0.9 },
  amount: { value: '24,90', confidence: 0.95, source: 'total line' },
  name: { value: 'Lidl groceries', confidence: 0.8, source: 'merchant header' },
  category_hint: { value: 'Groceries', confidence: 0.85 },
  duration_ms: 900,
}

export const fixtureTransactions: TransactionDto[] = [
  {
    id: 1,
    type: 'expense',
    amount: '42.50',
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
        original_filename: 'receipt.jpg',
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
  },
  {
    id: 2,
    type: 'income',
    amount: '2500.00',
    currency_code: 'EUR',
    category_id: 'cat-salary',
    category_name: 'Salary',
    user_id: fixtureUser.id,
    source: 'manual',
    name: 'October salary',
    created_at: '2026-10-01T08:00:00Z',
    description: '',
    attachments: [],
  },
]

export const fixtureSummary: SummaryDto = {
  period: 'month',
  period_start: '2026-10-01T00:00:00Z',
  period_end: '2026-10-31T23:59:59Z',
  currency_code: 'EUR',
  total_income: '2500.00',
  total_expense: '42.50',
  net: '2457.50',
  by_currency: {
    EUR: { currency_code: 'EUR', total_income: '2500.00', total_expense: '42.50', net: '2457.50' },
  },
  transaction_count: 2,
}

export const fixtureTokenResponse: TokenResponse = {
  access_token: 'fixture-access-token',
  token_type: 'Bearer',
}
