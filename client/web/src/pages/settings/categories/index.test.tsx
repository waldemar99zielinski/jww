import { screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { http, HttpResponse } from 'msw'
import { describe, expect, it } from 'vitest'
import { fixtureUser } from '@/test/fixtures'
import { createTestStore, renderWithProviders } from '@/test/renderWithProviders'
import { server } from '@/test/server'
import SettingsCategoriesPage from './index'

function setDesktopViewport() {
  Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 1280 })
}

function authedStore() {
  return createTestStore({ status: 'authenticated', user: fixtureUser, error: null })
}

describe('SettingsCategoriesPage', () => {
  it('lists expense categories by default and switches to income categories on tab change', async () => {
    setDesktopViewport()
    const user = userEvent.setup()
    renderWithProviders(<SettingsCategoriesPage />, { store: authedStore() })

    expect(await screen.findByText('Groceries')).toBeInTheDocument()
    expect(screen.getByText('Transport')).toBeInTheDocument()
    expect(screen.getByText('Rent')).toBeInTheDocument()
    expect(screen.queryByText('Salary')).not.toBeInTheDocument()

    await user.click(screen.getByRole('button', { name: 'Income' }))

    expect(await screen.findByText('Salary')).toBeInTheDocument()
    expect(screen.getByText('Freelance')).toBeInTheDocument()
    expect(screen.queryByText('Groceries')).not.toBeInTheDocument()
  })

  it('resends the existing name and sort_order alongside the flipped active flag', async () => {
    setDesktopViewport()
    const user = userEvent.setup()
    let capturedBody: unknown
    server.use(
      http.put('*/api/v1/transaction-categories/:id', async ({ request, params }) => {
        capturedBody = await request.json()
        return HttpResponse.json({ data: { id: String(params.id), type: 'expense', ...(capturedBody as object) } })
      }),
    )
    renderWithProviders(<SettingsCategoriesPage />, { store: authedStore() })

    const groceriesRow = (await screen.findByText('Groceries')).closest('div') as HTMLElement
    await user.click(within(groceriesRow).getByRole('button', { name: /off|deactivate/i }))

    await waitFor(() => expect(capturedBody).toEqual({ name: 'Groceries', active: false, sort_order: 0 }))
  })

  it('creates a new category and clears the name field on success', async () => {
    setDesktopViewport()
    const user = userEvent.setup()
    let capturedBody: unknown
    server.use(
      http.post('*/api/v1/transaction-categories', async ({ request }) => {
        capturedBody = await request.json()
        const body = capturedBody as { type: string; name: string; sort_order: number }
        return HttpResponse.json({ data: { id: 'new-id', type: body.type, name: body.name, active: true, sort_order: body.sort_order } })
      }),
    )
    renderWithProviders(<SettingsCategoriesPage />, { store: authedStore() })

    await screen.findByText('Groceries')
    const input = screen.getByPlaceholderText('e.g. Groceries')
    await user.type(input, 'Pets')
    await user.click(screen.getByRole('button', { name: 'Add' }))

    await waitFor(() => expect(capturedBody).toEqual({ type: 'expense', name: 'Pets', sort_order: 0 }))
    await waitFor(() => expect(input).toHaveValue(''))
  })
})
