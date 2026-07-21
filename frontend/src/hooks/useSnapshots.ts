import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { api } from '../lib/api'

export function useSnapshots() {
  return useQuery({
    queryKey: ['snapshots'],
    queryFn: api.snapshots.list,
  })
}

export function useLatestSnapshot() {
  return useQuery({
    queryKey: ['snapshot', 'latest'],
    queryFn: api.snapshots.latest,
  })
}

export function useSnapshotByDate(date: string | undefined) {
  return useQuery({
    queryKey: ['snapshot', date],
    queryFn: () => api.snapshots.byDate(date as string),
    enabled: !!date,
  })
}

export function useCreateSnapshot() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (input: { snapshot_date: string; copy_from_latest: boolean }) =>
      api.snapshots.create(input),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['snapshots'] })
      qc.invalidateQueries({ queryKey: ['snapshot'] })
      qc.invalidateQueries({ queryKey: ['dashboard'] })
      qc.invalidateQueries({ queryKey: ['progress'] })
    },
  })
}

export function useDeleteSnapshot() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => api.snapshots.remove(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['snapshots'] })
      // refetchType: 'all' so the cache entry for whichever date the UI
      // falls back to (not necessarily the currently-mounted one) is
      // refetched too, not just left stale until something re-observes it.
      qc.invalidateQueries({ queryKey: ['snapshot'], refetchType: 'all' })
      qc.invalidateQueries({ queryKey: ['dashboard'] })
      qc.invalidateQueries({ queryKey: ['progress'] })
    },
  })
}
