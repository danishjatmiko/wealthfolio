import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { api } from '../lib/api'
import type { PassiveIncomeInput } from '../types'

export function usePassiveIncome() {
  return useQuery({
    queryKey: ['passiveIncome'],
    queryFn: api.passiveIncome.list,
  })
}

export function useIncomeCalendar(year: number) {
  return useQuery({
    queryKey: ['incomeCalendar', year],
    queryFn: () => api.passiveIncome.calendar(year),
  })
}

function invalidate(qc: ReturnType<typeof useQueryClient>) {
  qc.invalidateQueries({ queryKey: ['passiveIncome'] })
  qc.invalidateQueries({ queryKey: ['dashboard'] })
  qc.invalidateQueries({ queryKey: ['targets'] })
  // The calendar merges these entries with the bond coupons, so any write
  // here moves a month bucket and the year's totals.
  qc.invalidateQueries({ queryKey: ['incomeCalendar'] })
}

export function useCreatePassiveIncome() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (input: PassiveIncomeInput) => api.passiveIncome.create(input),
    onSuccess: () => invalidate(qc),
  })
}

export function useUpdatePassiveIncome() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, input }: { id: string; input: PassiveIncomeInput }) =>
      api.passiveIncome.update(id, input),
    onSuccess: () => invalidate(qc),
  })
}

export function useDeletePassiveIncome() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => api.passiveIncome.remove(id),
    onSuccess: () => invalidate(qc),
  })
}
