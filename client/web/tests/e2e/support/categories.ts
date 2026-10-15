import type { APIRequestContext } from '@playwright/test'
import { authHeader } from './apiAuth'
import { API_BASE_URL } from './testUser'

export interface CategoryDto {
  id: string
  type: 'income' | 'expense'
  name: string
  active: boolean
  sort_order: number
}

/** Categories are global/bootstrapped (server/migrations/00006_seed_transaction_categories.sql),
 *  so every test user already sees the same set — look one up by its known seeded name. */
export async function findCategory(
  request: APIRequestContext,
  token: string,
  type: 'income' | 'expense',
  name: string,
): Promise<CategoryDto> {
  const response = await request.get(`${API_BASE_URL}/api/v1/transaction-categories?type=${type}`, {
    headers: authHeader(token),
  })
  if (!response.ok()) {
    throw new Error(`Failed to list ${type} categories for e2e setup: ${response.status()}`)
  }
  const { data } = (await response.json()) as { data: CategoryDto[] }
  const category = data.find((c) => c.name === name)
  if (!category) {
    throw new Error(`Expected bootstrapped category "${name}" (${type}) not found — was the seed migration run?`)
  }
  return category
}
