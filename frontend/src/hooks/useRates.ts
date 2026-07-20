import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { api } from '../lib/api'
import type { RateEntryInput } from '../types'

export function useRates() {
  return useQuery({
    queryKey: ['rates'],
    queryFn: api.rates.list,
  })
}

export function useLatestRate() {
  return useQuery({
    queryKey: ['rates', 'latest'],
    queryFn: api.rates.latest,
  })
}

export function useCreateRate() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (input: RateEntryInput) => api.rates.create(input),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['rates'] })
      qc.invalidateQueries({ queryKey: ['dashboard'] })
    },
  })
}
