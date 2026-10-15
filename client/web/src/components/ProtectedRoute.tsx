import { Navigate, Outlet, useLocation } from 'react-router-dom'
import { BootSplash } from '@/components/BootSplash'
import { useAppSelector } from '@/store/hooks'

export function ProtectedRoute() {
  const status = useAppSelector((state) => state.auth.status)
  const location = useLocation()

  if (status === 'loading') {
    return <BootSplash />
  }

  if (status !== 'authenticated') {
    return <Navigate to="/login" replace state={{ from: location }} />
  }

  return <Outlet />
}
