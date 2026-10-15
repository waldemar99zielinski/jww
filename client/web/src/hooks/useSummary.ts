import { useQuery } from '@tanstack/react-query'
import { getSummary, type SummaryParams } from '@/api/summary'

export function useSummary(params: SummaryParams) {
  return useQuery({
    queryKey: ['summary', params],
    queryFn: () => getSummary(params),
    enabled: params.user_id !== undefined,
  })
}
