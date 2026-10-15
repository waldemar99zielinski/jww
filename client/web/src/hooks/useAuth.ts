import { useCallback } from 'react'
import { useAppDispatch, useAppSelector } from '@/store/hooks'
import { login as loginThunk, logout as logoutThunk } from '@/store/authSlice'

export function useAuth() {
  const dispatch = useAppDispatch()
  const status = useAppSelector((state) => state.auth.status)
  const user = useAppSelector((state) => state.auth.user)
  const error = useAppSelector((state) => state.auth.error)

  const login = useCallback(
    (username: string, password: string) => dispatch(loginThunk({ username, password })).unwrap(),
    [dispatch],
  )
  const logout = useCallback(() => dispatch(logoutThunk()), [dispatch])

  return { status, user, error, login, logout }
}
