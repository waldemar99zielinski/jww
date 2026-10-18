import { http, HttpResponse } from 'msw'
import {
  fixtureAttachment,
  fixtureCategories,
  fixtureCurrencies,
  fixturePrefill,
  fixtureSummary,
  fixtureTokenResponse,
  fixtureTransactions,
  fixtureUser,
  fixtureUsers,
} from '@/test/fixtures'

/**
 * Handlers mirror each `api/*.ts` function's real path/method/query shape — keep them in sync
 * when a resource file changes (see this folder's README). Paths use the `*` wildcard prefix so
 * they match regardless of the API origin (which varies per machine/.env, and is same-origin in
 * dev), not just the default `http://localhost:8001`.
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
  http.get('*/api/v1/transactions/:id', ({ params }) => {
    const transaction = fixtureTransactions.find((t) => t.id === Number(params.id))
    if (!transaction) return new HttpResponse(null, { status: 404 })
    // GET /transactions/{id} is the only response that mints preview grants (for previewable media).
    const attachments = transaction.attachments.map((attachment) => ({
      ...attachment,
      preview_grant:
        attachment.media_type.startsWith('image/') || attachment.media_type === 'application/pdf'
          ? `grant-${attachment.id}`
          : null,
    }))
    return HttpResponse.json({ data: { ...transaction, attachments } })
  }),
  // PATCH mirrors GET's detail shape (preview grants included) and echoes the merged body.
  http.patch('*/api/v1/transactions/:id', async ({ request, params }) => {
    const transaction = fixtureTransactions.find((t) => t.id === Number(params.id))
    if (!transaction) return new HttpResponse(null, { status: 404 })
    const body = (await request.json()) as Record<string, unknown>
    const updated = { ...transaction, ...body }
    const attachments = updated.attachments.map((attachment) => ({
      ...attachment,
      preview_grant:
        attachment.media_type.startsWith('image/') || attachment.media_type === 'application/pdf'
          ? `grant-${attachment.id}`
          : null,
    }))
    return HttpResponse.json({ data: { ...updated, attachments } })
  }),
  http.get('*/api/v1/transactions/:id/attachments/:attachmentID/download', () => new HttpResponse(new Blob(['bytes']), { status: 200 })),
  http.delete('*/api/v1/transactions/:id/attachments/:attachmentID', () => new HttpResponse(null, { status: 204 })),
  http.get('*/api/v1/attachment-content', () => new HttpResponse(new Blob(['preview']), { status: 200 })),
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
      created_at?: string
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
        created_at: body.created_at ?? new Date(0).toISOString(),
        description: body.description ?? '',
      },
    })
  }),

  http.get('*/api/v1/summary', () => HttpResponse.json({ data: fixtureSummary })),

  http.post('*/api/v1/transaction-attachments', () => HttpResponse.json({ data: fixtureAttachment }, { status: 201 })),
  http.delete('*/api/v1/transaction-attachments/:attachmentID', () => new HttpResponse(null, { status: 204 })),
  http.post('*/api/v1/transaction-attachments/:attachmentID/extraction', ({ params }) =>
    HttpResponse.json({ data: { ...fixturePrefill, attachment_id: String(params.attachmentID) } }, { status: 201 }),
  ),
]
