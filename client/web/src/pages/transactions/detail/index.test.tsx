import { screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { HttpResponse, http } from 'msw'
import { Route, Routes } from 'react-router-dom'
import { beforeEach, describe, expect, it } from 'vitest'
import type { TransactionDto } from '@/api/types'
import TransactionDetailPage from '@/pages/transactions/detail/index'
import { fixtureItemisedTransaction, fixtureTransactions, fixtureUser } from '@/test/fixtures'
import { createTestStore, renderWithProviders } from '@/test/renderWithProviders'
import { server } from '@/test/server'

function renderPage(id = '1') {
  return renderWithProviders(
    <Routes>
      <Route path="/transactions" element={<div>Transactions list</div>} />
      <Route path="/transactions/:id" element={<TransactionDetailPage />} />
    </Routes>,
    {
      route: `/transactions/${id}`,
      store: createTestStore({ status: 'authenticated', user: fixtureUser, error: null }),
    },
  )
}

/** Deep-clones a fixture so a stateful override never mutates the shared object. */
function cloneTransaction(transaction: TransactionDto): TransactionDto {
  return JSON.parse(JSON.stringify(transaction)) as TransactionDto
}

/** Mirrors the server rule: whenever items exist the parent amount is their Σ. */
function deriveAmount(transaction: TransactionDto): TransactionDto {
  if (transaction.sub_transactions.length === 0) return transaction
  const total = transaction.sub_transactions.reduce((sum, item) => sum + item.amount, 0)
  return { ...transaction, amount: total }
}

interface ItemCalls {
  post: { transactionId: string; body: { name: string; label: string; amount: number } }[]
  patch: { transactionId: string; subTransactionId: string; body: Record<string, unknown> }[]
  remove: { transactionId: string; subTransactionId: string }[]
}

/**
 * A stateful GET + sub-transaction CRUD override for one already-itemised transaction
 * (`fixtureItemisedTransaction` is deliberately absent from the shared GET handler's list), so a
 * mutation's invalidation refetch observes the server's recomputed state — including the rule that
 * deleting the last item leaves `amount` at its last computed value. Records every mutation so a
 * test can assert the exact endpoint each row action hits.
 */
function useStatefulItemisedTransaction(initial: TransactionDto = fixtureItemisedTransaction) {
  const state = { current: cloneTransaction(initial), gets: 0, nextId: 1 }
  const calls: ItemCalls = { post: [], patch: [], remove: [] }

  server.use(
    http.get('*/api/v1/transactions/:id', ({ params }) => {
      if (Number(params.id) !== state.current.id) return new HttpResponse(null, { status: 404 })
      state.gets += 1
      return HttpResponse.json({ data: state.current })
    }),
    http.post('*/api/v1/transactions/:transactionID/sub-transactions', async ({ request, params }) => {
      const body = (await request.json()) as { name: string; label: string; amount: number }
      calls.post.push({ transactionId: String(params.transactionID), body })
      const item = {
        id: `sub-new-${state.nextId++}`,
        name: body.name,
        label: body.label,
        amount: body.amount,
      }
      state.current = deriveAmount({ ...state.current, sub_transactions: [...state.current.sub_transactions, item] })
      return HttpResponse.json({ data: state.current }, { status: 201 })
    }),
    http.patch('*/api/v1/transactions/:transactionID/sub-transactions/:subTransactionID', async ({ request, params }) => {
      const body = (await request.json()) as { name?: string; label?: string; amount?: number }
      calls.patch.push({
        transactionId: String(params.transactionID),
        subTransactionId: String(params.subTransactionID),
        body,
      })
      const target = String(params.subTransactionID)
      state.current = deriveAmount({
        ...state.current,
        sub_transactions: state.current.sub_transactions.map((item) =>
          item.id === target
            ? {
                ...item,
                ...(body.name !== undefined ? { name: body.name } : {}),
                ...(body.label !== undefined ? { label: body.label } : {}),
                ...(body.amount !== undefined ? { amount: body.amount } : {}),
              }
            : item,
        ),
      })
      return HttpResponse.json({ data: state.current })
    }),
    http.delete('*/api/v1/transactions/:transactionID/sub-transactions/:subTransactionID', ({ params }) => {
      calls.remove.push({
        transactionId: String(params.transactionID),
        subTransactionId: String(params.subTransactionID),
      })
      const remaining = state.current.sub_transactions.filter((item) => item.id !== String(params.subTransactionID))
      // Σ while items remain; once the list empties the server leaves `amount` at its last computed value.
      state.current =
        remaining.length > 0
          ? {
              ...state.current,
              sub_transactions: remaining,
              amount: remaining.reduce((sum, item) => sum + item.amount, 0),
            }
          : { ...state.current, sub_transactions: remaining }
      return new HttpResponse(null, { status: 204 })
    }),
  )

  return { state, calls }
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
    expect(screen.getByText('attachment.jpg')).toBeInTheDocument()
  })

  it('opens every previewable attachment from its row and offers actions for each', async () => {
    renderPage()

    await screen.findByText('Weekly shop')

    // The row itself opens the lightbox, for the image and the plain-text file alike…
    expect(screen.getByRole('button', { name: 'Preview attachment.jpg' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Preview note.txt' })).toBeInTheDocument()
    // …and each row has an actions drawer button.
    expect(screen.getByRole('button', { name: 'Actions for attachment.jpg' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Actions for note.txt' })).toBeInTheDocument()
  })

  it('offers preview, download and delete in the attachment actions drawer', async () => {
    const user = userEvent.setup()
    renderPage()

    await screen.findByText('Weekly shop')

    await user.click(screen.getByRole('button', { name: 'Actions for attachment.jpg' }))

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

  it('uploads a picked file directly onto the transaction and refreshes the attachment list', async () => {
    const user = userEvent.setup()
    const transaction = fixtureTransactions[0]
    let attachments = [...transaction.attachments]
    let uploadedName: string | null = null
    server.use(
      http.get('*/api/v1/transactions/:id', () =>
        HttpResponse.json({
          data: {
            ...transaction,
            attachments: attachments.map((attachment) => ({
              ...attachment,
              preview_grant: attachment.media_type.startsWith('image/') ? `grant-${attachment.id}` : null,
            })),
          },
        }),
      ),
      http.post('*/api/v1/transactions/:id/attachments', ({ params }) => {
        uploadedName = 'receipt.png'
        const attachment = {
          id: 'attachment-new',
          transaction_id: Number(params.id),
          original_filename: uploadedName,
          media_type: 'image/png',
          byte_size: 9,
          created_at: '2026-10-10T10:00:00Z',
        }
        attachments = [...attachments, attachment]
        return HttpResponse.json({ data: attachment }, { status: 201 })
      }),
    )

    const { container } = renderPage()
    await screen.findByText('Weekly shop')

    await user.click(screen.getByRole('button', { name: 'Add attachment' }))
    await user.click(await screen.findByRole('button', { name: 'Choose files' }))
    const fileInput = container.querySelectorAll('input[type="file"]')[1] as HTMLInputElement
    const file = new File(['new bytes'], 'receipt.png', { type: 'image/png' })
    await user.upload(fileInput, file)

    await waitFor(() => expect(uploadedName).toBe('receipt.png'))
    // The detail query refetches, so the newly linked row appears in the list.
    expect(await screen.findByRole('button', { name: 'Preview receipt.png' })).toBeInTheDocument()
  })

  it('shows an inline error when uploading an attachment fails', async () => {
    const user = userEvent.setup()
    server.use(http.post('*/api/v1/transactions/:id/attachments', () => new HttpResponse(null, { status: 500 })))

    const { container } = renderPage()
    await screen.findByText('Weekly shop')

    await user.click(screen.getByRole('button', { name: 'Add attachment' }))
    await user.click(await screen.findByRole('button', { name: 'Choose files' }))
    const fileInput = container.querySelectorAll('input[type="file"]')[1] as HTMLInputElement
    await user.upload(fileInput, new File(['bad'], 'broken.pdf', { type: 'application/pdf' }))

    expect(await screen.findByText(/couldn't upload that file/i)).toBeInTheDocument()
  })

  it('opens the big-preview lightbox for a clicked attachment and closes it again', async () => {
    const user = userEvent.setup()
    renderPage()

    await screen.findByText('Weekly shop')

    await user.click(screen.getByRole('button', { name: 'Preview attachment.jpg' }))

    const dialog = await screen.findByRole('dialog', { name: 'Preview of attachment.jpg' })
    expect(within(dialog).getByAltText('Preview of attachment.jpg')).toHaveAttribute(
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
    const numericAmount = { ...fixtureTransactions[0], amount: 4250, attachments: [] } as unknown as (typeof fixtureTransactions)[number]
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
        // The display comma is converted back to whole minor units (cents).
        amount: 4250,
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

  it('offers no extraction or re-analyze action on the read-only detail screen', async () => {
    // Regression (analysis §1 non-goals, §19): re-analyze/extraction live on the Add screen only.
    // A linked attachment keeps preview/download/delete and nothing else.
    const user = userEvent.setup()
    renderPage()
    await screen.findByText('Weekly shop')

    await user.click(screen.getByRole('button', { name: 'Actions for attachment.jpg' }))
    expect(await screen.findByRole('button', { name: 'Preview' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Download' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Delete' })).toBeInTheDocument()
    expect(
      screen.queryByRole('button', { name: /extract|prefill|re-?analy[sz]e|read the file/i }),
    ).not.toBeInTheDocument()
  })

  it('lists line items read-only on the detail screen and mounts the editable list in edit mode', async () => {
    const user = userEvent.setup()
    useStatefulItemisedTransaction()

    renderPage('3')
    await screen.findByText('Itemised shop')

    // Read-only: the persisted rows are listed, but the editing affordances are not.
    expect(screen.getByRole('heading', { name: 'Items' })).toBeInTheDocument()
    expect(screen.getByText('Milk 2%')).toBeInTheDocument()
    expect(screen.getByText('Sourdough bread')).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Add item' })).not.toBeInTheDocument()
    expect(screen.queryByRole('button', { name: /extract|prefill|re-?analy[sz]e/i })).not.toBeInTheDocument()

    // …and the editable list appears once the form opens.
    await user.click(screen.getByRole('button', { name: 'Edit' }))
    expect(await screen.findByRole('button', { name: 'Add item' })).toBeInTheDocument()
    expect(screen.getByText('Milk 2%')).toBeInTheDocument()
  })

  it('removes the transaction, its items, and its attachments after confirmation', async () => {
    const user = userEvent.setup()
    let deletedId: string | null = null
    server.use(
      http.get('*/api/v1/transactions/:id', () => HttpResponse.json({ data: cloneTransaction(fixtureItemisedTransaction) })),
      http.delete('*/api/v1/transactions/:id', ({ params }) => {
        deletedId = String(params.id)
        return new HttpResponse(null, { status: 204 })
      }),
    )

    renderPage('3')
    await screen.findByText('Itemised shop')

    await user.click(screen.getByRole('button', { name: 'Remove transaction' }))
    expect(await screen.findByText('Remove transaction?')).toBeInTheDocument()
    expect(screen.getByText(/all of its items, and all of its attachments/i)).toBeInTheDocument()

    await user.click(screen.getByRole('button', { name: 'Remove permanently' }))

    await waitFor(() => expect(deletedId).toBe('3'))
    // Success leaves the deleted detail screen for the refreshed transaction list.
    expect(await screen.findByText('Transactions list')).toBeInTheDocument()
  })

  it('keeps the confirmation open and shows an error when deleting the transaction fails', async () => {
    const user = userEvent.setup()
    server.use(
      http.get('*/api/v1/transactions/:id', () => HttpResponse.json({ data: cloneTransaction(fixtureItemisedTransaction) })),
      http.delete('*/api/v1/transactions/:id', () => new HttpResponse(null, { status: 500 })),
    )

    renderPage('3')
    await screen.findByText('Itemised shop')

    await user.click(screen.getByRole('button', { name: 'Remove transaction' }))
    await user.click(await screen.findByRole('button', { name: 'Remove permanently' }))

    expect(await screen.findByText(/couldn't remove the transaction/i)).toBeInTheDocument()
    expect(screen.getByText('Remove transaction?')).toBeInTheDocument()
  })

  describe('line items in edit mode', () => {
    it('adds a line item with a POST to the sub-transaction route and refetches the transaction', async () => {
      const user = userEvent.setup()
      const { state, calls } = useStatefulItemisedTransaction()

      renderPage('3')
      await screen.findByText('Itemised shop')
      await user.click(screen.getByRole('button', { name: 'Edit' }))

      // The derived Amount is read-only (no input) with the item-count note.
      expect(await screen.findByText('= sum of 2 items')).toBeInTheDocument()
      expect(screen.queryByLabelText('Amount')).not.toBeInTheDocument()

      await user.click(screen.getByRole('button', { name: 'Add item' }))
      await user.type(screen.getByLabelText('New item name'), 'Coffee')
      await user.type(screen.getByLabelText('New item amount'), '2,50')
      await user.click(screen.getByRole('button', { name: 'Add item' }))

      await waitFor(() =>
        expect(calls.post).toEqual([{ transactionId: '3', body: { name: 'Coffee', label: '', amount: 250 } }]),
      )

      // The mutation's invalidation refetches the detail query: the server's new row and Σ show up.
      expect(await screen.findByText('Coffee')).toBeInTheDocument()
      expect(screen.getByText('= sum of 3 items')).toBeInTheDocument()
      expect(screen.getByText('14,98')).toBeInTheDocument()
      expect(state.gets).toBeGreaterThan(1)
    })

    it('edits a line item with a PATCH to its own sub-transaction route and refetches', async () => {
      const user = userEvent.setup()
      const { calls } = useStatefulItemisedTransaction()

      renderPage('3')
      await screen.findByText('Itemised shop')
      await user.click(screen.getByRole('button', { name: 'Edit' }))

      await user.click(await screen.findByRole('button', { name: 'Edit Milk 2%' }))
      const nameInput = screen.getByLabelText('Item name for Milk 2%')
      const amountInput = screen.getByLabelText('Item amount for Milk 2%')
      await user.clear(nameInput)
      await user.type(nameInput, 'Milk 3%')
      await user.clear(amountInput)
      await user.type(amountInput, '7,00')
      await user.click(screen.getByRole('button', { name: 'Save' }))

      await waitFor(() =>
        expect(calls.patch).toEqual([
          {
            transactionId: '3',
            subTransactionId: 'sub-transaction-1',
            body: { name: 'Milk 3%', label: 'dairy', amount: 700 },
          },
        ]),
      )

      expect(await screen.findByText('Milk 3%')).toBeInTheDocument()
      expect(screen.getByText('= sum of 2 items')).toBeInTheDocument()
      expect(screen.getByText('12,99')).toBeInTheDocument()
    })

    it('removes line items with DELETE and unlocks a last-computed Amount when the list empties', async () => {
      const user = userEvent.setup()
      const { calls } = useStatefulItemisedTransaction()

      renderPage('3')
      await screen.findByText('Itemised shop')
      await user.click(screen.getByRole('button', { name: 'Edit' }))

      // Two items → derived read-only Amount showing their Σ, plus the count note.
      expect(await screen.findByText('= sum of 2 items')).toBeInTheDocument()
      expect(screen.queryByLabelText('Amount')).not.toBeInTheDocument()
      expect(screen.getByText('12,48')).toBeInTheDocument()

      await user.click(screen.getByRole('button', { name: 'Remove Milk 2%' }))

      // One item left: still derived, still read-only; the row is only dropped once the refetch lands.
      expect(await screen.findByText('= sum of 1 item')).toBeInTheDocument()
      expect(screen.queryByText('Milk 2%')).not.toBeInTheDocument()
      expect(screen.getByText('Sourdough bread')).toBeInTheDocument()

      await user.click(screen.getByRole('button', { name: 'Remove Sourdough bread' }))

      // Last item gone: the Amount is editable again and shows the server's *last computed* value
      // (5,99) — not the original 12,48 the form was seeded with (analysis §8).
      const amountInput = await screen.findByLabelText('Amount')
      expect(amountInput).toHaveValue('5,99')
      expect(amountInput).not.toHaveValue('12,48')
      expect(screen.queryByText('= sum of 1 item')).not.toBeInTheDocument()

      expect(calls.remove).toEqual([
        { transactionId: '3', subTransactionId: 'sub-transaction-1' },
        { transactionId: '3', subTransactionId: 'sub-transaction-2' },
      ])
    })

    it('shows an inline error and keeps the row when a mutation fails', async () => {
      const user = userEvent.setup()
      server.use(
        http.get('*/api/v1/transactions/:id', () => HttpResponse.json({ data: cloneTransaction(fixtureItemisedTransaction) })),
        // The parent is missing / not the caller's — the server 404s (analysis §8).
        http.delete(
          '*/api/v1/transactions/:transactionID/sub-transactions/:subTransactionID',
          () => new HttpResponse(null, { status: 404 }),
        ),
      )

      renderPage('3')
      await screen.findByText('Itemised shop')
      await user.click(screen.getByRole('button', { name: 'Edit' }))

      await user.click(await screen.findByRole('button', { name: 'Remove Milk 2%' }))

      expect(await screen.findByText(/couldn't remove that item/i)).toBeInTheDocument()
      // No optimistic removal: the row and the derived sum are untouched.
      expect(screen.getByText('Milk 2%')).toBeInTheDocument()
      expect(screen.getByText('= sum of 2 items')).toBeInTheDocument()
      expect(screen.getByText('Sourdough bread')).toBeInTheDocument()
    })
  })
})
