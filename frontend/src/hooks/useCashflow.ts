import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { api } from '../lib/api'
import type { ActiveIncomeInput } from '../types'

export function useCashflow(year: number) {
  return useQuery({ queryKey: ['cashflow', year], queryFn: () => api.cashflow.summary(year) })
}

export function useSaveActiveIncome() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ month, input }: { month: string; input: ActiveIncomeInput }) =>
      api.cashflow.saveActiveIncome(month, input),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['cashflow'] }) },
  })
}

export function useDeleteActiveIncome() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (month: string) => api.cashflow.deleteActiveIncome(month),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['cashflow'] }) },
  })
}
