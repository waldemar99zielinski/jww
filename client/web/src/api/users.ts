import { apiFetch } from '@/api/client'
import { mockApi, mockDelay } from '@/api/mock/data'
import type { UserDto } from '@/api/types'
import { isDesignMode } from '@/lib/designMode'

export function listUsers(): Promise<UserDto[]> {
  if (isDesignMode()) return mockDelay(mockApi.listUsers())
  return apiFetch<UserDto[]>('/users')
}
