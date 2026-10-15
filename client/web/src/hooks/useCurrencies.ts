import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { createCurrency, listCurrencies, updateCurrency } from '@/api/currencies'

export function useCurrencies() {
  return useQuery({ queryKey: ['currencies'], queryFn: listCurrencies })
}

export function useCreateCurrency() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ code, name, isBase }: { code: string; name: string; isBase: boolean }) =>
      createCurrency(code, name, isBase),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['currencies'] })
    },
  })
}

export function useUpdateCurrency() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ code, name, isBase }: { code: string; name: string; isBase: boolean }) =>
      updateCurrency(code, { name, is_base: isBase }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['currencies'] })
    },
  })
}
