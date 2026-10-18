import { renderHook, waitFor } from '@testing-library/react'
import { QueryClientProvider } from '@tanstack/react-query'
import type { ReactNode } from 'react'
import { describe, expect, it } from 'vitest'
import { useSummary } from '@/hooks/useSummary'
import { createTestQueryClient } from '@/test/renderWithProviders'
import { fixtureSummary, fixtureUser } from '@/test/fixtures'

function wrapper({ children }: { children: ReactNode }) {
  const queryClient = createTestQueryClient()
  return <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
}

describe('useSummary', () => {
  it('resolves with the fixture summary', async () => {
    const { result } = renderHook(
      () => useSummary({ user_id: fixtureUser.id, period_end: '2026-10-31', currency_code: 'EUR' }),
      { wrapper },
    )

    await waitFor(() => expect(result.current.isSuccess).toBe(true))

    expect(result.current.data).toEqual(fixtureSummary)
  })

  it('fetches without a user_id filter', async () => {
    const { result } = renderHook(() => useSummary({ period_end: '2026-10-31', currency_code: 'EUR' }), { wrapper })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))

    expect(result.current.data).toEqual(fixtureSummary)
  })
})
