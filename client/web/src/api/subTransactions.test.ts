import { http, HttpResponse } from 'msw'
import { describe, expect, it } from 'vitest'
import { createSubTransaction, deleteSubTransaction, updateSubTransaction } from '@/api/subTransactions'
import { fixtureItemisedTransaction } from '@/test/fixtures'
import { server } from '@/test/server'

/** One captured outgoing request: the parts a client contract cares about. */
interface CapturedRequest {
  method: string
  path: string
  body: unknown
}

describe('subTransactions api', () => {
  it('creates a line item with POST, carrying the amount in cents', async () => {
    let captured: CapturedRequest | null = null
    server.use(
      http.post('*/api/v1/transactions/3/sub-transactions', async ({ request }) => {
        captured = { method: request.method, path: new URL(request.url).pathname, body: await request.json() }
        return HttpResponse.json({ data: fixtureItemisedTransaction }, { status: 201 })
      }),
    )

    const result = await createSubTransaction(3, { name: 'Coffee beans', label: 'pantry', amount: 850 })

    expect(captured).toEqual({
      method: 'POST',
      path: '/api/v1/transactions/3/sub-transactions',
      body: { name: 'Coffee beans', label: 'pantry', amount: 850 },
    })
    expect(result).toEqual(fixtureItemisedTransaction)
  })

  it('updates a line item with PATCH and a partial body', async () => {
    let captured: CapturedRequest | null = null
    server.use(
      http.patch('*/api/v1/transactions/3/sub-transactions/sub-transaction-1', async ({ request }) => {
        captured = { method: request.method, path: new URL(request.url).pathname, body: await request.json() }
        return HttpResponse.json({ data: fixtureItemisedTransaction })
      }),
    )

    await updateSubTransaction(3, 'sub-transaction-1', { amount: 700 })

    expect(captured).toEqual({
      method: 'PATCH',
      path: '/api/v1/transactions/3/sub-transactions/sub-transaction-1',
      body: { amount: 700 },
    })
  })

  it('deletes a line item with DELETE and resolves with no body on a 204', async () => {
    let captured: { method: string; path: string } | null = null
    server.use(
      http.delete('*/api/v1/transactions/3/sub-transactions/sub-transaction-1', ({ request }) => {
        captured = { method: request.method, path: new URL(request.url).pathname }
        return new HttpResponse(null, { status: 204 })
      }),
    )

    await expect(deleteSubTransaction(3, 'sub-transaction-1')).resolves.toBeUndefined()

    expect(captured).toEqual({
      method: 'DELETE',
      path: '/api/v1/transactions/3/sub-transactions/sub-transaction-1',
    })
  })

  it('rejects with the server status when a mutation fails', async () => {
    server.use(
      http.post('*/api/v1/transactions/999/sub-transactions', () => new HttpResponse(null, { status: 404 })),
    )

    await expect(createSubTransaction(999, { name: 'Ghost', label: '', amount: 100 })).rejects.toMatchObject({
      status: 404,
    })
  })
})
