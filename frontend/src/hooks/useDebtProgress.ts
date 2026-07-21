import { useQuery } from '@tanstack/react-query'
import { api } from '../lib/api'
import type { DebtProgressGranularity } from '../types'

export function useDebtProgress(granularity: DebtProgressGranularity) {
  return useQuery({
    queryKey: ['debtProgress', granularity],
    queryFn: () => api.debtProgress.get(granularity),
  })
}
