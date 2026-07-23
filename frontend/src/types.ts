// Core API types, matching the Wealthfolio backend contract (v1).
// All monetary fields are integers representing THOUSANDS of IDR, unless
// noted otherwise (rate.usd_idr is full IDR per 1 USD).

export interface User {
  id: string
  email: string
  display_name: string
  avatar_url: string
}

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

export interface DebtSnapshotSummary {
  id: string
  snapshot_date: string
  is_editable: boolean
  entries_count: number
  i_owe_idr: number
  owed_to_me_idr: number
}

export interface DebtEntry {
  id: string
  debt_snapshot_id: string
  name: string
  type: string
  value_idr: number
  direction: DebtDirection
  created_at: string
  updated_at: string
}

export interface DebtSnapshot {
  id: string
  snapshot_date: string
  is_editable: boolean
  entries: DebtEntry[]
}

export interface DebtEntryInput {
  name: string
  type: string
  value_idr: number
  direction: DebtDirection
}

export type DebtProgressGranularity = 'monthly' | 'quarterly' | 'yearly'

export interface DebtProgressPoint {
  label: string
  date: string
  debt_idr: number
  owed_to_me_idr: number
  ratio_pct: number
}

export interface DebtProgress {
  granularity: DebtProgressGranularity
  series: DebtProgressPoint[]
  latest_debt_idr: number
  latest_ratio_pct: number
  delta_idr: number
  delta_pct: number
}

export interface ExpenseCategory {
  id: string
  name: string
  created_at: string
}

export interface ExpenseCategoryInput {
  name: string
}

export interface ExpensePeriodSummary {
  id: string
  start_date: string
  end_date: string
  label: string
  actual_total_idr: number
  committed_total_idr: number
}

export interface FixedExpense {
  id: string
  period_id: string
  envelope_id: string
  name: string
  amount_idr: number
  created_at: string
  updated_at: string
}

export interface BudgetEnvelope {
  id: string
  period_id: string
  category_id: string
  category_name: string
  name: string
  committed_amount_idr: number
  created_at: string
  updated_at: string
}

export interface BudgetEnvelopeDetail {
  id: string
  category_id: string
  category_name: string
  name: string
  committed_amount_idr: number
  actual_total_idr: number
  fixed_expenses: FixedExpense[]
}

export interface ExpensePeriodDetail {
  id: string
  start_date: string
  end_date: string
  label: string
  envelopes: BudgetEnvelopeDetail[]
  actual_total_idr: number
  committed_total_idr: number
}

export interface BudgetEnvelopeInput {
  category_id: string
  name: string
  committed_amount_idr: number
}

export interface FixedExpenseInput {
  name: string
  amount_idr: number
  envelope_id: string
}

export interface CreateExpensePeriodInput {
  year: number
  month: number
  copy_envelopes: boolean
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
    as_of_date: string | null
  }
  debt: {
    total_debt_idr: number
    total_receivable_idr: number
    ratio_pct: number
    updated_at: string | null
  }
  passive: {
    per_year_idr: number
    target_per_year_idr: number
    percent: number
    per_month_idr: number
    per_month_target_idr: number
    updated_at: string | null
  }
  expense: {
    period_label: string
    actual_total_idr: number
    committed_total_idr: number
    actual_by_category: DashboardCategoryRow[]
    committed_by_category: DashboardCategoryRow[]
    updated_at: string | null
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
