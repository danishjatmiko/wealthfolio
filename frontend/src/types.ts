// Core API types, matching the Wealthfolio backend contract (v1).
// All monetary fields are integers representing THOUSANDS of IDR, unless
// noted otherwise (rate.usd_idr is full IDR per 1 USD).

export type CategoryKind = 'asset' | 'liability'

export interface Category {
  id: number
  key: string
  label: string
  color_oklch: string
  kind: CategoryKind
  price_linked: boolean
  sort_order: number
}

export interface RateEntry {
  id: string
  entry_date: string
  antam: number
  kinghalim: number
  ubs: number
  usd_idr: number
  created_at: string
}

export interface RateEntryInput {
  entry_date: string
  antam: number
  kinghalim: number
  ubs: number
  usd_idr: number
}

export interface SnapshotSummary {
  id: string
  snapshot_date: string
  is_editable: boolean
  holdings_count: number
  net_equity_idr: number
}

export type HoldingCurrency = 'IDR' | 'USD'

export interface Holding {
  id: string
  snapshot_id: string
  category_id: number
  category_key: string
  category_label: string
  name: string
  detail: string | null
  value_idr: number
  is_liability: boolean
  gram: number | null
  qty: number | null
  brand: string | null
  usd_value: number | null
  currency: HoldingCurrency | null
  created_at: string
  updated_at: string
}

export interface Snapshot {
  id: string
  snapshot_date: string
  is_editable: boolean
  holdings: Holding[]
}

export interface HoldingInput {
  category_id: number
  name: string
  gram?: number | null
  qty?: number | null
  brand?: string | null
  usd_value?: number | null
  currency?: HoldingCurrency | null
  value_idr?: number | null
  detail?: string | null
}

export type DebtDirection = 'i_owe' | 'owed_to_me'

export interface Debt {
  id: string
  name: string
  type: string
  value_idr: number
  direction: DebtDirection
}

export interface DebtInput {
  name: string
  type: string
  value_idr: number
  direction: DebtDirection
}

export interface PassiveIncomeSource {
  id: string
  category_id: number
  category_key: string
  category_label: string
  name: string
  per_year_idr: number
}

export interface PassiveIncomeInput {
  category_id: number
  name: string
  per_year_idr: number
}

export type TargetMetricType =
  | 'equity'
  | 'gold_grams'
  | 'passive_income'
  | 'debt_ratio'
  | 'custom'

export interface Target {
  id: string
  name: string
  year: number
  metric_type: TargetMetricType
  target_value: number
  unit: string
  current_value: number
  percent: number
  lower_is_better: boolean
}

export interface TargetInput {
  name: string
  year: number
  metric_type: TargetMetricType
  target_value: number
  unit: string
  manual_current_value?: number | null
}

export interface DashboardCategoryRow {
  category_key: string
  label: string
  color_oklch: string
  value_idr: number
  percent: number
}

export interface Dashboard {
  equity: {
    total_idr: number
    invested_idr: number
    incl_passive_idr: number
    mom_change_idr: number
    mom_change_pct: number
    by_category: DashboardCategoryRow[]
  }
  debt: {
    total_debt_idr: number
    total_receivable_idr: number
    ratio_pct: number
  }
  passive: {
    per_year_idr: number
    target_per_year_idr: number
    percent: number
    per_month_idr: number
    per_month_target_idr: number
  }
  allocation: DashboardCategoryRow[]
}

export type ProgressGranularity = 'monthly' | 'quarterly' | 'yearly'

export interface ProgressPoint {
  label: string
  date: string
  net_equity_idr: number
}

export interface Progress {
  granularity: ProgressGranularity
  series: ProgressPoint[]
  latest_value_idr: number
  delta_idr: number
  delta_pct: number
}
