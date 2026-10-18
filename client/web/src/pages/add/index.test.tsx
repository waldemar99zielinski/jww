import { screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { http, HttpResponse } from 'msw'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { fixtureAttachment, fixturePrefill, fixtureUser } from '@/test/fixtures'
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
  await user.click(await screen.findByRole('button', { name: flow === 'prefill' ? /import & prefill/i : /just attach/i }))
  await user.click(await screen.findByRole('button', { name: 'Choose files' }))
  const fileInput = container.querySelectorAll('input[type="file"]')[1] as HTMLInputElement
  await user.upload(fileInput, files)
}

async function pickAttachment(
  user: ReturnType<typeof userEvent.setup>,
  container: HTMLElement,
  flow: 'attach' | 'prefill' = 'attach',
) {
  await pickAttachmentFiles(user, container, [new File(['content'], 'receipt.jpg', { type: 'image/jpeg' })], flow)
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
        // The comma is display-only; the payload carries whole minor units.
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

  describe('attachment upload', () => {
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

      expect(await screen.findByText('Attachment ready')).toBeInTheDocument()
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

      // Each pick is an independent draft: both banners show, both uploads happened.
      expect(await screen.findAllByText('Attachment ready')).toHaveLength(2)
      expect(uploadCount).toBe(2)

      await user.type(screen.getByLabelText('Transaction name'), 'Coffee')
      await user.type(screen.getByLabelText('Amount'), '25')
      await selectCategory(user, /groceries/i)
      await user.click(screen.getByRole('button', { name: /save transaction/i }))

      await waitFor(() => expect(createBody).toMatchObject({ attachment_ids: attachmentIds }))
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

      await waitFor(() => expect(screen.getAllByText('Attachment ready')).toHaveLength(2))
      expect(uploadCount).toBe(2)

      await user.type(screen.getByLabelText('Transaction name'), 'Coffee')
      await user.type(screen.getByLabelText('Amount'), '25')
      await selectCategory(user, /groceries/i)
      await user.click(screen.getByRole('button', { name: /save transaction/i }))

      await waitFor(() => expect(createBody).toMatchObject({ attachment_ids: attachmentIds }))
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
      await screen.findByText('Attachment ready')

      await user.click(screen.getByRole('button', { name: /remove attachment/i }))

      await waitFor(() => expect(deletedId).toBe(fixtureAttachment.id))
      expect(screen.queryByText('Attachment ready')).not.toBeInTheDocument()
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

      expect(await screen.findByText('Upload failed')).toBeInTheDocument()
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
      await screen.findByText('Attachment ready')
      expect(screen.getByRole('button', { name: /save transaction/i })).not.toBeDisabled()
    })

    it('reads the file and prefills untouched fields when the prefill flow is chosen', async () => {
      const user = userEvent.setup()
      let releaseRead: () => void = () => {}
      const readGate = new Promise<void>((resolve) => {
        releaseRead = resolve
      })
      server.use(
        http.post('*/api/v1/transaction-attachments', () => HttpResponse.json({ data: fixtureAttachment }, { status: 201 })),
        http.post('*/api/v1/transaction-attachments/:attachmentID/extraction', async () => {
          await readGate
          return HttpResponse.json({ data: fixturePrefill }, { status: 201 })
        }),
      )

      const { container } = renderPage()
      await pickAttachment(user, container, 'prefill')

      // A dedicated processing state shows while the read is in flight, and save is blocked.
      expect(await screen.findByText(/reading receipt\.jpg/i)).toBeInTheDocument()

      releaseRead()

      // The read lands: untouched fields fill from the file, category matched, save ready.
      expect(await screen.findByLabelText('Transaction name')).toHaveValue('Lidl groceries')
      expect(screen.getByLabelText('Amount')).toHaveValue('24,90')
      expect(screen.getByRole('button', { name: /groceries/i })).toBeInTheDocument()
      expect(screen.getByRole('button', { name: /save transaction/i })).not.toBeDisabled()
      await waitFor(() => expect(screen.queryByText(/reading receipt\.jpg/i)).not.toBeInTheDocument())
    })

    it('keeps the file attached when the read fails', async () => {
      const user = userEvent.setup()
      server.use(
        http.post('*/api/v1/transaction-attachments', () => HttpResponse.json({ data: fixtureAttachment }, { status: 201 })),
        http.post('*/api/v1/transaction-attachments/:attachmentID/extraction', () =>
          HttpResponse.json({ error: { message: 'read failed' } }, { status: 500 }),
        ),
      )

      const { container } = renderPage()
      await pickAttachment(user, container, 'prefill')

      expect(await screen.findByText(/couldn't read the file/i)).toBeInTheDocument()
      expect(screen.getByRole('button', { name: /remove attachment/i })).toBeInTheDocument()
    })
  })
})
