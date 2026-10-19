import { renderHook, waitFor } from '@testing-library/react'
import { QueryClientProvider } from '@tanstack/react-query'
import type { ReactNode } from 'react'
import { describe, expect, it } from 'vitest'
import { useCreateTransaction, useDeleteTransaction, useTransaction, useTransactions, useUpdateTransaction } from '@/hooks/useTransactions'
import { createTestQueryClient } from '@/test/renderWithProviders'
import { fixtureTransactions, fixtureUser } from '@/test/fixtures'

function wrapper({ children }: { children: ReactNode }) {
  const queryClient = createTestQueryClient()
  return <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
}

describe('useTransactions', () => {
  it('resolves with transactions matching the given filters', async () => {
    const { result } = renderHook(() => useTransactions({ user_id: fixtureUser.id, type: 'income' }), { wrapper })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))

    expect(result.current.data).toHaveLength(1)
    expect(result.current.data?.[0].name).toBe('October salary')
  })

  it('fetches every transaction when no user_id filter is provided', async () => {
    const { result } = renderHook(() => useTransactions({}), { wrapper })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))

    expect(result.current.data).toHaveLength(fixtureTransactions.length)
  })
})

describe('useTransaction', () => {
  it('resolves with the single transaction matching the id', async () => {
    const { result } = renderHook(() => useTransaction(1), { wrapper })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))

    expect(result.current.data?.name).toBe('Weekly shop')
    expect(result.current.data?.attachments).toHaveLength(2)
    expect(result.current.data?.attachments[0].preview_grant).toBe('grant-attachment-1')
    expect(result.current.data?.attachments[1].preview_grant).toBeNull()
  })

  it('is disabled (never fetches) when the id is not a whole number', () => {
    const { result } = renderHook(() => useTransaction(Number.NaN), { wrapper })

    expect(result.current.fetchStatus).toBe('idle')
  })
})

describe('useCreateTransaction', () => {
  it('resolves with the created transaction and invalidates transactions + summary queries', async () => {
    const queryClient = createTestQueryClient()
    function localWrapper({ children }: { children: ReactNode }) {
      return <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
    }

    const { result } = renderHook(() => useCreateTransaction(), { wrapper: localWrapper })

    result.current.mutate({
      type: 'expense',
      amount: 12.5,
      currency_code: 'EUR',
      category_id: 'cat-groceries',
      user_id: fixtureUser.id,
      source: 'manual',
      name: 'Coffee',
    })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data?.name).toBe('Coffee')
    expect(result.current.data?.category_name).toBe('Groceries')
  })
})

describe('useUpdateTransaction', () => {
  it('patches the transaction and returns the detail shape', async () => {
    const { result } = renderHook(() => useUpdateTransaction(), { wrapper })

    result.current.mutate({ id: 1, input: { name: 'Renamed', amount: 4200 } })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data?.id).toBe(1)
    expect(result.current.data?.name).toBe('Renamed')
  })
})

describe('useDeleteTransaction', () => {
  it('deletes the transaction and drops its detail query from the cache', async () => {
    const queryClient = createTestQueryClient()
    function localWrapper({ children }: { children: ReactNode }) {
      return <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
    }
    queryClient.setQueryData(['transactions', 'detail', 1], fixtureTransactions[0])

    const { result } = renderHook(() => useDeleteTransaction(), { wrapper: localWrapper })

    result.current.mutate(1)

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(queryClient.getQueryData(['transactions', 'detail', 1])).toBeUndefined()
  })
})
