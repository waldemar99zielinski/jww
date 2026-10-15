import { screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { beforeEach, describe, expect, it } from 'vitest'
import TransactionsPage from '@/pages/transactions/index'
import { createTestStore, renderWithProviders } from '@/test/renderWithProviders'
import { fixtureUser } from '@/test/fixtures'

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

  it('lists the signed-in user\'s transactions', async () => {
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
})
