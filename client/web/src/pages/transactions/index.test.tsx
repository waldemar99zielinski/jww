import { screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { http, HttpResponse } from 'msw'
import { beforeEach, describe, expect, it } from 'vitest'
import TransactionsPage from '@/pages/transactions/index'
import { fixtureTransactions, fixtureUser } from '@/test/fixtures'
import { createTestStore, renderWithProviders } from '@/test/renderWithProviders'
import { server } from '@/test/server'

function renderPage() {
  return renderWithProviders(<TransactionsPage />, {
    store: createTestStore({ status: 'authenticated', user: fixtureUser, error: null }),
  })
}

describe('TransactionsPage', () => {
  beforeEach(() => {
    // Force the desktop layout variant so the filter panel renders inline (not as a BottomSheet).
    window.innerWidth = 1280
    window.dispatchEvent(new Event('resize'))
  })

  it('lists the transactions returned for the default (unfiltered) request', async () => {
    renderPage()

    expect(await screen.findByText('Weekly shop')).toBeInTheDocument()
    expect(await screen.findByText('October salary')).toBeInTheDocument()
  })

  it('narrows the list to one type after applying the Type filter', async () => {
    const user = userEvent.setup()
    renderPage()

    await screen.findByText('Weekly shop')
    await screen.findByText('October salary')

    await user.click(screen.getByRole('button', { name: /filters/i }))
    await user.selectOptions(screen.getByLabelText('Type'), 'Income')
    await user.click(screen.getByRole('button', { name: /^apply$/i }))

    await waitFor(() => expect(screen.queryByText('Weekly shop')).not.toBeInTheDocument())
    expect(screen.getByText('October salary')).toBeInTheDocument()
  })

  it('queries without a user filter and renders rows in the order the server returns them', async () => {
    const requestedUserIds: Array<string | null> = []
    server.use(
      http.get('*/api/v1/transactions', ({ request }) => {
        requestedUserIds.push(new URL(request.url).searchParams.get('user_id'))
        return HttpResponse.json({
          data: [
            { ...fixtureTransactions[1], id: 3, name: 'Newer shop', created_at: '2026-10-10T12:00:00Z' },
            { ...fixtureTransactions[0], id: 1, name: 'Older shop', created_at: '2026-10-10T08:00:00Z' },
          ],
        })
      }),
    )

    renderPage()

    const newer = await screen.findByText('Newer shop')
    const older = screen.getByText('Older shop')

    // No default user scope on the request…
    expect(requestedUserIds).toContain(null)
    // …and no client re-sort: the server's first row stays first in the DOM.
    expect(newer.compareDocumentPosition(older) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy()
  })
})
