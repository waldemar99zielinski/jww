import { screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { http, HttpResponse } from 'msw'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import type { TransactionPrefillDto } from '@/api/types'
import { fixtureAttachment, fixturePrefill, fixturePrefillWithSubTransactions, fixtureUser } from '@/test/fixtures'
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

async function pickAttachmentFiles(
  user: ReturnType<typeof userEvent.setup>,
  container: HTMLElement,
  files: File[],
  flow: 'attach' | 'prefill' = 'attach',
) {
  await user.click(await screen.findByRole('button', { name: 'Add attachment' }))
  await user.click(await screen.findByRole('button', { name: 'Choose files' }))
  const fileInput = container.querySelectorAll('input[type="file"]')[1] as HTMLInputElement
  await user.upload(fileInput, files)
  if (flow === 'prefill') {
    await user.click(await screen.findByRole('button', { name: 'Scan attachment' }))
  }
}

async function pickAttachment(
  user: ReturnType<typeof userEvent.setup>,
  container: HTMLElement,
  flow: 'attach' | 'prefill' = 'attach',
) {
  await pickAttachmentFiles(user, container, [new File(['content'], 'attachment.jpg', { type: 'image/jpeg' })], flow)
}

describe('AddTransactionPage', () => {
  beforeEach(() => {
    // jsdom doesn't implement these; the attachment-picking flow calls them.
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
    await user.type(screen.getByLabelText('Amount'), '25,99')
    await selectCategory(user, /groceries/i)

    await user.click(screen.getByRole('button', { name: /save transaction/i }))

    await waitFor(() =>
      expect(capturedBody).toMatchObject({
        type: 'expense',
        // The comma is display-only; the wire carries whole cents.
        amount: 2599,
        currency_code: 'EUR',
        category_id: 'cat-groceries',
        user_id: fixtureUser.id,
        source: 'manual',
        name: 'Coffee',
        // The picker's date/time is sent as an RFC3339 instant.
        created_at: expect.any(String),
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

  describe('attachments and extraction', () => {
    it('uploads the picked file as a draft and links it to the transaction on save', async () => {
      const user = userEvent.setup()
      let uploaded = false
      let createBody: Record<string, unknown> | null = null
      server.use(
        http.post('*/api/v1/transaction-attachments', () => {
          uploaded = true
          return HttpResponse.json({ data: fixtureAttachment }, { status: 201 })
        }),
        http.post('*/api/v1/transactions', async ({ request }) => {
          createBody = (await request.json()) as Record<string, unknown>
          return HttpResponse.json({
            data: {
              id: 1,
              category_name: 'Groceries',
              created_at: '2026-10-10T00:00:00Z',
              description: '',
              attachments: [],
              ...createBody,
            },
          })
        }),
      )

      const { container } = renderPage()
      await pickAttachment(user, container)

      expect(await screen.findByText('1 attachment')).toBeInTheDocument()
      expect(screen.getByText('Ready to scan.')).toBeInTheDocument()
      // The file list is tucked behind the details disclosure, so the top of the page stays compact.
      expect(screen.queryByText('attachment.jpg')).not.toBeInTheDocument()
      expect(uploaded).toBe(true)

      await user.type(screen.getByLabelText('Transaction name'), 'Coffee')
      await user.type(screen.getByLabelText('Amount'), '25')
      await selectCategory(user, /groceries/i)
      await user.click(screen.getByRole('button', { name: /save transaction/i }))

      await waitFor(() => expect(createBody).toMatchObject({ attachment_ids: [fixtureAttachment.id] }))
    })

    it('uploads every picked file and links them all to the transaction on save', async () => {
      const user = userEvent.setup()
      const attachmentIds = ['attachment-1', 'attachment-2']
      let uploadCount = 0
      let createBody: Record<string, unknown> | null = null
      server.use(
        http.post('*/api/v1/transaction-attachments', () => {
          const id = attachmentIds[uploadCount++]
          return HttpResponse.json({ data: { ...fixtureAttachment, id } }, { status: 201 })
        }),
        http.post('*/api/v1/transactions', async ({ request }) => {
          createBody = (await request.json()) as Record<string, unknown>
          return HttpResponse.json({
            data: {
              id: 1,
              category_name: 'Groceries',
              created_at: '2026-10-10T00:00:00Z',
              description: '',
              attachments: [],
              ...createBody,
            },
          })
        }),
      )

      const { container } = renderPage()
      await pickAttachment(user, container)
      await pickAttachment(user, container)

      expect(await screen.findByText('2 attachments')).toBeInTheDocument()
      expect(uploadCount).toBe(2)

      await user.type(screen.getByLabelText('Transaction name'), 'Coffee')
      await user.type(screen.getByLabelText('Amount'), '25')
      await selectCategory(user, /groceries/i)
      await user.click(screen.getByRole('button', { name: /save transaction/i }))

      await waitFor(() =>
        expect(createBody).toMatchObject({ attachment_ids: expect.arrayContaining(attachmentIds) }),
      )
      expect((createBody as unknown as { attachment_ids: string[] }).attachment_ids).toHaveLength(attachmentIds.length)
    })

    it('uploads every file from a single multi-select and links them all on save', async () => {
      const user = userEvent.setup()
      const attachmentIds = ['attachment-1', 'attachment-2']
      let uploadCount = 0
      let createBody: Record<string, unknown> | null = null
      server.use(
        http.post('*/api/v1/transaction-attachments', () => {
          const id = attachmentIds[uploadCount++]
          return HttpResponse.json({ data: { ...fixtureAttachment, id } }, { status: 201 })
        }),
        http.post('*/api/v1/transactions', async ({ request }) => {
          createBody = (await request.json()) as Record<string, unknown>
          return HttpResponse.json({
            data: {
              id: 1,
              category_name: 'Groceries',
              created_at: '2026-10-10T00:00:00Z',
              description: '',
              attachments: [],
              ...createBody,
            },
          })
        }),
      )

      const { container } = renderPage()
      await pickAttachmentFiles(user, container, [
        new File(['a'], 'one.jpg', { type: 'image/jpeg' }),
        new File(['b'], 'two.pdf', { type: 'application/pdf' }),
      ])

      await waitFor(() => expect(screen.getByText('2 attachments')).toBeInTheDocument())
      expect(uploadCount).toBe(2)

      await user.type(screen.getByLabelText('Transaction name'), 'Coffee')
      await user.type(screen.getByLabelText('Amount'), '25')
      await selectCategory(user, /groceries/i)
      await user.click(screen.getByRole('button', { name: /save transaction/i }))

      await waitFor(() =>
        expect(createBody).toMatchObject({ attachment_ids: expect.arrayContaining(attachmentIds) }),
      )
      expect((createBody as unknown as { attachment_ids: string[] }).attachment_ids).toHaveLength(attachmentIds.length)
    })

    it('extracts multiple uploaded images in one combined model request', async () => {
      const user = userEvent.setup()
      const attachmentIds = ['attachment-1', 'attachment-2']
      let uploadCount = 0
      let extractionCount = 0
      let extractedAttachmentIds: string[] = []
      server.use(
        http.post('*/api/v1/transaction-attachments', () => {
          const id = attachmentIds[uploadCount++]
          return HttpResponse.json({ data: { ...fixtureAttachment, id } }, { status: 201 })
        }),
        http.post('*/api/v1/transaction-attachments/extraction', async ({ request }) => {
          extractionCount += 1
          const body = (await request.json()) as { attachment_ids: string[] }
          extractedAttachmentIds = body.attachment_ids
          return HttpResponse.json({
            data: {
              attachment_id: body.attachment_ids[0],
              attachment_ids: body.attachment_ids,
              name: { value: 'Combined purchase' },
              sub_transactions: [],
            },
          })
        }),
      )

      const { container } = renderPage()
      await pickAttachmentFiles(user, container, [
        new File(['a'], 'front.jpg', { type: 'image/jpeg' }),
        new File(['b'], 'back.jpg', { type: 'image/jpeg' }),
      ])
      await waitFor(() => expect(screen.getByText('2 attachments')).toBeInTheDocument())

      await user.click(screen.getByRole('button', { name: 'Scan attachment' }))

      expect(await screen.findByText('Attachment read')).toBeInTheDocument()
      expect(screen.getByLabelText('Transaction name')).toHaveValue('Combined purchase')
      expect(extractionCount).toBe(1)
      expect(extractedAttachmentIds).toEqual(attachmentIds)
    })

    it('opens the shared attachment preview from an uploaded file row', async () => {
      const user = userEvent.setup()
      server.use(
        http.post('*/api/v1/transaction-attachments', () =>
          HttpResponse.json({ data: fixtureAttachment }, { status: 201 }),
        ),
      )

      const { container } = renderPage()
      await pickAttachment(user, container)
      await screen.findByText('1 attachment')

      await user.click(screen.getByRole('button', { name: 'Show attachment details' }))
      await user.click(screen.getByRole('button', { name: 'Preview attachment.jpg' }))

      expect(screen.getByRole('dialog', { name: 'Preview of attachment.jpg' })).toBeInTheDocument()
      expect(screen.getByAltText('Preview of attachment.jpg')).toHaveAttribute('src', 'blob:fixture')

      await user.click(screen.getByRole('button', { name: 'Close preview' }))
      expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
    })

    it('deletes the draft when the attachment is removed before saving', async () => {
      const user = userEvent.setup()
      let deletedId: string | null = null
      server.use(
        http.post('*/api/v1/transaction-attachments', () =>
          HttpResponse.json({ data: fixtureAttachment }, { status: 201 }),
        ),
        http.delete('*/api/v1/transaction-attachments/:attachmentID', ({ params }) => {
          deletedId = String(params.attachmentID)
          return new HttpResponse(null, { status: 204 })
        }),
      )

      const { container } = renderPage()
      await pickAttachment(user, container)
      expect(await screen.findByText('1 attachment')).toBeInTheDocument()

      await user.click(screen.getByRole('button', { name: 'Show attachment details' }))
      await user.click(screen.getByRole('button', { name: /remove attachment/i }))

      await waitFor(() => expect(deletedId).toBe(fixtureAttachment.id))
      expect(screen.queryByRole('region', { name: 'Attachment' })).not.toBeInTheDocument()
      expect(screen.queryByText('attachment.jpg')).not.toBeInTheDocument()
    })

    it('shows an error state when the upload fails', async () => {
      const user = userEvent.setup()
      server.use(
        http.post('*/api/v1/transaction-attachments', () =>
          HttpResponse.json({ error: { message: 'upload failed' } }, { status: 500 }),
        ),
      )

      const { container } = renderPage()
      await pickAttachment(user, container)

      expect(await screen.findByText('1 attachment')).toBeInTheDocument()
      await user.click(screen.getByRole('button', { name: 'Show attachment details' }))
      expect(await screen.findByText(/upload failed/i)).toBeInTheDocument()
    })

    it('keeps saving disabled while the attachment is still uploading', async () => {
      const user = userEvent.setup()
      let releaseUpload: () => void = () => {}
      const gate = new Promise<void>((resolve) => {
        releaseUpload = resolve
      })
      server.use(
        http.post('*/api/v1/transaction-attachments', async () => {
          await gate
          return HttpResponse.json({ data: fixtureAttachment }, { status: 201 })
        }),
      )

      const { container } = renderPage()
      await user.type(await screen.findByLabelText('Transaction name'), 'Coffee')
      await user.type(screen.getByLabelText('Amount'), '25')
      await selectCategory(user, /groceries/i)
      expect(screen.getByRole('button', { name: /save transaction/i })).not.toBeDisabled()

      await pickAttachment(user, container)
      expect(screen.getByRole('button', { name: /save transaction/i })).toBeDisabled()

      releaseUpload()
      expect(await screen.findByText('1 attachment')).toBeInTheDocument()
      expect(screen.getByRole('button', { name: /save transaction/i })).not.toBeDisabled()
    })
  })

  describe('successful extraction', () => {
    it('fills every extracted field and preloads all line items on the main page', async () => {
      const user = userEvent.setup()
      let createBody: Record<string, unknown> | null = null
      server.use(
        http.post('*/api/v1/transaction-attachments/extraction', () =>
          HttpResponse.json({ data: fixturePrefillWithSubTransactions }, { status: 200 }),
        ),
        http.post('*/api/v1/transactions', async ({ request }) => {
          createBody = (await request.json()) as Record<string, unknown>
          return HttpResponse.json({
            data: { id: 1, category_name: 'Groceries', created_at: '2026-10-10T00:00:00Z', description: '', ...createBody },
          })
        }),
      )

      const { container } = renderPage()
      await pickAttachment(user, container, 'prefill')

      expect(await screen.findByText('2 items added')).toBeInTheDocument()
      expect(screen.getByLabelText('Transaction name')).toHaveValue('Lidl groceries')
      // The category hint matched the loaded list without any extra action.
      expect(screen.getByRole('button', { name: /groceries/i })).toBeInTheDocument()
      // The extracted time is applied to the date/time fields.
      await user.click(screen.getByRole('button', { name: 'Today' }))
      expect(screen.getByLabelText('Time')).toHaveValue('14:35')
      // The attachment amount is replaced by the item sum, shown read-only.
      expect(screen.queryByLabelText('Amount')).not.toBeInTheDocument()
      expect(screen.getByText('= sum of 2 items')).toBeInTheDocument()
      const itemsCard = screen.getByRole('region', { name: 'Sub-transactions' })
      // The panel starts collapsed with a summary; rows appear only when expanded.
      expect(within(itemsCard).getByText('2 items · 12,48')).toBeInTheDocument()
      expect(within(itemsCard).queryByText('Milk 2%')).not.toBeInTheDocument()
      await user.click(within(itemsCard).getByRole('button', { name: 'Expand sub-transactions' }))
      expect(within(itemsCard).getByText('Milk 2%')).toBeInTheDocument()
      expect(within(itemsCard).getByText('Sourdough bread')).toBeInTheDocument()
      expect(within(itemsCard).getByText('12,48')).toBeInTheDocument()

      await user.click(screen.getByRole('button', { name: /save transaction/i }))

      await waitFor(() =>
        expect(createBody).toMatchObject({
          amount: 1248,
          name: 'Lidl groceries',
          category_id: 'cat-groceries',
          sub_transactions: [
            { name: 'Milk 2%', label: 'dairy', amount: 649 },
            { name: 'Sourdough bread', label: 'bakery', amount: 599 },
          ],
        }),
      )
    })

    it('outlines every extracted field scored below 80 until the user reviews it', async () => {
      const user = userEvent.setup()
      const lowConfidencePrefill: TransactionPrefillDto = {
        attachment_id: fixtureAttachment.id,
        attachment_ids: [fixtureAttachment.id],
        type: { value: 'expense', confidence: 10 },
        amount: { value: 2490, confidence: 20 },
        currency_code: { value: 'EUR', confidence: 30 },
        occurred_on: { value: '2026-10-10', confidence: 40 },
        occurred_at: { value: '14:35', confidence: 50 },
        name: { value: 'Low confidence shop', confidence: 60 },
        description: { value: 'Needs review', confidence: 70 },
        category_hint: { value: 'Groceries', confidence: 75 },
        sub_transactions: [],
      }
      server.use(
        http.post('*/api/v1/transaction-attachments/extraction', () =>
          HttpResponse.json({ data: lowConfidencePrefill }, { status: 200 }),
        ),
      )

      const { container } = renderPage()
      await pickAttachment(user, container, 'prefill')
      await screen.findByText('Attachment read')

      // Each extracted scalar carries the review outline…
      expect(screen.getByRole('group', { name: 'Transaction type' })).toHaveAttribute('aria-invalid', 'true')
      expect(screen.getByLabelText('Amount')).toHaveAttribute('aria-invalid', 'true')
      expect(screen.getByLabelText('Currency')).toHaveAttribute('aria-invalid', 'true')
      expect(screen.getByLabelText('Transaction name')).toHaveAttribute('aria-invalid', 'true')
      expect(screen.getByLabelText('Transaction description')).toHaveAttribute('aria-invalid', 'true')
      expect(await screen.findByRole('button', { name: /groceries/i })).toHaveAttribute('aria-invalid', 'true')
      expect(screen.getByRole('button', { name: /10 Oct/ })).toHaveAttribute('aria-invalid', 'true')
      expect(screen.getByLabelText('Transaction name')).toHaveClass('ring-critical')

      // Opening the date fields reveals the separately-scored date and time outlines.
      await user.click(screen.getByRole('button', { name: /10 Oct/ }))
      expect(screen.getByLabelText('Date')).toHaveAttribute('aria-invalid', 'true')
      expect(screen.getByLabelText('Time')).toHaveAttribute('aria-invalid', 'true')

      // Reviewing a field retires only that field's outline.
      const nameInput = screen.getByLabelText('Transaction name')
      await user.clear(nameInput)
      await user.type(nameInput, 'Reviewed shop')
      expect(nameInput).not.toHaveAttribute('aria-invalid')

      // "Use today" is an explicit replacement for both date fields.
      await user.click(screen.getByRole('button', { name: 'Use today' }))
      expect(screen.getByLabelText('Date')).not.toHaveAttribute('aria-invalid')
      expect(screen.getByLabelText('Time')).not.toHaveAttribute('aria-invalid')
    })

    it('does not outline a field scored exactly at the threshold', async () => {
      const user = userEvent.setup()
      server.use(
        http.post('*/api/v1/transaction-attachments/extraction', () =>
          HttpResponse.json({ data: fixturePrefill }, { status: 200 }),
        ),
      )

      const { container } = renderPage()
      await pickAttachment(user, container, 'prefill')
      await screen.findByText('Attachment read')

      // `name` is scored exactly 80 in the fixture; only strictly lower scores are outlined.
      expect(screen.getByLabelText('Transaction name')).not.toHaveAttribute('aria-invalid')
    })

    it('outlines extracted line items scored below the threshold until they are reviewed', async () => {
      const user = userEvent.setup()
      server.use(
        http.post('*/api/v1/transaction-attachments/extraction', () =>
          HttpResponse.json({
            data: {
              ...fixturePrefill,
              sub_transactions: [
                { name: 'Milk 2%', label: 'dairy', amount: 649, confidence: 40 },
                { name: 'Sourdough bread', label: 'bakery', amount: 599, confidence: 90 },
              ],
            },
          }),
        ),
      )

      const { container } = renderPage()
      await pickAttachment(user, container, 'prefill')
      await screen.findByText('2 items added')

      const panel = screen.getByRole('region', { name: 'Sub-transactions' })
      await user.click(within(panel).getByRole('button', { name: 'Expand sub-transactions' }))

      const lowItem = within(panel).getByRole('button', { name: 'Edit Milk 2%' })
      expect(lowItem).toHaveAttribute('aria-invalid', 'true')
      expect(lowItem).toHaveClass('ring-critical')
      expect(within(panel).getByRole('button', { name: 'Edit Sourdough bread' })).not.toHaveAttribute(
        'aria-invalid',
      )

      // Saving an edited row counts as reviewing it.
      await user.click(lowItem)
      await user.click(screen.getByRole('button', { name: 'Save' }))

      expect(within(panel).getByRole('button', { name: 'Edit Milk 2%' })).not.toHaveAttribute('aria-invalid')
    })

    it('resolves a category hint after a proposed type change loads the new category list', async () => {
      const user = userEvent.setup()
      const incomePrefill: TransactionPrefillDto = {
        attachment_id: fixtureAttachment.id,
        attachment_ids: [fixtureAttachment.id],
        type: { value: 'income' },
        amount: { value: 250000 },
        name: { value: 'October salary' },
        category_hint: { value: 'Salary' },
        sub_transactions: [],
      }
      server.use(
        http.post('*/api/v1/transaction-attachments/extraction', () =>
          HttpResponse.json({ data: incomePrefill }, { status: 200 }),
        ),
      )

      const { container } = renderPage()
      await pickAttachment(user, container, 'prefill')

      expect(await screen.findByText('Attachment read')).toBeInTheDocument()
      expect(screen.getByLabelText('Transaction name')).toHaveValue('October salary')
      // The income query resolves after the expense list was already mounted.
      expect(await screen.findByRole('button', { name: /salary/i })).toBeInTheDocument()
      expect(screen.queryByText(/isn't in your list/i)).not.toBeInTheDocument()
    })

    it('keeps the extracted values and items when the attachment is removed', async () => {
      const user = userEvent.setup()
      server.use(
        http.post('*/api/v1/transaction-attachments/extraction', () =>
          HttpResponse.json({ data: fixturePrefillWithSubTransactions }, { status: 200 }),
        ),
      )

      const { container } = renderPage()
      await pickAttachment(user, container, 'prefill')
      await screen.findByText('2 items added')

      await user.click(screen.getByRole('button', { name: 'Show attachment details' }))
      await user.click(screen.getByRole('button', { name: 'Remove attachment attachment.jpg' }))

      await waitFor(() => expect(screen.queryByText('2 items added')).not.toBeInTheDocument())
      expect(screen.getByLabelText('Transaction name')).toHaveValue('Lidl groceries')
      const itemsCard = screen.getByRole('region', { name: 'Sub-transactions' })
      expect(within(itemsCard).getByText('2 items · 12,48')).toBeInTheDocument()
      await user.click(within(itemsCard).getByRole('button', { name: 'Expand sub-transactions' }))
      expect(within(itemsCard).getByText('Milk 2%')).toBeInTheDocument()
      expect(screen.getByText('= sum of 2 items')).toBeInTheDocument()
    })

    it('undoes a scan and restores the form that existed before it', async () => {
      const user = userEvent.setup()
      server.use(
        http.post('*/api/v1/transaction-attachments/extraction', () =>
          HttpResponse.json({ data: fixturePrefillWithSubTransactions }, { status: 200 }),
        ),
      )

      const { container } = renderPage()
      await user.type(screen.getByLabelText('Transaction name'), 'Typed by hand')
      await user.type(screen.getByLabelText('Amount'), '30')
      await selectCategory(user, /transport/i)

      await pickAttachment(user, container, 'prefill')
      await screen.findByText('2 items added')
      expect(screen.getByLabelText('Transaction name')).toHaveValue('Lidl groceries')

      await user.click(screen.getByRole('button', { name: 'Show attachment details' }))
      await user.click(screen.getByRole('button', { name: 'Undo scan' }))

      expect(screen.getByLabelText('Transaction name')).toHaveValue('Typed by hand')
      expect(screen.getByLabelText('Amount')).toHaveValue('30,00')
      expect(screen.getByRole('button', { name: /transport/i })).toBeInTheDocument()
      expect(screen.queryByText('Milk 2%')).not.toBeInTheDocument()
      expect(screen.getByText('1 attachment')).toBeInTheDocument()
    })

    it('keeps the file attached when the read fails and retries without re-uploading', async () => {
      const user = userEvent.setup()
      let extractionCount = 0
      server.use(
        http.post('*/api/v1/transaction-attachments/extraction', () => {
          if (extractionCount++ === 0) {
            return HttpResponse.json({ error: { message: 'read failed' } }, { status: 500 })
          }
          return HttpResponse.json({ data: fixturePrefill }, { status: 200 })
        }),
      )

      const { container } = renderPage()
      await pickAttachment(user, container, 'prefill')

      expect(await screen.findByText("Couldn't read attachment")).toBeInTheDocument()
      await user.click(screen.getByRole('button', { name: 'Show attachment details' }))
      expect(screen.getByRole('button', { name: 'Remove attachment attachment.jpg' })).toBeInTheDocument()
      expect(screen.getByRole('button', { name: 'Try again' })).toBeInTheDocument()

      await user.click(screen.getByRole('button', { name: 'Try again' }))

      expect(await screen.findByText('Attachment read')).toBeInTheDocument()
      expect(screen.getByLabelText('Transaction name')).toHaveValue('Lidl groceries')
      expect(screen.queryByText("Couldn't read attachment")).not.toBeInTheDocument()
    })

    it('replaces extracted rows on a re-scan but keeps rows added by hand', async () => {
      const user = userEvent.setup()
      let extractionCount = 0
      const readings: TransactionPrefillDto[] = [
        fixturePrefillWithSubTransactions,
        {
          attachment_id: fixtureAttachment.id,
          attachment_ids: [fixtureAttachment.id],
          name: { value: 'Netto' },
          sub_transactions: [{ name: 'Coffee', label: 'pantry', amount: 300 }],
        },
      ]
      server.use(
        http.post('*/api/v1/transaction-attachments/extraction', () =>
          HttpResponse.json({ data: readings[extractionCount++] }, { status: 200 }),
        ),
      )

      const { container } = renderPage()
      await pickAttachment(user, container, 'prefill')
      await screen.findByText('2 items added')

      // Add a manual row through the panel's modal.
      const itemsPanel = screen.getByRole('region', { name: 'Sub-transactions' })
      await user.click(within(itemsPanel).getByRole('button', { name: 'Expand sub-transactions' }))
      await user.click(within(itemsPanel).getByRole('button', { name: 'Add sub-transaction' }))
      await user.type(screen.getByLabelText('Sub-transaction name'), 'Manual chocolate')
      await user.type(screen.getByLabelText('Sub-transaction amount'), '2,50')
      await user.click(screen.getByRole('button', { name: 'Add item' }))

      await user.click(screen.getByRole('button', { name: 'Scan again' }))

      expect(await screen.findByText('1 item added')).toBeInTheDocument()
      expect(screen.getByLabelText('Transaction name')).toHaveValue('Netto')
      const itemsCard = screen.getByRole('region', { name: 'Sub-transactions' })
      expect(within(itemsCard).getByText('Coffee')).toBeInTheDocument()
      expect(within(itemsCard).getByText('Manual chocolate')).toBeInTheDocument()
      expect(screen.queryByText('Milk 2%')).not.toBeInTheDocument()
      expect(screen.queryByText('Sourdough bread')).not.toBeInTheDocument()
      expect(screen.getByText('= sum of 2 items')).toBeInTheDocument()
    })
  })

  describe('sub-transactions', () => {
    it('adds and removes items through the panel and modal, updating the derived amount', async () => {
      const user = userEvent.setup()
      let createBody: Record<string, unknown> | null = null
      server.use(
        http.post('*/api/v1/transactions', async ({ request }) => {
          createBody = (await request.json()) as Record<string, unknown>
          return HttpResponse.json({
            data: { id: 1, category_name: 'Groceries', created_at: '2026-10-10T00:00:00Z', description: '', ...createBody },
          })
        }),
      )

      renderPage()
      const panel = screen.getByRole('region', { name: 'Sub-transactions' })
      expect(within(panel).getByText('No sub-transactions')).toBeInTheDocument()

      await user.click(within(panel).getByRole('button', { name: 'Add sub-transaction' }))
      expect(await screen.findByRole('dialog', { name: 'Add sub-transaction' })).toBeInTheDocument()

      await user.type(screen.getByLabelText('Sub-transaction name'), 'Coffee beans')
      await user.type(screen.getByLabelText('Sub-transaction label'), 'pantry')
      await user.type(screen.getByLabelText('Sub-transaction amount'), '8,50')
      await user.click(screen.getByRole('button', { name: 'Add item' }))

      expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
      expect(within(panel).getByText('Coffee beans')).toBeInTheDocument()
      expect(screen.getByText('= sum of 1 item')).toBeInTheDocument()
      expect(screen.queryByLabelText('Amount')).not.toBeInTheDocument()

      // The row's trash button removes it immediately, releasing the typed Amount again.
      await user.click(within(panel).getByRole('button', { name: 'Remove Coffee beans' }))

      expect(screen.queryByText('Coffee beans')).not.toBeInTheDocument()
      expect(screen.getByLabelText('Amount')).toBeInTheDocument()

      // Finish the form and prove the item payload only contains what survived.
      await user.type(screen.getByLabelText('Transaction name'), 'Shopping')
      await user.type(screen.getByLabelText('Amount'), '10')
      await selectCategory(user, /groceries/i)
      await user.click(screen.getByRole('button', { name: /save transaction/i }))

      await waitFor(() => expect(createBody).toMatchObject({ amount: 1000, name: 'Shopping' }))
      expect((createBody as unknown as Record<string, unknown>).sub_transactions).toBeUndefined()
    })

    it('starts collapsed, then expands and edits one row in a modal', async () => {
      const user = userEvent.setup()
      server.use(
        http.post('*/api/v1/transaction-attachments/extraction', () =>
          HttpResponse.json({ data: fixturePrefillWithSubTransactions }, { status: 200 }),
        ),
      )

      const { container } = renderPage()
      await pickAttachment(user, container, 'prefill')
      await screen.findByText('2 items added')

      const panel = screen.getByRole('region', { name: 'Sub-transactions' })
      expect(within(panel).getByText('2 items · 12,48')).toBeInTheDocument()
      expect(within(panel).queryByText('Milk 2%')).not.toBeInTheDocument()

      await user.click(within(panel).getByRole('button', { name: 'Expand sub-transactions' }))
      expect(within(panel).getByText('Milk 2%')).toBeInTheDocument()

      await user.click(within(panel).getByRole('button', { name: 'Edit Milk 2%' }))
      const dialog = await screen.findByRole('dialog', { name: 'Edit sub-transaction' })
      expect(within(dialog).getByLabelText('Sub-transaction name')).toHaveValue('Milk 2%')
      expect(within(dialog).getByLabelText('Sub-transaction label')).toHaveValue('dairy')
      expect(within(dialog).getByLabelText('Sub-transaction amount')).toHaveValue('6,49')

      await user.clear(within(dialog).getByLabelText('Sub-transaction amount'))
      await user.type(within(dialog).getByLabelText('Sub-transaction amount'), '7,00')
      await user.click(within(dialog).getByRole('button', { name: 'Save' }))

      expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
      expect(within(panel).getByText('7,00')).toBeInTheDocument()
      expect(within(panel).getByText('2 items · 12,99')).toBeInTheDocument()
      expect(screen.getByText('= sum of 2 items')).toBeInTheDocument()
    })

    it('shows the attachment-total difference in the collapsed summary', async () => {
      const user = userEvent.setup()
      server.use(
        http.post('*/api/v1/transaction-attachments/extraction', () =>
          HttpResponse.json({ data: fixturePrefillWithSubTransactions }, { status: 200 }),
        ),
      )

      const { container } = renderPage()
      await pickAttachment(user, container, 'prefill')
      await screen.findByText('2 items added')

      const panel = screen.getByRole('region', { name: 'Sub-transactions' })
      expect(within(panel).getByText(/attachment total.*difference/i)).toBeInTheDocument()

      await user.click(within(panel).getByRole('button', { name: 'Expand sub-transactions' }))
      await user.click(within(panel).getByRole('button', { name: 'Remove Sourdough bread' }))
      await user.click(within(panel).getByRole('button', { name: 'Collapse sub-transactions' }))

      expect(within(panel).getByText('1 item · 6,49')).toBeInTheDocument()
      expect(within(panel).getByText(/attachment total.*difference/i)).toBeInTheDocument()
    })
  })
})
