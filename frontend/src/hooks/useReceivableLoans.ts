import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { api } from '../lib/api'
import type { ReceivableLoanInput } from '../types'

// A loan-terms write moves the Passive Income calendar (receivable
// payments are one of its three sources), the dashboard's passive-income
// block and any passive-income target, so all of them get invalidated —
// same fan-out useBondPurchases uses for coupons.
function invalidate(qc: ReturnType<typeof useQueryClient>) {
  qc.invalidateQueries({ queryKey: ['receivableLoans'] })
  qc.invalidateQueries({ queryKey: ['incomeCalendar'] })
  qc.invalidateQueries({ queryKey: ['dashboard'] })
  qc.invalidateQueries({ queryKey: ['targets'] })
}

export function useReceivableLoans() {
  return useQuery({ queryKey: ['receivableLoans'], queryFn: () => api.receivableLoans.list() })
}

export function useCreateReceivableLoan() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (input: ReceivableLoanInput) => api.receivableLoans.create(input),
    onSuccess: () => invalidate(qc),
  })
}

export function useUpdateReceivableLoan() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, input }: { id: string; input: ReceivableLoanInput }) =>
      api.receivableLoans.update(id, input),
    onSuccess: () => invalidate(qc),
  })
}

export function useDeleteReceivableLoan() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => api.receivableLoans.remove(id),
    onSuccess: () => invalidate(qc),
  })
}
