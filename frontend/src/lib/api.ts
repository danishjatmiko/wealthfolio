import type {
  Category,
  Dashboard,
  DebtEntry,
  DebtEntryInput,
  DebtProgress,
  DebtProgressGranularity,
  DebtSnapshot,
  DebtSnapshotSummary,
  Holding,
  HoldingInput,
  PassiveIncomeInput,
  PassiveIncomeSource,
  Progress,
  ProgressGranularity,
  RateEntry,
  RateEntryInput,
  Snapshot,
  SnapshotSummary,
  Target,
  TargetInput,
} from '../types'

const BASE = '/api/v1'

export class ApiError extends Error {
  status: number
  constructor(status: number, message: string) {
    super(message)
    this.status = status
    this.name = 'ApiError'
  }
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE}${path}`, {
    headers: { 'Content-Type': 'application/json' },
    ...init,
  })
  if (!res.ok) {
    let message = `Request failed (${res.status})`
    try {
      const body = await res.json()
      if (body && typeof body === 'object' && 'message' in body && typeof body.message === 'string') {
        message = body.message
      } else if (body && typeof body === 'object' && 'error' in body && typeof body.error === 'string') {
        message = body.error
      }
    } catch {
      // response wasn't JSON; keep the default message
    }
    throw new ApiError(res.status, message)
  }
  if (res.status === 204) return undefined as T
  return (await res.json()) as T
}

const get = <T>(path: string) => request<T>(path)
const post = <T>(path: string, body: unknown) =>
  request<T>(path, { method: 'POST', body: JSON.stringify(body) })
const put = <T>(path: string, body: unknown) =>
  request<T>(path, { method: 'PUT', body: JSON.stringify(body) })
const del = <T>(path: string) => request<T>(path, { method: 'DELETE' })

export const api = {
  categories: {
    list: () => get<Category[]>('/categories'),
  },
  rates: {
    list: () => get<RateEntry[]>('/rates'),
    latest: () => get<RateEntry>('/rates/latest'),
    create: (input: RateEntryInput) => post<RateEntry>('/rates', input),
  },
  snapshots: {
    list: () => get<SnapshotSummary[]>('/snapshots'),
    latest: () => get<Snapshot>('/snapshots/latest'),
    byDate: (date: string) => get<Snapshot>(`/snapshots/${date}`),
    create: (input: { snapshot_date: string; copy_from_latest: boolean }) =>
      post<Snapshot>('/snapshots', input),
    remove: (id: string) => del<void>(`/snapshots/${id}`),
  },
  holdings: {
    create: (date: string, input: HoldingInput) =>
      post<Holding>(`/snapshots/${date}/holdings`, input),
    update: (id: string, input: HoldingInput) => put<Holding>(`/holdings/${id}`, input),
    remove: (id: string) => del<void>(`/holdings/${id}`),
  },
  debtSnapshots: {
    list: () => get<DebtSnapshotSummary[]>('/debt-snapshots'),
    latest: () => get<DebtSnapshot>('/debt-snapshots/latest'),
    byDate: (date: string) => get<DebtSnapshot>(`/debt-snapshots/${date}`),
    create: (input: { snapshot_date: string; copy_from_latest: boolean }) =>
      post<DebtSnapshot>('/debt-snapshots', input),
    remove: (id: string) => del<void>(`/debt-snapshots/${id}`),
  },
  debtEntries: {
    create: (date: string, input: DebtEntryInput) =>
      post<DebtEntry>(`/debt-snapshots/${date}/entries`, input),
    update: (id: string, input: DebtEntryInput) => put<DebtEntry>(`/debt-entries/${id}`, input),
    remove: (id: string) => del<void>(`/debt-entries/${id}`),
  },
  passiveIncome: {
    list: () => get<PassiveIncomeSource[]>('/passive-income'),
    create: (input: PassiveIncomeInput) => post<PassiveIncomeSource>('/passive-income', input),
    update: (id: string, input: PassiveIncomeInput) =>
      put<PassiveIncomeSource>(`/passive-income/${id}`, input),
    remove: (id: string) => del<void>(`/passive-income/${id}`),
  },
  targets: {
    list: () => get<Target[]>('/targets'),
    create: (input: TargetInput) => post<Target>('/targets', input),
    update: (id: string, input: TargetInput) => put<Target>(`/targets/${id}`, input),
    remove: (id: string) => del<void>(`/targets/${id}`),
  },
  dashboard: {
    get: () => get<Dashboard>('/dashboard'),
  },
  progress: {
    get: (granularity: ProgressGranularity) =>
      get<Progress>(`/progress?granularity=${granularity}`),
  },
  debtProgress: {
    get: (granularity: DebtProgressGranularity) =>
      get<DebtProgress>(`/debt-progress?granularity=${granularity}`),
  },
}
