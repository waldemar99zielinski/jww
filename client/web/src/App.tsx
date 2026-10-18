import { lazy, Suspense, use, type ReactNode } from 'react'
import { Navigate, Route, Routes } from 'react-router-dom'
import { AppShell } from '@/components/AppShell'
import { BootSplash } from '@/components/BootSplash'
import { OfflineGate } from '@/components/OfflineGate'
import { ProtectedRoute } from '@/components/ProtectedRoute'
import { store } from '@/store'
import { restoreSession } from '@/store/authSlice'

const LoginPage = lazy(() => import('@/pages/login/LoginPage'))
const HomePage = lazy(() => import('@/pages/home'))
const TransactionsPage = lazy(() => import('@/pages/transactions'))
const TransactionDetailPage = lazy(() => import('@/pages/transactions/detail'))
const AddTransactionPage = lazy(() => import('@/pages/add'))
const BudgetsPage = lazy(() => import('@/pages/budgets'))
const SettingsPage = lazy(() => import('@/pages/settings'))
const SettingsCategoriesPage = lazy(() => import('@/pages/settings/categories'))
const SettingsCurrenciesPage = lazy(() => import('@/pages/settings/currencies'))

/** Floor so BootSplash's <Logo animated /> unfold is never cut off before it's visible. */
const MIN_BOOT_SPLASH_MS = 1000

/**
 * Module-level singleton: the session check, the two likely-first-screen chunks, and the
 * display floor, bundled into one promise created exactly once. `use()` in BootGate suspends
 * on it until it settles — this is the sole boot-readiness mechanism, no separate "ready"
 * state or pre-Suspense gate.
 */
let bootPromise: Promise<void> | undefined

function getBootPromise(): Promise<void> {
  bootPromise ??= Promise.all([
    store.dispatch(restoreSession()),
    import('@/pages/login/LoginPage'),
    import('@/pages/home'),
    new Promise((resolve) => setTimeout(resolve, MIN_BOOT_SPLASH_MS)),
  ]).then(() => undefined)
  return bootPromise
}

/** Suspends (showing the nearest Suspense fallback) until the app is ready, then renders children. */
function BootGate({ children }: { children: ReactNode }) {
  use(getBootPromise())
  return children
}

export function App() {
  return (
    <OfflineGate>
      <Suspense fallback={<BootSplash />}>
        <BootGate>
          <Routes>
            <Route path="/login" element={<LoginPage />} />

            <Route element={<ProtectedRoute />}>
              <Route element={<AppShell />}>
                <Route path="/" element={<HomePage />} />
                <Route path="/transactions" element={<TransactionsPage />} />
                <Route path="/transactions/:id" element={<TransactionDetailPage />} />
                <Route path="/add" element={<AddTransactionPage />} />
                <Route path="/budgets" element={<BudgetsPage />} />
                <Route path="/settings" element={<SettingsPage />} />
                <Route path="/settings/categories" element={<SettingsCategoriesPage />} />
                <Route path="/settings/currencies" element={<SettingsCurrenciesPage />} />
              </Route>
            </Route>

            <Route path="*" element={<Navigate to="/" replace />} />
          </Routes>
        </BootGate>
      </Suspense>
    </OfflineGate>
  )
}
