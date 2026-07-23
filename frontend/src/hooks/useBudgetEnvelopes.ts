import { useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '../lib/api'
import type { BudgetEnvelopeInput } from '../types'

function invalidateAll(qc: ReturnType<typeof useQueryClient>) {
  qc.invalidateQueries({ queryKey: ['expensePeriod'], refetchType: 'all' })
  qc.invalidateQueries({ queryKey: ['expensePeriods'] })
  qc.invalidateQueries({ queryKey: ['dashboard'] })
}

export function useCreateBudgetEnvelope() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ periodId, input }: { periodId: string; input: BudgetEnvelopeInput }) =>
      api.budgetEnvelopes.create(periodId, input),
    onSuccess: () => invalidateAll(qc),
  })
}

export function useUpdateBudgetEnvelope() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, input }: { id: string; input: BudgetEnvelopeInput }) =>
      api.budgetEnvelopes.update(id, input),
    onSuccess: () => invalidateAll(qc),
  })
}

export function useDeleteBudgetEnvelope() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => api.budgetEnvelopes.remove(id),
    onSuccess: () => invalidateAll(qc),
  })
}
