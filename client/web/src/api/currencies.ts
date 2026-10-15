import { apiFetch } from '@/api/client'
import { mockApi, mockDelay } from '@/api/mock/data'
import type { CurrencyDto } from '@/api/types'
import { isDesignMode } from '@/lib/designMode'

export function listCurrencies(): Promise<CurrencyDto[]> {
  if (isDesignMode()) return mockDelay(mockApi.listCurrencies())
  return apiFetch<CurrencyDto[]>('/currencies')
}

export function createCurrency(code: string, name: string, isBase: boolean): Promise<CurrencyDto> {
  if (isDesignMode()) return mockDelay(mockApi.createCurrency(code, name, isBase))
  return apiFetch<CurrencyDto>('/currencies', {
    method: 'POST',
    body: JSON.stringify({ code, name, is_base: isBase }),
  })
}

export function updateCurrency(code: string, input: { name: string; is_base: boolean }): Promise<CurrencyDto> {
  if (isDesignMode()) return mockDelay(mockApi.updateCurrency(code, input))
  return apiFetch<CurrencyDto>(`/currencies/${code}`, {
    method: 'PUT',
    body: JSON.stringify(input),
  })
}
