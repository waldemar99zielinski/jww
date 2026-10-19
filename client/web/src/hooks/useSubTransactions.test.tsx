import { QueryClientProvider } from '@tanstack/react-query'
import { renderHook, waitFor } from '@testing-library/react'
import type { ReactNode } from 'react'
import { describe, expect, it, vi } from 'vitest'
import { useCreateSubTransaction, useDeleteSubTransaction, useUpdateSubTransaction } from '@/hooks/useSubTransactions'
import { createTestQueryClient } from '@/test/renderWithProviders'

/**
 * A wrapper whose QueryClient's `invalidateQueries` is spied on, so a test can assert which query
 * keys a mutation refreshed without depending on cache internals.
 */
function setup() {
  const queryClient = createTestQueryClient()
  const invalidateSpy = vi.spyOn(queryClient, 'invalidateQueries')
  function wrapper({ children }: { children: ReactNode }) {
    return <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  }
  return { wrapper, invalidateSpy }
}

describe('useCreateSubTransaction', () => {
  it('invalidates the transaction and summary queries after adding a line item', async () => {
    const { wrapper, invalidateSpy } = setup()
    const { result } = renderHook(() => useCreateSubTransaction(), { wrapper })

    result.current.mutate({ transactionId: 3, input: { name: 'Coffee beans', label: 'pantry', amount: 850 } })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(invalidateSpy).toHaveBeenCalledWith({ queryKey: ['transactions'] })
    expect(invalidateSpy).toHaveBeenCalledWith({ queryKey: ['summary'] })
  })
})

describe('useUpdateSubTransaction', () => {
  it('invalidates the transaction and summary queries after editing a line item', async () => {
    const { wrapper, invalidateSpy } = setup()
    const { result } = renderHook(() => useUpdateSubTransaction(), { wrapper })

    result.current.mutate({ transactionId: 3, subTransactionId: 'sub-transaction-1', input: { amount: 700 } })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(invalidateSpy).toHaveBeenCalledWith({ queryKey: ['transactions'] })
    expect(invalidateSpy).toHaveBeenCalledWith({ queryKey: ['summary'] })
  })
})

describe('useDeleteSubTransaction', () => {
  it('invalidates the transaction and summary queries after removing a line item', async () => {
    const { wrapper, invalidateSpy } = setup()
    const { result } = renderHook(() => useDeleteSubTransaction(), { wrapper })

    result.current.mutate({ transactionId: 3, subTransactionId: 'sub-transaction-1' })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(invalidateSpy).toHaveBeenCalledWith({ queryKey: ['transactions'] })
    expect(invalidateSpy).toHaveBeenCalledWith({ queryKey: ['summary'] })
  })
})
