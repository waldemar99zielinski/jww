import type { APIRequestContext } from '@playwright/test'
import { authHeader } from './apiAuth'
import { API_BASE_URL } from './testUser'

export interface CurrencyDto {
  code: string
  name: string
  is_base: boolean
}

/**
 * Currencies are NOT seeded by any migration (unlike categories) — a brand-new database has
 * none, and Add Transaction needs one to exist. Reuse whatever currency is already there (likely,
 * on a dev DB that's been used before) or create one, so specs don't depend on run order/DB state.
 */
export async function ensureCurrency(request: APIRequestContext, token: string): Promise<CurrencyDto> {
  const listResponse = await request.get(`${API_BASE_URL}/api/v1/currencies`, { headers: authHeader(token) })
  if (!listResponse.ok()) {
    throw new Error(`Failed to list currencies for e2e setup: ${listResponse.status()}`)
  }
  const { data: existing } = (await listResponse.json()) as { data: CurrencyDto[] }
  if (existing.length > 0) return existing[0]

  const createResponse = await request.post(`${API_BASE_URL}/api/v1/currencies`, {
    headers: authHeader(token),
    data: { code: 'EUR', name: 'Euro', is_base: true },
  })
  if (!createResponse.ok()) {
    throw new Error(`Failed to create a fallback currency for e2e setup: ${createResponse.status()}`)
  }
  const { data: created } = (await createResponse.json()) as { data: CurrencyDto }
  return created
}
