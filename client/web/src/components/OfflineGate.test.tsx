import { act, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider, onlineManager, useQuery } from '@tanstack/react-query'
import { afterEach, describe, expect, it } from 'vitest'
import { OfflineGate } from '@/components/OfflineGate'

function setOnLine(value: boolean) {
  Object.defineProperty(window.navigator, 'onLine', { configurable: true, get: () => value })
}

function goOffline() {
  act(() => {
    setOnLine(false)
    window.dispatchEvent(new Event('offline'))
  })
}

function goOnline() {
  act(() => {
    setOnLine(true)
    window.dispatchEvent(new Event('online'))
  })
}

afterEach(() => {
  setOnLine(true)
  // The default TanStack online manager is a module singleton; leave it online for the next test.
  onlineManager.setOnline(true)
})

describe('OfflineGate', () => {
  it('renders children interactively while online, with no offline screen and no inert wrapper', () => {
    const { container } = render(
      <OfflineGate>
        <button type="button">Save</button>
      </OfflineGate>,
    )

    expect(screen.getByRole('button', { name: 'Save' })).toBeInTheDocument()
    expect(screen.queryByRole('alertdialog')).not.toBeInTheDocument()
    expect(container.querySelector('[inert]')).toBeNull()
  })

  it('covers the app, marks it inert, and focuses Retry when offline', async () => {
    const { container } = render(
      <OfflineGate>
        <button type="button">Save</button>
      </OfflineGate>,
    )
    screen.getByRole('button', { name: 'Save' }).focus()

    goOffline()

    expect(await screen.findByRole('alertdialog')).toBeInTheDocument()
    const wrapper = container.querySelector('[inert]')
    expect(wrapper).not.toBeNull()
    expect(wrapper).toContainElement(screen.getByRole('button', { name: 'Save' }))
    expect(screen.getByRole('button', { name: 'Retry' })).toHaveFocus()
  })

  it('restores focus and announces "Back online" on reconnect', () => {
    render(
      <OfflineGate>
        <button type="button">Save</button>
      </OfflineGate>,
    )
    const save = screen.getByRole('button', { name: 'Save' })
    save.focus()

    goOffline()
    goOnline()

    expect(screen.queryByRole('alertdialog')).not.toBeInTheDocument()
    expect(save).toHaveFocus()
    expect(screen.getByRole('status')).toHaveTextContent('Back online')
  })

  it('keeps the block and updates the message when Retry finds the connection still down', async () => {
    render(
      <OfflineGate>
        <button type="button">Save</button>
      </OfflineGate>,
    )

    goOffline()
    await userEvent.click(await screen.findByRole('button', { name: 'Retry' }))

    expect(screen.getByRole('alertdialog')).toBeInTheDocument()
    expect(screen.getByRole('status')).toHaveTextContent('Still offline — check your connection.')
  })

  it('pauses a query that starts offline and resumes it on reconnect instead of erroring', async () => {
    function Probe({ enabled }: { enabled: boolean }) {
      const { data, fetchStatus, isError } = useQuery({
        queryKey: ['probe'],
        queryFn: () => Promise.resolve('resolved'),
        enabled,
      })
      return <span>{isError ? 'error' : (data ?? fetchStatus)}</span>
    }

    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })
    function tree(enabled: boolean) {
      return (
        <QueryClientProvider client={queryClient}>
          <OfflineGate>
            <Probe enabled={enabled} />
          </OfflineGate>
        </QueryClientProvider>
      )
    }

    const { rerender } = render(tree(false))
    expect(screen.getByText('idle')).toBeInTheDocument()

    // The provider has mounted, so the default online manager is now listening to connectivity.
    goOffline()
    rerender(tree(true))

    expect(await screen.findByText('paused')).toBeInTheDocument()
    expect(screen.queryByText('error')).not.toBeInTheDocument()

    goOnline()
    await waitFor(() => expect(screen.getByText('resolved')).toBeInTheDocument())
  })
})
