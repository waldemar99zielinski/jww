import { apiFetch, toQueryString } from '@/api/client'
import { mockApi, mockDelay } from '@/api/mock/data'
import type { TransactionCategoryDto, TransactionType } from '@/api/types'
import { isDesignMode } from '@/lib/designMode'

export function listCategories(type: TransactionType, activeOnly = true): Promise<TransactionCategoryDto[]> {
  if (isDesignMode()) return mockDelay(mockApi.listCategories(type, activeOnly))
  return apiFetch<TransactionCategoryDto[]>(
    `/transaction-categories${toQueryString({ type, active_only: activeOnly })}`,
  )
}

export function createCategory(type: TransactionType, name: string, sortOrder = 0): Promise<TransactionCategoryDto> {
  if (isDesignMode()) return mockDelay(mockApi.createCategory(type, name, sortOrder))
  return apiFetch<TransactionCategoryDto>('/transaction-categories', {
    method: 'POST',
    body: JSON.stringify({ type, name, sort_order: sortOrder }),
  })
}

export function updateCategory(
  id: string,
  input: { name: string; active: boolean; sort_order: number },
): Promise<TransactionCategoryDto> {
  if (isDesignMode()) return mockDelay(mockApi.updateCategory(id, input))
  return apiFetch<TransactionCategoryDto>(`/transaction-categories/${id}`, {
    method: 'PUT',
    body: JSON.stringify(input),
  })
}
