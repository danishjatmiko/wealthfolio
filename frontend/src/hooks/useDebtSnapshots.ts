import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { api } from '../lib/api'
import type { DebtEntryInput } from '../types'

export function useDebtSnapshots() {
  return useQuery({
    queryKey: ['debtSnapshots'],
    queryFn: api.debtSnapshots.list,
  })
}

export function useLatestDebtSnapshot() {
  return useQuery({
    queryKey: ['debtSnapshot', 'latest'],
    queryFn: api.debtSnapshots.latest,
  })
}

export function useDebtSnapshotByDate(date: string | undefined) {
  return useQuery({
    queryKey: ['debtSnapshot', date],
    queryFn: () => api.debtSnapshots.byDate(date as string),
    enabled: !!date,
  })
}

function invalidateAll(qc: ReturnType<typeof useQueryClient>) {
  qc.invalidateQueries({ queryKey: ['debtSnapshot'] })
  qc.invalidateQueries({ queryKey: ['debtSnapshots'] })
  qc.invalidateQueries({ queryKey: ['dashboard'] })
  qc.invalidateQueries({ queryKey: ['targets'] })
  qc.invalidateQueries({ queryKey: ['debtProgress'] })
}

export function useCreateDebtSnapshot() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (input: { snapshot_date: string; copy_from_latest: boolean }) =>
      api.debtSnapshots.create(input),
    onSuccess: () => invalidateAll(qc),
  })
}

export function useDeleteDebtSnapshot() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => api.debtSnapshots.remove(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['debtSnapshots'] })
      // refetchType: 'all' so the cache entry for whichever date the UI
      // falls back to (not necessarily the currently-mounted one) is
      // refetched too, not just left stale until something re-observes it.
      qc.invalidateQueries({ queryKey: ['debtSnapshot'], refetchType: 'all' })
      qc.invalidateQueries({ queryKey: ['dashboard'] })
      qc.invalidateQueries({ queryKey: ['targets'] })
      qc.invalidateQueries({ queryKey: ['debtProgress'] })
    },
  })
}

export function useCreateDebtEntry() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ date, input }: { date: string; input: DebtEntryInput }) =>
      api.debtEntries.create(date, input),
    onSuccess: () => invalidateAll(qc),
  })
}

export function useUpdateDebtEntry() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, input }: { id: string; input: DebtEntryInput }) =>
      api.debtEntries.update(id, input),
    onSuccess: () => invalidateAll(qc),
  })
}

export function useDeleteDebtEntry() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => api.debtEntries.remove(id),
    onSuccess: () => invalidateAll(qc),
  })
}
