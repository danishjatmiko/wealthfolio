import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { api } from '../lib/api'
import type { CreateExpensePeriodInput } from '../types'

export function useExpensePeriods() {
  return useQuery({
    queryKey: ['expensePeriods'],
    queryFn: api.expensePeriods.list,
  })
}

export function useLatestExpensePeriod() {
  return useQuery({
    queryKey: ['expensePeriod', 'latest'],
    queryFn: api.expensePeriods.latest,
  })
}

export function useExpensePeriodById(id: string | undefined) {
  return useQuery({
    queryKey: ['expensePeriod', id],
    queryFn: () => api.expensePeriods.byId(id as string),
    enabled: !!id,
  })
}

export function useCreateExpensePeriod() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (input: CreateExpensePeriodInput) => api.expensePeriods.create(input),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['expensePeriod'], refetchType: 'all' })
      qc.invalidateQueries({ queryKey: ['expensePeriods'] })
      qc.invalidateQueries({ queryKey: ['dashboard'] })
    },
  })
}

export function useDeleteExpensePeriod() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => api.expensePeriods.remove(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['expensePeriods'] })
      // refetchType: 'all' so the cache entry for whichever period the UI
      // falls back to (not necessarily the currently-mounted one) is
      // refetched too, not just left stale until something re-observes it.
      qc.invalidateQueries({ queryKey: ['expensePeriod'], refetchType: 'all' })
      qc.invalidateQueries({ queryKey: ['dashboard'] })
    },
  })
}
