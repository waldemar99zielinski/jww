import { screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { HttpResponse, http } from 'msw'
import { Route, Routes } from 'react-router-dom'
import { beforeEach, describe, expect, it } from 'vitest'
import TransactionDetailPage from '@/pages/transactions/detail/index'
import { fixtureTransactions, fixtureUser } from '@/test/fixtures'
import { createTestStore, renderWithProviders } from '@/test/renderWithProviders'
import { server } from '@/test/server'

function renderPage(id = '1') {
  return renderWithProviders(
    <Routes>
      <Route path="/transactions/:id" element={<TransactionDetailPage />} />
    </Routes>,
    {
      route: `/transactions/${id}`,
      store: createTestStore({ status: 'authenticated', user: fixtureUser, error: null }),
    },
  )
}

describe('TransactionDetailPage', () => {
  beforeEach(() => {
    // Force the desktop layout variant so the details render as a plain card, not full-screen.
    window.innerWidth = 1280
    window.dispatchEvent(new Event('resize'))
  })

  it('renders every field of the transaction, including its attachments', async () => {
    renderPage()

    expect(await screen.findByText('Weekly shop')).toBeInTheDocument()
    expect(screen.getByText('Groceries')).toBeInTheDocument()
    expect(screen.getByText('Expense')).toBeInTheDocument()
    expect(screen.getByText('Waldemar')).toBeInTheDocument()
    expect(screen.getByText('Manual')).toBeInTheDocument()
    expect(screen.getByText('EUR')).toBeInTheDocument()
    expect(screen.getByText('Milk, bread and coffee')).toBeInTheDocument()
    expect(screen.getByText('#1')).toBeInTheDocument()
    expect(screen.getByText('receipt.jpg')).toBeInTheDocument()
  })

  it('opens every previewable attachment from its row and offers actions for each', async () => {
    renderPage()

    await screen.findByText('Weekly shop')

    // The row itself opens the lightbox, for the image and the plain-text file alike…
    expect(screen.getByRole('button', { name: 'Preview receipt.jpg' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Preview note.txt' })).toBeInTheDocument()
    // …and each row has an actions drawer button.
    expect(screen.getByRole('button', { name: 'Actions for receipt.jpg' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Actions for note.txt' })).toBeInTheDocument()
  })

  it('offers preview, download and delete in the attachment actions drawer', async () => {
    const user = userEvent.setup()
    renderPage()

    await screen.findByText('Weekly shop')

    await user.click(screen.getByRole('button', { name: 'Actions for receipt.jpg' }))

    expect(await screen.findByRole('button', { name: 'Preview' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Download' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Delete' })).toBeInTheDocument()
  })

  it('deletes an attachment from its actions drawer', async () => {
    const user = userEvent.setup()
    const transaction = fixtureTransactions[0]
    let attachments = [...transaction.attachments]
    let deletedId: string | null = null
    server.use(
      http.get('*/api/v1/transactions/:id', () => HttpResponse.json({ data: { ...transaction, attachments } })),
      http.delete('*/api/v1/transactions/:id/attachments/:attachmentID', ({ params }) => {
        deletedId = String(params.attachmentID)
        attachments = attachments.filter((attachment) => attachment.id !== deletedId)
        return new HttpResponse(null, { status: 204 })
      }),
    )

    renderPage()
    await screen.findByText('Weekly shop')

    await user.click(screen.getByRole('button', { name: 'Actions for note.txt' }))
    await user.click(await screen.findByRole('button', { name: 'Delete' }))
    await user.click(await screen.findByRole('button', { name: 'Delete attachment' }))

    await waitFor(() => expect(deletedId).toBe('attachment-2'))
    // The refetched transaction no longer carries the deleted attachment's row.
    await waitFor(() => expect(screen.queryByRole('button', { name: 'Preview note.txt' })).not.toBeInTheDocument())
  })

  it('opens the big-preview lightbox for a clicked attachment and closes it again', async () => {
    const user = userEvent.setup()
    renderPage()

    await screen.findByText('Weekly shop')

    await user.click(screen.getByRole('button', { name: 'Preview receipt.jpg' }))

    const dialog = await screen.findByRole('dialog', { name: 'Preview of receipt.jpg' })
    expect(within(dialog).getByAltText('Preview of receipt.jpg')).toHaveAttribute(
      'src',
      '/api/v1/attachment-content?grant=grant-attachment-1',
    )

    await user.click(within(dialog).getByRole('button', { name: 'Close preview' }))
    await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument())
  })

  it('shows a not-found message for an unknown id', async () => {
    server.use(http.get('*/api/v1/transactions/:id', () => new HttpResponse(null, { status: 404 })))

    renderPage('999')

    expect(await screen.findByText(/couldn't load this transaction/i)).toBeInTheDocument()
    await waitFor(() => expect(screen.queryByText('Weekly shop')).not.toBeInTheDocument())
  })

  it('opens edit mode when the server sends the amount as a bare JSON number', async () => {
    // The Go `Money` type marshals as a number, not a string — the client type is optimistic.
    const user = userEvent.setup()
    const numericAmount = { ...fixtureTransactions[0], amount: 42.5, attachments: [] } as unknown as (typeof fixtureTransactions)[number]
    server.use(http.get('*/api/v1/transactions/:id', () => HttpResponse.json({ data: numericAmount })))

    renderPage()
    await screen.findByText('Weekly shop')

    await user.click(screen.getByRole('button', { name: 'Edit' }))

    expect(await screen.findByLabelText('Transaction name')).toHaveValue('Weekly shop')
    expect(screen.getByLabelText('Amount')).toHaveValue('42,50')
  })

  it('opens edit mode, patches the changed fields, and returns to the read-only view', async () => {
    const user = userEvent.setup()
    let capturedBody: Record<string, unknown> | null = null
    server.use(
      http.patch('*/api/v1/transactions/:id', async ({ request, params }) => {
        capturedBody = (await request.json()) as Record<string, unknown>
        const transaction = fixtureTransactions.find((t) => t.id === Number(params.id))!
        return HttpResponse.json({ data: { ...transaction, ...capturedBody, attachments: [] } })
      }),
    )

    renderPage()
    await screen.findByText('Weekly shop')

    await user.click(screen.getByRole('button', { name: 'Edit' }))

    const nameInput = await screen.findByLabelText('Transaction name')
    expect(nameInput).toHaveValue('Weekly shop')
    // A pristine form offers Cancel, not Reset.
    expect(screen.getByRole('button', { name: 'Cancel' })).toBeInTheDocument()
    await user.clear(nameInput)
    await user.type(nameInput, 'Big shop')
    // Editing flips the secondary button to Reset.
    expect(screen.getByRole('button', { name: 'Reset' })).toBeInTheDocument()
    await user.click(screen.getByRole('button', { name: 'Save changes' }))

    await waitFor(() =>
      expect(capturedBody).toMatchObject({
        type: 'expense',
        // The display comma is converted back to whole minor units → the API's decimal form.
        amount: 42.5,
        currency_code: 'EUR',
        category_id: 'cat-groceries',
        user_id: fixtureUser.id,
        name: 'Big shop',
        description: 'Milk, bread and coffee',
      }),
    )
    // Saving leaves edit mode: the inputs are gone and the read-only fields are back.
    await waitFor(() => expect(screen.queryByLabelText('Transaction name')).not.toBeInTheDocument())
    expect(screen.getByText('Weekly shop')).toBeInTheDocument()
  })

  it('resets every edited field back to the saved values', async () => {
    const user = userEvent.setup()
    renderPage()
    await screen.findByText('Weekly shop')

    await user.click(screen.getByRole('button', { name: 'Edit' }))

    const nameInput = await screen.findByLabelText('Transaction name')
    await user.clear(nameInput)
    await user.type(nameInput, 'Changed')
    expect(nameInput).toHaveValue('Changed')

    await user.click(screen.getByRole('button', { name: 'Reset' }))

    expect(screen.getByLabelText('Transaction name')).toHaveValue('Weekly shop')
    expect(screen.getByLabelText('Transaction description')).toHaveValue('Milk, bread and coffee')
    // Back to pristine: Reset becomes Cancel again.
    expect(screen.getByRole('button', { name: 'Cancel' })).toBeInTheDocument()
  })

  it('cancels edit mode without a request when nothing was changed', async () => {
    const user = userEvent.setup()
    let patched = false
    server.use(
      http.patch('*/api/v1/transactions/:id', () => {
        patched = true
        return HttpResponse.json({ data: { ...fixtureTransactions[0], attachments: [] } })
      }),
    )

    renderPage()
    await screen.findByText('Weekly shop')

    await user.click(screen.getByRole('button', { name: 'Edit' }))
    await screen.findByLabelText('Transaction name')

    await user.click(screen.getByRole('button', { name: 'Cancel' }))

    await waitFor(() => expect(screen.queryByLabelText('Transaction name')).not.toBeInTheDocument())
    expect(screen.getByText('Weekly shop')).toBeInTheDocument()
    expect(patched).toBe(false)
  })
})
