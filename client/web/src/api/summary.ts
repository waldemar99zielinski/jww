import { apiFetch, toQueryString } from '@/api/client'
import { mockApi, mockDelay } from '@/api/mock/data'
import type { SummaryDto, SummaryPeriod } from '@/api/types'
import { isDesignMode } from '@/lib/designMode'

export interface SummaryParams {
  period_end: string
  currency_code: string
  period?: SummaryPeriod
  period_start?: string
  user_id?: string
}

export function getSummary(params: SummaryParams): Promise<SummaryDto> {
  if (isDesignMode()) return mockDelay(mockApi.getSummary(params))
  return apiFetch<SummaryDto>(`/summary${toQueryString(params)}`)
}
