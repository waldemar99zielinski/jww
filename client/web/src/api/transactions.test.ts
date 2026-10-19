import { http, HttpResponse } from 'msw'
import { describe, expect, it } from 'vitest'
import { createTransaction, listTransactions, updateTransaction } from '@/api/transactions'
import { fixtureTransactions } from '@/test/fixtures'
import { server } from '@/test/server'

type WireBody = {
  amount?: number
  sub_transactions?: { name: string; label: string; amount: number }[]
}

describe('transactions api money contract', () => {
  it('sends the sort key and order as query params', async () => {
    let requested: URL | undefined
    server.use(
      http.get('*/api/v1/transactions', ({ request }) => {
        requested = new URL(request.url)
        return HttpResponse.json({ data: [] })
      }),
    )

    await listTransactions({ sort_key: 'date', sort_order: 'asc' })

    expect(requested?.searchParams.get('sort_key')).toBe('date')
    expect(requested?.searchParams.get('sort_order')).toBe('asc')
  })

  it('sends the top-level amount in cents', async () => {
    let captured: WireBody = {}
    server.use(
      http.post('*/api/v1/transactions', async ({ request }) => {
        captured = (await request.json()) as WireBody
        return HttpResponse.json({ data: fixtureTransactions[0] }, { status: 201 })
      }),
    )

    await createTransaction({
      type: 'expense',
      amount: 4127,
      currency_code: 'EUR',
      category_id: 'cat-groceries',
      user_id: 'user-1',
      source: 'manual',
      name: 'Lunch',
    })

    expect(captured.amount).toBe(4127)
  })

  it('also sends every line-item amount in cents', async () => {
    let captured: WireBody = {}
    server.use(
      http.post('*/api/v1/transactions', async ({ request }) => {
        captured = (await request.json()) as WireBody
        return HttpResponse.json({ data: fixtureTransactions[0] }, { status: 201 })
      }),
    )

    await createTransaction({
      type: 'expense',
      amount: 0,
      currency_code: 'EUR',
      category_id: 'cat-groceries',
      user_id: 'user-1',
      source: 'manual',
      name: 'Lunch',
      sub_transactions: [
        { name: 'Milk', label: 'dairy', amount: 649 },
        { name: 'Coupon', label: 'discount', amount: -100 },
      ],
    })

    expect(captured.sub_transactions).toEqual([
      { name: 'Milk', label: 'dairy', amount: 649 },
      { name: 'Coupon', label: 'discount', amount: -100 },
    ])
  })

  it('sends a patched amount in cents', async () => {
    let captured: WireBody = {}
    server.use(
      http.patch('*/api/v1/transactions/1', async ({ request }) => {
        captured = (await request.json()) as WireBody
        return HttpResponse.json({ data: fixtureTransactions[0] })
      }),
    )

    await updateTransaction(1, { amount: 4127 })

    expect(captured.amount).toBe(4127)
  })
})
