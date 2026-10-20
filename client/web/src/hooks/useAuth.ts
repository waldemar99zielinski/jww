import { useCallback } from 'react'
import { useAppDispatch, useAppSelector } from '@/store/hooks'
import { login as loginThunk, loginWithPasskey as loginWithPasskeyThunk, logout as logoutThunk } from '@/store/authSlice'

export function useAuth() {
  const dispatch = useAppDispatch()
  const status = useAppSelector((state) => state.auth.status)
  const user = useAppSelector((state) => state.auth.user)
  const error = useAppSelector((state) => state.auth.error)

  const login = useCallback(
    (username: string, password: string) => dispatch(loginThunk({ username, password })).unwrap(),
    [dispatch],
  )
  const loginWithPasskey = useCallback(() => dispatch(loginWithPasskeyThunk()).unwrap(), [dispatch])
  const logout = useCallback(() => dispatch(logoutThunk()), [dispatch])

  return { status, user, error, login, loginWithPasskey, logout }
}
