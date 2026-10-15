import type { APIRequestContext } from '@playwright/test'
import { API_BASE_URL, type TestUser } from './testUser'

export async function getAccessToken(request: APIRequestContext, user: TestUser): Promise<string> {
  const response = await request.post(`${API_BASE_URL}/api/v1/auth/login`, {
    data: { username: user.username, password: user.password },
  })
  if (!response.ok()) {
    throw new Error(`Failed to log in "${user.username}" for e2e API setup: ${response.status()}`)
  }
  const { data } = (await response.json()) as { data: { access_token: string } }
  return data.access_token
}

export function authHeader(token: string): Record<string, string> {
  return { Authorization: `Bearer ${token}` }
}
