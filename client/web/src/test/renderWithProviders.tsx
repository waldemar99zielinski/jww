import { configureStore } from '@reduxjs/toolkit'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render } from '@testing-library/react'
import type { ReactElement } from 'react'
import { Provider as ReduxProvider } from 'react-redux'
import { MemoryRouter } from 'react-router-dom'
import { authReducer, type AuthState } from '@/store/authSlice'

/** Fresh per test: `retry: false`/`gcTime: 0` so a mocked error/loading state surfaces immediately
 *  instead of retrying or getting cached across tests. */
export function createTestQueryClient(): QueryClient {
  return new QueryClient({
    defaultOptions: {
      queries: { retry: false, gcTime: 0 },
      mutations: { retry: false },
    },
  })
}

/** Pass `authState` to preload a specific session (e.g. an authenticated user) instead of the
 *  default unauthenticated/loading initial state. */
export function createTestStore(authState?: AuthState) {
  return configureStore({
    reducer: { auth: authReducer },
    ...(authState ? { preloadedState: { auth: authState } } : {}),
  })
}

type TestStore = ReturnType<typeof createTestStore>

interface RenderWithProvidersOptions {
  route?: string
  queryClient?: QueryClient
  store?: TestStore
}

/** Mirrors `src/providers.tsx`, swapping `BrowserRouter` for `MemoryRouter` and constructing a
 *  fresh store/query client per call instead of reusing the app's singletons. */
export function renderWithProviders(
  ui: ReactElement,
  { route = '/', queryClient = createTestQueryClient(), store = createTestStore() }: RenderWithProvidersOptions = {},
) {
  return {
    ...render(
      <ReduxProvider store={store}>
        <QueryClientProvider client={queryClient}>
          <MemoryRouter initialEntries={[route]}>{ui}</MemoryRouter>
        </QueryClientProvider>
      </ReduxProvider>,
    ),
    queryClient,
    store,
  }
}
