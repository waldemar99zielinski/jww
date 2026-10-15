import { act, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { http, HttpResponse } from 'msw'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { fixtureUser } from '@/test/fixtures'
import { renderWithProviders, createTestStore } from '@/test/renderWithProviders'
import { server } from '@/test/server'
import AddTransactionPage from './index'

function renderPage() {
  return renderWithProviders(<AddTransactionPage />, {
    route: '/add',
    store: createTestStore({ status: 'authenticated', user: fixtureUser, error: null }),
  })
}

async function selectCategory(user: ReturnType<typeof userEvent.setup>, name: string | RegExp) {
  await user.click(await screen.findByRole('button', { name: /select category/i }))
  await user.click(await screen.findByRole('button', { name }))
}

describe('AddTransactionPage', () => {
  beforeEach(() => {
    // jsdom doesn't implement these; the receipt-picking flow calls them.
    window.URL.createObjectURL = vi.fn(() => 'blob:fixture')
    window.URL.revokeObjectURL = vi.fn()
  })

  it('submits the filled-in form with the right payload shape', async () => {
    const user = userEvent.setup()
    let capturedBody: Record<string, unknown> | null = null
    server.use(
      http.post('*/api/v1/transactions', async ({ request }) => {
        const body = (await request.json()) as Record<string, unknown>
        capturedBody = body
        return HttpResponse.json({
          data: { id: 1, category_name: 'Groceries', created_at: '2026-10-10T00:00:00Z', description: '', ...body },
        })
      }),
    )

    renderPage()

    await user.type(await screen.findByLabelText('Transaction name'), 'Coffee')
    await user.type(screen.getByLabelText('Amount'), '25')
    await selectCategory(user, /groceries/i)

    await user.click(screen.getByRole('button', { name: /save transaction/i }))

    await waitFor(() =>
      expect(capturedBody).toMatchObject({
        type: 'expense',
        amount: 25,
        currency_code: 'EUR',
        category_id: 'cat-groceries',
        user_id: fixtureUser.id,
        source: 'manual',
        name: 'Coffee',
      }),
    )
  })

  it('disables saving until a category, name, and amount are all set', async () => {
    const user = userEvent.setup()
    renderPage()

    expect(await screen.findByRole('button', { name: /save transaction/i })).toBeDisabled()

    await user.type(screen.getByLabelText('Transaction name'), 'Coffee')
    await user.type(screen.getByLabelText('Amount'), '25')
    expect(screen.getByRole('button', { name: /save transaction/i })).toBeDisabled()

    await selectCategory(user, /groceries/i)
    expect(screen.getByRole('button', { name: /save transaction/i })).not.toBeDisabled()
  })

  describe('receipt picking (fake OCR)', () => {
    // `shouldAdvanceTime` lets RTL's internal `findBy*`/`waitFor` polling (which relies on real
    // timers) keep progressing, while we still control the 1200ms OCR timeout explicitly below.
    beforeEach(() => vi.useFakeTimers({ shouldAdvanceTime: true }))
    afterEach(() => vi.useRealTimers())

    it('fills amount/name with the canned values once "parsing" finishes', async () => {
      const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime })
      const { container } = renderPage()

      await user.click(await screen.findByRole('button', { name: 'Add attachment' }))
      await user.click(await screen.findByRole('button', { name: 'Choose a file' }))

      const fileInput = container.querySelectorAll('input[type="file"]')[1] as HTMLInputElement
      const file = new File(['content'], 'receipt.jpg', { type: 'image/jpeg' })
      await user.upload(fileInput, file)

      expect(screen.getByText('Reading file…')).toBeInTheDocument()

      await act(() => vi.advanceTimersByTimeAsync(1200))

      expect(screen.getByLabelText('Amount')).toHaveValue('184,32')
      expect(screen.getByLabelText('Transaction name')).toHaveValue('Supermarket receipt')
      expect(screen.getByText('Attachment ready')).toBeInTheDocument()
    })
  })
})
