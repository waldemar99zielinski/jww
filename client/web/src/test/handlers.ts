import { http, HttpResponse } from 'msw'
import {
  fixtureCategories,
  fixtureCurrencies,
  fixtureSummary,
  fixtureTokenResponse,
  fixtureTransactions,
  fixtureUser,
  fixtureUsers,
} from '@/test/fixtures'

/**
 * Handlers mirror each `api/*.ts` function's real path/method/query shape — keep them in sync
 * when a resource file changes (see this folder's README). Paths use the `*` wildcard prefix so
 * they match regardless of `VITE_API_BASE_URL` (which varies per machine/.env), not just the
 * default `http://localhost:8000`.
 */
export const handlers = [
  http.post('*/api/v1/auth/login', () => HttpResponse.json({ data: fixtureTokenResponse })),
  http.get('*/api/v1/auth/me', () => HttpResponse.json({ data: fixtureUser })),

  http.get('*/api/v1/users', () => HttpResponse.json({ data: fixtureUsers })),

  http.get('*/api/v1/currencies', () => HttpResponse.json({ data: fixtureCurrencies })),
  http.post('*/api/v1/currencies', async ({ request }) => {
    const body = (await request.json()) as { code: string; name: string; is_base: boolean }
    return HttpResponse.json({ data: { code: body.code, name: body.name, is_base: body.is_base } })
  }),
  http.put('*/api/v1/currencies/:code', async ({ request, params }) => {
    const body = (await request.json()) as { name: string; is_base: boolean }
    return HttpResponse.json({ data: { code: String(params.code), name: body.name, is_base: body.is_base } })
  }),

  http.get('*/api/v1/transaction-categories', ({ request }) => {
    const url = new URL(request.url)
    const type = url.searchParams.get('type')
    const activeOnly = url.searchParams.get('active_only') !== 'false'
    const filtered = fixtureCategories.filter(
      (category) => (type ? category.type === type : true) && (activeOnly ? category.active : true),
    )
    return HttpResponse.json({ data: filtered })
  }),
  http.post('*/api/v1/transaction-categories', async ({ request }) => {
    const body = (await request.json()) as { type: string; name: string; sort_order: number }
    return HttpResponse.json({
      data: { id: 'new-category-id', type: body.type, name: body.name, active: true, sort_order: body.sort_order },
    })
  }),
  http.put('*/api/v1/transaction-categories/:id', async ({ request, params }) => {
    const body = (await request.json()) as { name: string; active: boolean; sort_order: number }
    const existing = fixtureCategories.find((category) => category.id === params.id)
    return HttpResponse.json({
      data: { id: String(params.id), type: existing?.type ?? 'expense', ...body },
    })
  }),

  http.get('*/api/v1/transactions', ({ request }) => {
    const url = new URL(request.url)
    const filtered = fixtureTransactions.filter((transaction) => {
      const userId = url.searchParams.get('user_id')
      const categoryId = url.searchParams.get('category_id')
      const type = url.searchParams.get('type')
      const currencyCode = url.searchParams.get('currency_code')
      const source = url.searchParams.get('source')
      return (
        (!userId || transaction.user_id === userId) &&
        (!categoryId || transaction.category_id === categoryId) &&
        (!type || transaction.type === type) &&
        (!currencyCode || transaction.currency_code === currencyCode) &&
        (!source || transaction.source === source)
      )
    })
    return HttpResponse.json({ data: filtered })
  }),
  http.post('*/api/v1/transactions', async ({ request }) => {
    const body = (await request.json()) as {
      type: string
      amount: number
      currency_code: string
      category_id: string
      user_id: string
      source: string
      name: string
      description?: string
    }
    const category = fixtureCategories.find((c) => c.id === body.category_id)
    return HttpResponse.json({
      data: {
        id: 999,
        type: body.type,
        amount: body.amount.toFixed(2),
        currency_code: body.currency_code,
        category_id: body.category_id,
        category_name: category?.name ?? 'Unknown',
        user_id: body.user_id,
        source: body.source,
        name: body.name,
        created_at: new Date(0).toISOString(),
        description: body.description ?? '',
      },
    })
  }),

  http.get('*/api/v1/summary', () => HttpResponse.json({ data: fixtureSummary })),
]
