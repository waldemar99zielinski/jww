import type {
  CurrencyDto,
  SummaryDto,
  SummaryPeriod,
  TransactionCategoryDto,
  TransactionDto,
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
      .sort((a, b) => a.created_at.localeCompare(b.created_at))
  },
  createTransaction(input: {
    type: TransactionType
    amount: number
    currency_code: string
    category_id: string
    user_id: string
    source: TransactionSource
    name: string
    description?: string
  }): TransactionDto {
    const category = categories.find((c) => c.id === input.category_id)
    const transaction: TransactionDto = {
      id: nextTransactionId++,
      type: input.type,
      amount: input.amount.toFixed(2),
      currency_code: input.currency_code,
      category_id: input.category_id,
      category_name: category?.name ?? 'Other',
      user_id: input.user_id,
      source: input.source,
      name: input.name,
      created_at: new Date().toISOString(),
      description: input.description ?? '',
    }
    transactions.push(transaction)
    return transaction
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
