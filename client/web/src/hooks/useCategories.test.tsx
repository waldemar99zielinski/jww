import { renderHook, waitFor } from '@testing-library/react'
import type { ReactNode } from 'react'
import { QueryClientProvider } from '@tanstack/react-query'
import { describe, expect, it } from 'vitest'
import { useCategories, useCreateCategory, useUpdateCategory } from '@/hooks/useCategories'
import { createTestQueryClient } from '@/test/renderWithProviders'

function wrapper({ children }: { children: ReactNode }) {
  const queryClient = createTestQueryClient()
  return <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
}

describe('useCategories', () => {
  it('resolves with the active categories for the given type', async () => {
    const { result } = renderHook(() => useCategories('expense'), { wrapper })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))

    expect(result.current.data?.map((c) => c.name)).toEqual(['Groceries', 'Transport', 'Rent'])
  })
})

describe('useCreateCategory', () => {
  it('invalidates the categories query for the created category\'s type on success', async () => {
    const queryClient = createTestQueryClient()
    function localWrapper({ children }: { children: ReactNode }) {
      return <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
    }

    const { result } = renderHook(() => useCreateCategory(), { wrapper: localWrapper })

    result.current.mutate({ type: 'expense', name: 'Pets' })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data?.name).toBe('Pets')
  })
})

describe('useUpdateCategory', () => {
  it('resolves with the updated category', async () => {
    const { result } = renderHook(() => useUpdateCategory(), { wrapper })

    result.current.mutate({ id: 'cat-groceries', name: 'Groceries', active: false, sort_order: 0 })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data).toEqual({ id: 'cat-groceries', type: 'expense', name: 'Groceries', active: false, sort_order: 0 })
  })
})
