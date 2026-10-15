import { apiFetch } from '@/api/client'
import type { UserDto } from '@/api/types'

export interface TokenResponse {
  access_token: string
  token_type: string
}

export function login(username: string, password: string): Promise<TokenResponse> {
  return apiFetch<TokenResponse>('/auth/login', {
    method: 'POST',
    body: JSON.stringify({ username, password }),
  })
}

export function getMe(): Promise<UserDto> {
  return apiFetch<UserDto>('/auth/me')
}
