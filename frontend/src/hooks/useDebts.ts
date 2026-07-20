import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { api } from '../lib/api'
import type { DebtInput } from '../types'

export function useDebts() {
  return useQuery({
    queryKey: ['debts'],
    queryFn: api.debts.list,
  })
}

function invalidate(qc: ReturnType<typeof useQueryClient>) {
  qc.invalidateQueries({ queryKey: ['debts'] })
  qc.invalidateQueries({ queryKey: ['dashboard'] })
  qc.invalidateQueries({ queryKey: ['targets'] })
}

export function useCreateDebt() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (input: DebtInput) => api.debts.create(input),
    onSuccess: () => invalidate(qc),
  })
}

export function useUpdateDebt() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, input }: { id: string; input: DebtInput }) => api.debts.update(id, input),
    onSuccess: () => invalidate(qc),
  })
}

export function useDeleteDebt() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => api.debts.remove(id),
    onSuccess: () => invalidate(qc),
  })
}
