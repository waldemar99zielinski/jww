import { http, HttpResponse } from 'msw'
import { describe, expect, it } from 'vitest'
import HomePage from '@/pages/home/index'
import { server } from '@/test/server'
import { fixtureSummary, fixtureUser } from '@/test/fixtures'
import { createTestStore, renderWithProviders } from '@/test/renderWithProviders'
import { normalizeText } from '@/test/normalizeText'
import { formatAmount, formatSignedAmount } from '@/lib/currency'

function renderHome() {
  const store = createTestStore({ status: 'authenticated', user: fixtureUser, error: null })
  return renderWithProviders(<HomePage />, { store, route: '/' })
}

describe('HomePage', () => {
  it('renders balance, income/expense totals, and a category breakdown from mocked data', async () => {
    const { findByText, findAllByText } = renderHome()

    await findByText(fixtureUser.display_name)

    // Summary totals come from the mocked /summary response (2500 income, 42.5 expense). The
    // signed amounts can legitimately appear twice (HeroBalanceCard's total + a matching
    // transaction row), so assert presence rather than a single unique match.
    expect(await findByText(normalizeText(formatAmount(2457.5)))).toBeInTheDocument()
    expect((await findAllByText(normalizeText(formatSignedAmount(2500, 'income')))).length).toBeGreaterThan(0)
    expect((await findAllByText(normalizeText(formatSignedAmount(42.5, 'expense')))).length).toBeGreaterThan(0)

    // Category breakdown is derived from the mocked /transactions response's single expense.
    // "Groceries" legitimately appears twice (breakdown item + the matching transaction row's
    // subtitle), so assert presence via findAllByText rather than a single unique match.
    expect((await findAllByText('Groceries')).length).toBeGreaterThan(0)
    expect(await findByText(normalizeText(formatAmount(42.5)))).toBeInTheDocument()

    // Recent transactions list shows both mocked transactions.
    expect(await findByText('Weekly shop')).toBeInTheDocument()
    expect(await findByText('October salary')).toBeInTheDocument()
  })

  it('requests the summary without a user filter', async () => {
    const summaryUserIds: Array<string | null> = []
    server.use(
      http.get('*/api/v1/summary', ({ request }) => {
        summaryUserIds.push(new URL(request.url).searchParams.get('user_id'))
        return HttpResponse.json({ data: fixtureSummary })
      }),
    )

    const { findByText } = renderHome()
    await findByText(fixtureUser.display_name)

    expect(summaryUserIds).toContain(null)
  })

  it('shows empty states when there are no transactions this month', async () => {
    server.use(
      http.get('*/api/v1/transactions', () => HttpResponse.json({ data: [] })),
      http.get('*/api/v1/summary', () =>
        HttpResponse.json({
          data: {
            period: 'month',
            period_start: null,
            period_end: '2026-10-31T23:59:59Z',
            currency_code: 'EUR',
            total_income: '0',
            total_expense: '0',
            net: '0',
            by_currency: {},
            transaction_count: 0,
          },
        }),
      ),
    )

    const { findByText } = renderHome()

    expect(await findByText('No expenses yet this month.')).toBeInTheDocument()
    expect(await findByText('No transactions yet.')).toBeInTheDocument()
    expect(await findByText(normalizeText(formatAmount(0)))).toBeInTheDocument()
  })
})
