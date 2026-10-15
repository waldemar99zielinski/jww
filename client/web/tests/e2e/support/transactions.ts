import type { APIRequestContext } from '@playwright/test'
import { authHeader } from './apiAuth'
import { API_BASE_URL } from './testUser'

export interface CreateTransactionInput {
  type: 'income' | 'expense'
  amount: number
  currency_code: string
  category_id: string
  user_id: string
  source: 'manual' | 'recurring'
  name: string
  description?: string
}

export interface TransactionDto extends CreateTransactionInput {
  id: number
  category_name: string
  created_at: string
}

export async function createTransaction(
  request: APIRequestContext,
  token: string,
  input: CreateTransactionInput,
): Promise<TransactionDto> {
  const response = await request.post(`${API_BASE_URL}/api/v1/transactions`, {
    headers: authHeader(token),
    data: input,
  })
  if (!response.ok()) {
    throw new Error(`Failed to create an e2e fixture transaction: ${response.status()} ${await response.text()}`)
  }
  const { data } = (await response.json()) as { data: TransactionDto }
  return data
}

interface MeDto {
  id: string
}

export async function getCurrentUserId(request: APIRequestContext, token: string): Promise<string> {
  const response = await request.get(`${API_BASE_URL}/api/v1/auth/me`, { headers: authHeader(token) })
  if (!response.ok()) {
    throw new Error(`Failed to resolve the current e2e user id: ${response.status()}`)
  }
  const { data } = (await response.json()) as { data: MeDto }
  return data.id
}
