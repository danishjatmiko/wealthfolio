import { useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '../lib/api'
import type { FixedExpenseInput } from '../types'

function invalidateAll(qc: ReturnType<typeof useQueryClient>) {
  qc.invalidateQueries({ queryKey: ['expensePeriod'], refetchType: 'all' })
  qc.invalidateQueries({ queryKey: ['expensePeriods'] })
  qc.invalidateQueries({ queryKey: ['dashboard'] })
}

export function useCreateFixedExpense() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ periodId, input }: { periodId: string; input: FixedExpenseInput }) =>
      api.fixedExpenses.create(periodId, input),
    onSuccess: () => invalidateAll(qc),
  })
}

export function useUpdateFixedExpense() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, input }: { id: string; input: FixedExpenseInput }) =>
      api.fixedExpenses.update(id, input),
    onSuccess: () => invalidateAll(qc),
  })
}

export function useDeleteFixedExpense() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => api.fixedExpenses.remove(id),
    onSuccess: () => invalidateAll(qc),
  })
}
