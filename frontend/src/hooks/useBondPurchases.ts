import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { api } from '../lib/api'
import type { BondPurchaseInput } from '../types'

// A ledger write moves the Passive Income calendar (coupons are half of
// what it shows), the dashboard's passive-income block and any
// passive-income target, so all of them get invalidated.
function invalidate(qc: ReturnType<typeof useQueryClient>) {
  qc.invalidateQueries({ queryKey: ['bondPurchases'] })
  qc.invalidateQueries({ queryKey: ['bondSummary'] })
  qc.invalidateQueries({ queryKey: ['incomeCalendar'] })
  qc.invalidateQueries({ queryKey: ['dashboard'] })
  qc.invalidateQueries({ queryKey: ['targets'] })
}

export function useBondPurchases() {
  return useQuery({ queryKey: ['bondPurchases'], queryFn: () => api.bondPurchases.list() })
}

export function useBondSummary() {
  return useQuery({ queryKey: ['bondSummary'], queryFn: () => api.bondPurchases.summary() })
}

export function useCreateBondPurchase() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (input: BondPurchaseInput) => api.bondPurchases.create(input),
    onSuccess: () => invalidate(qc),
  })
}

export function useUpdateBondPurchase() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, input }: { id: string; input: BondPurchaseInput }) =>
      api.bondPurchases.update(id, input),
    onSuccess: () => invalidate(qc),
  })
}

export function useDeleteBondPurchase() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => api.bondPurchases.remove(id),
    onSuccess: () => invalidate(qc),
  })
}
