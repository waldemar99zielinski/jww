import { renderHook, waitFor } from '@testing-library/react'
import { QueryClientProvider } from '@tanstack/react-query'
import type { ReactNode } from 'react'
import { HttpResponse, http } from 'msw'
import { describe, expect, it } from 'vitest'
import { useDefaultCurrency } from '@/hooks/useDefaultCurrency'
import { createTestQueryClient } from '@/test/renderWithProviders'
import { server } from '@/test/server'

function wrapper({ children }: { children: ReactNode }) {
  const queryClient = createTestQueryClient()
  return <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
}

describe('useDefaultCurrency', () => {
  it('picks the is_base currency regardless of its position in the list', async () => {
    server.use(
      http.get('*/api/v1/currencies', () =>
        HttpResponse.json({
          data: [
            { code: 'USD', name: 'US Dollar', is_base: false },
            { code: 'EUR', name: 'Euro', is_base: true },
          ],
        }),
      ),
    )

    const { result } = renderHook(() => useDefaultCurrency(), { wrapper })

    await waitFor(() => expect(result.current.isLoading).toBe(false))
    expect(result.current.currency?.code).toBe('EUR')
  })

  it('falls back to the first currency when none is marked is_base', async () => {
    server.use(
      http.get('*/api/v1/currencies', () =>
        HttpResponse.json({
          data: [
            { code: 'USD', name: 'US Dollar', is_base: false },
            { code: 'EUR', name: 'Euro', is_base: false },
          ],
        }),
      ),
    )

    const { result } = renderHook(() => useDefaultCurrency(), { wrapper })

    await waitFor(() => expect(result.current.isLoading).toBe(false))
    expect(result.current.currency?.code).toBe('USD')
  })

  it('resolves to null when there are no currencies at all', async () => {
    server.use(http.get('*/api/v1/currencies', () => HttpResponse.json({ data: [] })))

    const { result } = renderHook(() => useDefaultCurrency(), { wrapper })

    await waitFor(() => expect(result.current.isLoading).toBe(false))
    expect(result.current.currency).toBeNull()
    expect(result.current.currencies).toEqual([])
  })
})
