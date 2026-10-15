import { waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { http, HttpResponse } from 'msw'
import { describe, expect, it } from 'vitest'
import LoginPage from '@/pages/login/LoginPage'
import { server } from '@/test/server'
import { renderWithProviders } from '@/test/renderWithProviders'

describe('LoginPage', () => {
  it('shows an error and stays unauthenticated for invalid credentials', async () => {
    server.use(
      http.post('*/api/v1/auth/login', () =>
        HttpResponse.json({ error: { message: 'unauthorized' } }, { status: 401 }),
      ),
    )

    const user = userEvent.setup()
    const { getByLabelText, findByText, store } = renderWithProviders(<LoginPage />, { route: '/login' })

    await user.type(getByLabelText('Username'), 'someone')
    await user.type(getByLabelText('Password'), 'wrong-password')
    await user.click(getByLabelText('Password'))
    await user.keyboard('{Enter}')

    expect(await findByText('Invalid username or password.')).toBeInTheDocument()
    expect(store.getState().auth.status).toBe('error')
  })

  it('authenticates and reaches the authenticated store state for valid credentials', async () => {
    const user = userEvent.setup()
    const { getByLabelText, getByRole, store } = renderWithProviders(<LoginPage />, { route: '/login' })

    await user.type(getByLabelText('Username'), 'waldemar')
    await user.type(getByLabelText('Password'), 'correct-password')
    await user.click(getByRole('button', { name: /sign in/i }))

    await waitFor(() => expect(store.getState().auth.status).toBe('authenticated'))
    expect(store.getState().auth.user?.username).toBe('waldemar')
  })
})
