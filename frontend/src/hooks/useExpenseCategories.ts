import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { api } from '../lib/api'
import type { ExpenseCategoryInput } from '../types'

export function useExpenseCategories() {
  return useQuery({
    queryKey: ['expenseCategories'],
    queryFn: api.expenseCategories.list,
  })
}

export function useCreateExpenseCategory() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (input: ExpenseCategoryInput) => api.expenseCategories.create(input),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['expenseCategories'] })
    },
  })
}
