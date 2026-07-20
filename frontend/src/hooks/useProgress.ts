import { useQuery } from '@tanstack/react-query'
import { api } from '../lib/api'
import type { ProgressGranularity } from '../types'

export function useProgress(granularity: ProgressGranularity) {
  return useQuery({
    queryKey: ['progress', granularity],
    queryFn: () => api.progress.get(granularity),
  })
}
