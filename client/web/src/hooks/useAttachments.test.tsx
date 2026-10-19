import { QueryClientProvider } from '@tanstack/react-query'
import { renderHook, waitFor } from '@testing-library/react'
import type { ReactNode } from 'react'
import { describe, expect, it } from 'vitest'
import { useDeleteDraftAttachment, useExtractTransactionPrefill, useUploadDraftAttachment, useUploadTransactionAttachment } from '@/hooks/useAttachments'
import { fixtureAttachment, fixturePrefill } from '@/test/fixtures'
import { createTestQueryClient } from '@/test/renderWithProviders'

function wrapper({ children }: { children: ReactNode }) {
  const queryClient = createTestQueryClient()
  return <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
}

describe('useUploadDraftAttachment', () => {
  it('resolves with the uploaded draft attachment', async () => {
    const { result } = renderHook(() => useUploadDraftAttachment(), { wrapper })

    result.current.mutate(new File(['content'], 'attachment.jpg', { type: 'image/jpeg' }))

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data).toEqual(fixtureAttachment)
  })
})

describe('useDeleteDraftAttachment', () => {
  it('resolves after deleting a draft', async () => {
    const { result } = renderHook(() => useDeleteDraftAttachment(), { wrapper })

    result.current.mutate(fixtureAttachment.id)

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
  })
})

describe('useUploadTransactionAttachment', () => {
  it('resolves with the attachment linked to the transaction', async () => {
    const { result } = renderHook(() => useUploadTransactionAttachment(), { wrapper })

    result.current.mutate({ transactionId: 1, file: new File(['content'], 'receipt.png', { type: 'image/png' }) })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data?.transaction_id).toBe(1)
  })
})

describe('useExtractTransactionPrefill', () => {
  it('resolves with the reader output for the uploaded draft', async () => {
    const { result } = renderHook(() => useExtractTransactionPrefill(), { wrapper })

    result.current.mutate({ attachmentIds: [fixtureAttachment.id] })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data).toEqual(fixturePrefill)
  })
})
