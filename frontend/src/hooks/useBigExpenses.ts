import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { api } from '../lib/api'
import type { BigExpenseInput } from '../types'

function invalidate(qc: ReturnType<typeof useQueryClient>) {
  qc.invalidateQueries({ queryKey: ['bigExpenses'] })
  qc.invalidateQueries({ queryKey: ['bigExpenseSummary'] })
}

export function useBigExpenses() {
  return useQuery({ queryKey: ['bigExpenses'], queryFn: () => api.bigExpenses.list() })
}

export function useBigExpenseSummary(year: number) {
  return useQuery({
    queryKey: ['bigExpenseSummary', year],
    queryFn: () => api.bigExpenses.summary(year),
  })
}

export function useCreateBigExpense() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (input: BigExpenseInput) => api.bigExpenses.create(input),
    onSuccess: () => invalidate(qc),
  })
}

export function useUpdateBigExpense() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, input }: { id: string; input: BigExpenseInput }) =>
      api.bigExpenses.update(id, input),
    onSuccess: () => invalidate(qc),
  })
}

export function useDeleteBigExpense() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => api.bigExpenses.remove(id),
    onSuccess: () => invalidate(qc),
  })
}
