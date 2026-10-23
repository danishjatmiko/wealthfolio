import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { api } from '../lib/api'

export function useDashboard() {
  return useQuery({
    queryKey: ['dashboard'],
    queryFn: api.dashboard.get,
  })
}

/** The Home page's equity hero and allocation as of one asset snapshot.
 *  Disabled until a snapshot other than the latest is picked, for the same
 *  reasons as useDashboardExpense below. */
export function useDashboardAssets(snapshotDate: string | undefined) {
  return useQuery({
    queryKey: ['dashboard', 'assets', snapshotDate],
    queryFn: () => api.dashboard.assets(snapshotDate as string),
    enabled: !!snapshotDate,
    placeholderData: keepPreviousData,
  })
}

/** The Home page's expense block for one period. Disabled until a period
 *  other than the latest is picked — the latest already arrives inside
 *  useDashboard. Keyed under 'dashboard' so every expense write that
 *  invalidates the dashboard refreshes this too. Keeps showing the previous
 *  period while the next loads, so switching doesn't blank the charts. */
export function useDashboardExpense(periodId: string | undefined) {
  return useQuery({
    queryKey: ['dashboard', 'expense', periodId],
    queryFn: () => api.dashboard.expense(periodId as string),
    enabled: !!periodId,
    placeholderData: keepPreviousData,
  })
}
