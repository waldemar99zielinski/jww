import { renderHook, waitFor } from '@testing-library/react'
import { QueryClientProvider } from '@tanstack/react-query'
import type { ReactNode } from 'react'
import { describe, expect, it } from 'vitest'
import { useCreateTransaction, useTransactions } from '@/hooks/useTransactions'
import { createTestQueryClient } from '@/test/renderWithProviders'
import { fixtureUser } from '@/test/fixtures'

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

  it('is disabled (never fetches) when user_id is not provided', () => {
    const { result } = renderHook(() => useTransactions({}), { wrapper })

    expect(result.current.fetchStatus).toBe('idle')
    expect(result.current.isPending).toBe(true)
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
