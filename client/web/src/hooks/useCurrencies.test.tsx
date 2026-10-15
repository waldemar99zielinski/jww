import { renderHook, waitFor } from '@testing-library/react'
import { QueryClientProvider } from '@tanstack/react-query'
import type { ReactNode } from 'react'
import { describe, expect, it } from 'vitest'
import { useCreateCurrency, useCurrencies, useUpdateCurrency } from '@/hooks/useCurrencies'
import { createTestQueryClient } from '@/test/renderWithProviders'

function wrapper({ children }: { children: ReactNode }) {
  const queryClient = createTestQueryClient()
  return <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
}

describe('useCurrencies', () => {
  it('resolves with the fixture currencies', async () => {
    const { result } = renderHook(() => useCurrencies(), { wrapper })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))

    expect(result.current.data?.map((c) => c.code)).toEqual(['EUR', 'USD'])
  })
})

describe('useCreateCurrency', () => {
  it('resolves with the created currency and invalidates the currencies query', async () => {
    const queryClient = createTestQueryClient()
    function localWrapper({ children }: { children: ReactNode }) {
      return <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
    }

    const { result } = renderHook(() => useCreateCurrency(), { wrapper: localWrapper })

    result.current.mutate({ code: 'PLN', name: 'Polish złoty', isBase: false })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data).toEqual({ code: 'PLN', name: 'Polish złoty', is_base: false })
  })
})

describe('useUpdateCurrency', () => {
  it('resolves with the updated currency', async () => {
    const { result } = renderHook(() => useUpdateCurrency(), { wrapper })

    result.current.mutate({ code: 'USD', name: 'US Dollar', isBase: true })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data).toEqual({ code: 'USD', name: 'US Dollar', is_base: true })
  })
})
