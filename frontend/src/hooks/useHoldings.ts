import { useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '../lib/api'
import type { HoldingInput } from '../types'

function invalidateAll(qc: ReturnType<typeof useQueryClient>) {
  qc.invalidateQueries({ queryKey: ['snapshot'] })
  qc.invalidateQueries({ queryKey: ['snapshots'] })
  qc.invalidateQueries({ queryKey: ['dashboard'] })
  qc.invalidateQueries({ queryKey: ['progress'] })
  qc.invalidateQueries({ queryKey: ['targets'] })
}

export function useCreateHolding() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ date, input }: { date: string; input: HoldingInput }) =>
      api.holdings.create(date, input),
    onSuccess: () => invalidateAll(qc),
  })
}

export function useUpdateHolding() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, input }: { id: string; input: HoldingInput }) =>
      api.holdings.update(id, input),
    onSuccess: () => invalidateAll(qc),
  })
}

export function useDeleteHolding() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => api.holdings.remove(id),
    onSuccess: () => invalidateAll(qc),
  })
}
