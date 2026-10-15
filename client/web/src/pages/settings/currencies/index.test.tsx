import { screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { http, HttpResponse } from 'msw'
import { describe, expect, it } from 'vitest'
import { fixtureUser } from '@/test/fixtures'
import { createTestStore, renderWithProviders } from '@/test/renderWithProviders'
import { server } from '@/test/server'
import SettingsCurrenciesPage from './index'

function setDesktopViewport() {
  Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 1280 })
}

function authedStore() {
  return createTestStore({ status: 'authenticated', user: fixtureUser, error: null })
}

describe('SettingsCurrenciesPage', () => {
  it('auto-sets is_base on the very first currency ever created', async () => {
    setDesktopViewport()
    const user = userEvent.setup()
    server.use(http.get('*/api/v1/currencies', () => HttpResponse.json({ data: [] })))
    let capturedBody: unknown
    server.use(
      http.post('*/api/v1/currencies', async ({ request }) => {
        capturedBody = await request.json()
        return HttpResponse.json({ data: capturedBody })
      }),
    )
    renderWithProviders(<SettingsCurrenciesPage />, { store: authedStore() })

    await screen.findByText(/no currencies yet/i)
    await user.type(screen.getByPlaceholderText('PLN'), 'pln')
    await user.type(screen.getByPlaceholderText('Polish złoty'), 'Polish złoty')
    await user.click(screen.getByRole('button', { name: 'Add' }))

    await waitFor(() => expect(capturedBody).toEqual({ code: 'PLN', name: 'Polish złoty', is_base: true }))
  })

  it('resends the existing name alongside is_base:true when making a currency the base', async () => {
    setDesktopViewport()
    const user = userEvent.setup()
    let capturedBody: unknown
    server.use(
      http.put('*/api/v1/currencies/:code', async ({ request, params }) => {
        capturedBody = await request.json()
        return HttpResponse.json({ data: { code: String(params.code), ...(capturedBody as object) } })
      }),
    )
    renderWithProviders(<SettingsCurrenciesPage />, { store: authedStore() })

    // Only USD is non-base in the default fixtures, so this button is unambiguous without
    // scoping to its row.
    await user.click(await screen.findByRole('button', { name: /make base/i }))

    await waitFor(() => expect(capturedBody).toEqual({ name: 'US Dollar', is_base: true }))
  })

  it('keeps Add disabled until the code is exactly 3 characters and a name is set', async () => {
    setDesktopViewport()
    const user = userEvent.setup()
    renderWithProviders(<SettingsCurrenciesPage />, { store: authedStore() })

    await screen.findByText('Euro')
    const addButton = screen.getByRole('button', { name: 'Add' })
    expect(addButton).toBeDisabled()

    await user.type(screen.getByPlaceholderText('PLN'), 'PL')
    await user.type(screen.getByPlaceholderText('Polish złoty'), 'Polish złoty')
    expect(addButton).toBeDisabled()
  })
})
