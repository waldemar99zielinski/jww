import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { createCategory, listCategories, updateCategory } from '@/api/categories'
import type { TransactionType } from '@/api/types'

export function useCategories(type: TransactionType, activeOnly = true) {
  return useQuery({
    queryKey: ['categories', type, activeOnly],
    queryFn: () => listCategories(type, activeOnly),
  })
}

export function useCreateCategory() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ type, name }: { type: TransactionType; name: string }) => createCategory(type, name),
    onSuccess: (category) => {
      queryClient.invalidateQueries({ queryKey: ['categories', category.type] })
    },
  })
}

export function useUpdateCategory() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({
      id,
      ...input
    }: {
      id: string
      name: string
      active: boolean
      sort_order: number
    }) => updateCategory(id, input),
    onSuccess: (category) => {
      queryClient.invalidateQueries({ queryKey: ['categories', category.type] })
    },
  })
}
