import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { api } from '../lib/api'
import type { TargetInput } from '../types'

export function useTargets() {
  return useQuery({
    queryKey: ['targets'],
    queryFn: api.targets.list,
  })
}

function invalidate(qc: ReturnType<typeof useQueryClient>) {
  qc.invalidateQueries({ queryKey: ['targets'] })
}

export function useCreateTarget() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (input: TargetInput) => api.targets.create(input),
    onSuccess: () => invalidate(qc),
  })
}

export function useUpdateTarget() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, input }: { id: string; input: TargetInput }) => api.targets.update(id, input),
    onSuccess: () => invalidate(qc),
  })
}

export function useDeleteTarget() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => api.targets.remove(id),
    onSuccess: () => invalidate(qc),
  })
}
