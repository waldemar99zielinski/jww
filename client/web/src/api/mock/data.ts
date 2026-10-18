import type {
  CurrencyDto,
  SummaryDto,
  SummaryPeriod,
  TransactionAttachmentDto,
  TransactionCategoryDto,
  TransactionDetailsDto,
  TransactionDto,
  TransactionPrefillDto,
  TransactionSource,
  TransactionType,
  UserDto,
} from '@/api/types'

export const MOCK_USER: UserDto = {
  id: 'design-user',
  username: 'design',
  display_name: 'Design Reviewer',
  created_at: '2025-01-01T00:00:00.000Z',
}

const MOCK_PARTNER: UserDto = {
  id: 'design-partner',
  username: 'partner',
  display_name: 'Design Partner',
  created_at: '2025-01-02T00:00:00.000Z',
}

/** Small artificial delay so loading/pending states are actually visible while reviewing design. */
export function mockDelay<T>(value: T, ms = 300): Promise<T> {
  return new Promise((resolve) => setTimeout(() => resolve(value), ms))
}

function daysAgo(days: number, hour: number): string {
  const date = new Date()
  date.setDate(date.getDate() - days)
  date.setHours(hour, 0, 0, 0)
  return date.toISOString()
}

const currencies: CurrencyDto[] = [
  { code: 'PLN', name: 'Polish złoty', is_base: true },
  { code: 'USD', name: 'US dollar', is_base: false },
  { code: 'EUR', name: 'Euro', is_base: false },
]

const categories: TransactionCategoryDto[] = [
  { id: 'cat-groceries', type: 'expense', name: 'Groceries', active: true, sort_order: 0 },
  { id: 'cat-transport', type: 'expense', name: 'Transport', active: true, sort_order: 1 },
  { id: 'cat-entertainment', type: 'expense', name: 'Entertainment', active: true, sort_order: 2 },
  { id: 'cat-health', type: 'expense', name: 'Health', active: true, sort_order: 3 },
  { id: 'cat-rent', type: 'expense', name: 'Rent', active: true, sort_order: 4 },
  { id: 'cat-shopping', type: 'expense', name: 'Shopping', active: true, sort_order: 5 },
  { id: 'cat-subscriptions', type: 'expense', name: 'Subscriptions', active: true, sort_order: 6 },
  { id: 'cat-travel', type: 'expense', name: 'Travel', active: true, sort_order: 7 },
  { id: 'cat-utilities', type: 'expense', name: 'Utilities', active: true, sort_order: 8 },
  { id: 'cat-education', type: 'expense', name: 'Education', active: true, sort_order: 9 },
  { id: 'cat-personal', type: 'expense', name: 'Personal care', active: true, sort_order: 10 },
  { id: 'cat-gifts', type: 'expense', name: 'Gifts', active: true, sort_order: 11 },
  { id: 'cat-pets', type: 'expense', name: 'Pets', active: true, sort_order: 12 },
  { id: 'cat-dining', type: 'expense', name: 'Dining out', active: true, sort_order: 13 },
  { id: 'cat-salary', type: 'income', name: 'Salary', active: true, sort_order: 0 },
  { id: 'cat-freelance', type: 'income', name: 'Freelance', active: true, sort_order: 1 },
  { id: 'cat-investments', type: 'income', name: 'Investments', active: true, sort_order: 2 },
  { id: 'cat-refunds', type: 'income', name: 'Refunds', active: true, sort_order: 3 },
  { id: 'cat-bonus', type: 'income', name: 'Bonus', active: true, sort_order: 4 },
  { id: 'cat-income-gifts', type: 'income', name: 'Gifts', active: true, sort_order: 5 },
]

let nextTransactionId = 1000
let nextAttachmentId = 1

/** Unattached drafts live here until a transaction links them (transaction_id stays null). */
const attachments: TransactionAttachmentDto[] = []

function transaction(input: {
  type: TransactionType
  amount: string
  categoryId: string
  daysAgo: number
  hour: number
  description: string
  name?: string
  userId?: string
  currencyCode?: string
  source?: TransactionSource
}): TransactionDto {
  const category = categories.find((c) => c.id === input.categoryId)
  return {
    id: nextTransactionId++,
    type: input.type,
    amount: input.amount,
    currency_code: input.currencyCode ?? 'PLN',
    category_id: input.categoryId,
    category_name: category?.name ?? 'Other',
    user_id: input.userId ?? MOCK_USER.id,
    source: input.source ?? 'manual',
    name: input.name ?? input.description,
    created_at: daysAgo(input.daysAgo, input.hour),
    description: input.description,
    attachments: [],
  }
}

const transactions: TransactionDto[] = [
  // Today
  transaction({ type: 'expense', amount: '32.00', categoryId: 'cat-transport', daysAgo: 0, hour: 8, description: 'Morning tram' }),
  transaction({ type: 'expense', amount: '15.99', categoryId: 'cat-entertainment', daysAgo: 0, hour: 13, description: 'Movie rental', source: 'manual' }),
  transaction({ type: 'expense', amount: '184.32', categoryId: 'cat-groceries', daysAgo: 0, hour: 18, description: 'Weekly shopping' }),
  // Yesterday
  transaction({ type: 'expense', amount: '45.00', categoryId: 'cat-transport', daysAgo: 1, hour: 8, description: 'Fuel' }),
  transaction({ type: 'income', amount: '320.00', categoryId: 'cat-freelance', daysAgo: 1, hour: 12, description: 'Logo design', source: 'manual' }),
  transaction({ type: 'expense', amount: '78.50', categoryId: 'cat-groceries', daysAgo: 1, hour: 19, description: 'Corner store' }),
  // 2 days ago
  transaction({ type: 'expense', amount: '12.00', categoryId: 'cat-transport', daysAgo: 2, hour: 7, description: 'Bus ticket', userId: MOCK_PARTNER.id }),
  transaction({ type: 'expense', amount: '24.30', categoryId: 'cat-groceries', daysAgo: 2, hour: 9, description: 'Bakery' }),
  transaction({ type: 'expense', amount: '59.99', categoryId: 'cat-entertainment', daysAgo: 2, hour: 20, description: 'Streaming subscription', source: 'recurring' }),
  // 3 days ago
  transaction({ type: 'income', amount: '6500.00', categoryId: 'cat-salary', daysAgo: 3, hour: 9, description: 'Monthly salary', source: 'recurring' }),
  transaction({ type: 'expense', amount: '99.00', categoryId: 'cat-health', daysAgo: 3, hour: 11, description: 'Gym membership', source: 'recurring' }),
  transaction({ type: 'expense', amount: '210.75', categoryId: 'cat-groceries', daysAgo: 3, hour: 18, description: 'Supermarket' }),
  // 4 days ago
  transaction({ type: 'expense', amount: '120.50', categoryId: 'cat-health', daysAgo: 4, hour: 14, description: 'Pharmacy' }),
  transaction({ type: 'expense', amount: '54.20', categoryId: 'cat-entertainment', daysAgo: 4, hour: 21, description: 'Concert tickets', userId: MOCK_PARTNER.id, source: 'manual' }),
  // 5 days ago
  transaction({ type: 'expense', amount: '1800.00', categoryId: 'cat-rent', daysAgo: 5, hour: 10, description: 'Monthly rent', source: 'recurring' }),
  transaction({ type: 'expense', amount: '38.90', categoryId: 'cat-transport', daysAgo: 5, hour: 17, description: 'Taxi home', source: 'manual' }),
  // 6 days ago
  transaction({ type: 'expense', amount: '145.00', categoryId: 'cat-groceries', daysAgo: 6, hour: 12, description: 'Costco run' }),
  transaction({ type: 'income', amount: '950.00', categoryId: 'cat-freelance', daysAgo: 6, hour: 16, description: 'Freelance project' }),
  transaction({ type: 'expense', amount: '29.99', categoryId: 'cat-entertainment', daysAgo: 6, hour: 22, description: 'Music streaming', userId: MOCK_PARTNER.id, source: 'recurring' }),
  // 7 days ago
  transaction({ type: 'expense', amount: '18.75', categoryId: 'cat-transport', daysAgo: 7, hour: 9, description: 'Parking' }),
  transaction({ type: 'expense', amount: '62.40', categoryId: 'cat-health', daysAgo: 7, hour: 13, description: 'Dentist visit' }),
  // 8 days ago
  transaction({ type: 'income', amount: '2400.00', categoryId: 'cat-salary', daysAgo: 8, hour: 9, description: 'Salary bonus', userId: MOCK_PARTNER.id }),
  transaction({ type: 'expense', amount: '305.60', categoryId: 'cat-groceries', daysAgo: 8, hour: 18, description: 'Monthly stock-up' }),
  // 10 days ago
  transaction({ type: 'expense', amount: '42.10', categoryId: 'cat-transport', daysAgo: 10, hour: 8, description: 'Monthly transit pass', source: 'recurring' }),
  transaction({ type: 'expense', amount: '75.00', categoryId: 'cat-entertainment', daysAgo: 10, hour: 20, description: 'Board game night' }),
  // 12 days ago
  transaction({ type: 'income', amount: '120.00', categoryId: 'cat-freelance', daysAgo: 12, hour: 11, description: 'Consulting call', source: 'manual' }),
  transaction({ type: 'expense', amount: '58.30', categoryId: 'cat-health', daysAgo: 12, hour: 15, description: 'Pharmacy' }),
]

/** Mirrors the server's previewable media types (transaction_attachment.previewableMediaType). */
const PREVIEWABLE_MEDIA_TYPES = new Set([
  'image/jpeg',
  'image/png',
  'image/webp',
  'image/gif',
  'image/avif',
  'image/heic',
  'image/heif',
  'image/dng',
  'image/x-adobe-dng',
  'application/pdf',
])

// A few demo attachments on the "Weekly shopping" transaction so the detail screen's
// preview/download affordances are visible in design mode (the server returns these from
// GET /transactions/{id}, alongside a preview grant for the previewable ones).
const demoTransaction = transactions[2]
demoTransaction.attachments = [
  {
    id: 'design-attachment-1',
    transaction_id: demoTransaction.id,
    original_filename: 'receipt.jpg',
    media_type: 'image/jpeg',
    byte_size: 248_000,
    created_at: demoTransaction.created_at,
  },
  {
    id: 'design-attachment-2',
    transaction_id: demoTransaction.id,
    original_filename: 'statement.pdf',
    media_type: 'application/pdf',
    byte_size: 132_000,
    created_at: demoTransaction.created_at,
  },
  {
    id: 'design-attachment-3',
    transaction_id: demoTransaction.id,
    original_filename: 'invoice.csv',
    media_type: 'text/csv',
    byte_size: 1_200,
    created_at: demoTransaction.created_at,
  },
]

/** Attaches a design-mode preview grant to each previewable attachment, mirroring the server's GET detail response. */
function withPreviewGrants(transaction: TransactionDto): TransactionDetailsDto {
  return {
    ...transaction,
    attachments: transaction.attachments.map((attachment) => ({
      ...attachment,
      preview_grant: PREVIEWABLE_MEDIA_TYPES.has(attachment.media_type) ? `design-preview-grant-${attachment.id}` : null,
    })),
  }
}

let nextCategorySeq = 100

export const mockApi = {
  listCurrencies(): CurrencyDto[] {
    return currencies
  },
  listUsers(): UserDto[] {
    return [MOCK_USER, MOCK_PARTNER]
  },
  createCurrency(code: string, name: string, isBase: boolean): CurrencyDto {
    if (isBase) currencies.forEach((c) => (c.is_base = false))
    const currency: CurrencyDto = { code, name, is_base: isBase }
    currencies.push(currency)
    return currency
  },
  updateCurrency(code: string, input: { name: string; is_base: boolean }): CurrencyDto {
    const currency = currencies.find((c) => c.code === code)
    if (!currency) throw new Error(`Unknown currency: ${code}`)
    if (input.is_base) currencies.forEach((c) => (c.is_base = false))
    currency.name = input.name
    currency.is_base = input.is_base
    return currency
  },

  listCategories(type: TransactionType, activeOnly: boolean): TransactionCategoryDto[] {
    return categories.filter((c) => c.type === type && (!activeOnly || c.active))
  },
  createCategory(type: TransactionType, name: string, sortOrder: number): TransactionCategoryDto {
    const category: TransactionCategoryDto = { id: `cat-${nextCategorySeq++}`, type, name, active: true, sort_order: sortOrder }
    categories.push(category)
    return category
  },
  updateCategory(id: string, input: { name: string; active: boolean; sort_order: number }): TransactionCategoryDto {
    const category = categories.find((c) => c.id === id)
    if (!category) throw new Error(`Unknown category: ${id}`)
    Object.assign(category, input)
    return category
  },

  listTransactions(filters: {
    user_id?: string
    category_id?: string
    type?: TransactionType
    currency_code?: string
    source?: TransactionSource
    created_from?: string
    created_to?: string
  }): TransactionDto[] {
    return transactions
      .filter((t) => {
        if (filters.user_id && t.user_id !== filters.user_id) return false
        if (filters.category_id && t.category_id !== filters.category_id) return false
        if (filters.type && t.type !== filters.type) return false
        if (filters.currency_code && t.currency_code !== filters.currency_code) return false
        if (filters.source && t.source !== filters.source) return false
        if (filters.created_from && t.created_at < filters.created_from) return false
        if (filters.created_to && t.created_at > filters.created_to) return false
        return true
      })
      .sort((a, b) => b.id - a.id)
  },
  getTransaction(id: number): TransactionDetailsDto {
    const transaction = transactions.find((t) => t.id === id)
    if (!transaction) throw new Error(`Unknown transaction: ${id}`)
    return withPreviewGrants(transaction)
  },
  updateTransaction(
    id: number,
    input: {
      type?: TransactionType
      /** Whole minor units (cents); mirrors the DTO's decimal string. */
      amount?: number
      currency_code?: string
      category_id?: string
      user_id?: string
      source?: TransactionSource
      name?: string
      description?: string
      created_at?: string
    },
  ): TransactionDetailsDto {
    const transaction = transactions.find((t) => t.id === id)
    if (!transaction) throw new Error(`Unknown transaction: ${id}`)

    if (input.type !== undefined) transaction.type = input.type
    if (input.amount !== undefined) transaction.amount = (input.amount / 100).toFixed(2)
    if (input.currency_code !== undefined) transaction.currency_code = input.currency_code
    if (input.category_id !== undefined) {
      transaction.category_id = input.category_id
      transaction.category_name = categories.find((c) => c.id === input.category_id)?.name ?? 'Other'
    }
    if (input.user_id !== undefined) transaction.user_id = input.user_id
    if (input.source !== undefined) transaction.source = input.source
    if (input.name !== undefined) transaction.name = input.name
    if (input.description !== undefined) transaction.description = input.description
    if (input.created_at !== undefined) transaction.created_at = input.created_at

    return withPreviewGrants(transaction)
  },
  createTransaction(input: {
    type: TransactionType
    /** Whole minor units (cents); the DTO it returns carries the decimal string. */
    amount: number
    currency_code: string
    category_id: string
    user_id: string
    source: TransactionSource
    name: string
    description?: string
    created_at?: string
    attachment_ids?: string[]
  }): TransactionDto {
    const category = categories.find((c) => c.id === input.category_id)
    const transaction: TransactionDto = {
      id: nextTransactionId++,
      type: input.type,
      // The wire sends cents (4250); the DTO mirrors the server's decimal string ("42.50").
      amount: (input.amount / 100).toFixed(2),
      currency_code: input.currency_code,
      category_id: input.category_id,
      category_name: category?.name ?? 'Other',
      user_id: input.user_id,
      source: input.source,
      name: input.name,
      created_at: input.created_at ?? new Date().toISOString(),
      description: input.description ?? '',
      attachments: [],
    }

    const draftIds = new Set(input.attachment_ids ?? [])
    transaction.attachments = attachments.filter((a) => a.transaction_id === null && draftIds.has(a.id))
    transaction.attachments.forEach((a) => (a.transaction_id = transaction.id))

    transactions.push(transaction)
    return transaction
  },

  uploadDraftAttachment(file: File): TransactionAttachmentDto {
    const attachment: TransactionAttachmentDto = {
      id: `attachment-${nextAttachmentId++}`,
      transaction_id: null,
      original_filename: file.name,
      media_type: file.type || 'application/octet-stream',
      byte_size: file.size,
      created_at: new Date().toISOString(),
    }
    attachments.push(attachment)
    return attachment
  },
  deleteDraftAttachment(attachmentId: string): void {
    const index = attachments.findIndex((a) => a.id === attachmentId && a.transaction_id === null)
    if (index !== -1) attachments.splice(index, 1)
  },
  deleteTransactionAttachment(transactionId: number, attachmentId: string): void {
    const transaction = transactions.find((t) => t.id === transactionId)
    if (transaction) transaction.attachments = transaction.attachments.filter((a) => a.id !== attachmentId)
    const index = attachments.findIndex((a) => a.id === attachmentId)
    if (index !== -1) attachments.splice(index, 1)
  },
  downloadAttachment(): Blob {
    return new Blob(['design-mode attachment bytes'], { type: 'application/octet-stream' })
  },
  extractTransactionPrefill(attachmentId: string): TransactionPrefillDto {
    return {
      attachment_id: attachmentId,
      type: { value: 'expense', confidence: 0.9 },
      amount: { value: '24,90', confidence: 0.95, source: 'total line' },
      name: { value: 'Lidl groceries', confidence: 0.8, source: 'merchant header' },
      category_hint: { value: 'Groceries', confidence: 0.85 },
      duration_ms: 1400,
    }
  },

  getSummary(params: { period_end: string; currency_code: string; period?: SummaryPeriod; period_start?: string }): SummaryDto {
    const matching = transactions.filter(
      (t) =>
        t.currency_code === params.currency_code &&
        (!params.period_start || t.created_at >= params.period_start) &&
        t.created_at <= params.period_end,
    )
    const totalIncome = matching.filter((t) => t.type === 'income').reduce((sum, t) => sum + Number(t.amount), 0)
    const totalExpense = matching.filter((t) => t.type === 'expense').reduce((sum, t) => sum + Number(t.amount), 0)
    const net = totalIncome - totalExpense

    return {
      period: params.period ?? null,
      period_start: params.period_start ?? null,
      period_end: params.period_end,
      currency_code: params.currency_code,
      total_income: totalIncome.toFixed(2),
      total_expense: totalExpense.toFixed(2),
      net: net.toFixed(2),
      by_currency: {
        [params.currency_code]: {
          currency_code: params.currency_code,
          total_income: totalIncome.toFixed(2),
          total_expense: totalExpense.toFixed(2),
          net: net.toFixed(2),
        },
      },
      transaction_count: matching.length,
    }
  },
}
