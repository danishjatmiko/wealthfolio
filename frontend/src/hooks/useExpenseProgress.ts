import { useQuery } from '@tanstack/react-query'
import { api } from '../lib/api'
import type { ProgressGranularity } from '../types'

/** Spend and budget over time, one point per month/quarter/year with
 *  expense periods. Expense writes invalidate ['progress', 'expense']. */
export function useExpenseProgress(granularity: ProgressGranularity) {
  return useQuery({
    queryKey: ['progress', 'expense', granularity],
    queryFn: () => api.expenseProgress.get(granularity),
  })
}
