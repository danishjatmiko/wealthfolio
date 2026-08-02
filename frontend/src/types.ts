// Core API types, matching the Etherna backend contract (v1).
// All monetary fields are integers representing full/raw IDR (whole
// Rupiah) — no rounding, no scaling factor.

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
  /** Which notification source (e.g. "gopay") auto-created this expense —
   * null for anything entered manually. Set once at creation, never
   * changed by a later edit. */
  source: string | null
  notes: string | null
  created_at: string
  updated_at: string
}

export interface BudgetEnvelope {
  id: string
  period_id: string
  name: string
  committed_amount_idr: number
  created_at: string
  updated_at: string
}

export interface BudgetEnvelopeDetail {
  id: string
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
  name: string
  committed_amount_idr: number
}

export interface FixedExpenseInput {
  name: string
  amount_idr: number
  envelope_id: string
  notes?: string | null
}

export interface CreateExpensePeriodInput {
  year: number
  month: number
  copy_envelopes: boolean
}

/** One passive-income payment actually received on one date — a dividend, a
 *  realized capital gain, a fund redemption. `amount_idr` may be negative: a
 *  realized loss is part of what the year paid out. `income_type` (Dividen,
 *  Capital, IPO, …) is orthogonal to the category, which is the asset class
 *  the income came from. */
export interface PassiveIncomeEntry {
  id: string
  category_id: number
  category_key: string
  category_label: string
  name: string
  amount_idr: number
  received_date: string
  income_type: string
  note: string
  created_at: string
  updated_at: string
}

export interface PassiveIncomeInput {
  category_id: number
  name: string
  amount_idr: number
  received_date: string
  income_type: string
  note: string
}

/** One USD bond purchase. Fields after usd_idr_at_purchase are derived by
 *  the backend, never sent on writes. interest_rate is a percent (5.69 =
 *  5.69%); price_usd is the aggregate lot price, not per unit. */
export interface BondPurchase {
  id: string
  bond_name: string
  platform: string
  interest_rate: number
  buy_date: string
  maturity_date: string
  quantity: number
  face_value_usd: number
  price_usd: number
  accrued_interest_usd: number
  usd_idr_at_purchase: number

  /** Invested amount — the clean price, excluding accrued interest. */
  total_usd: number
  total_idr: number
  /** What actually left the account: clean price + accrued interest. */
  settlement_usd: number
  settlement_idr: number
  coupon_per_cycle_usd: number
  coupon_per_year_usd: number
  price_pct: number
  /** Annualised yield to maturity, percent. */
  ytm_pct: number
  coupon_months: number[]
  coupon_day: number
  is_matured: boolean

  created_at: string
  updated_at: string
}

export interface BondPurchaseInput {
  bond_name: string
  platform: string
  interest_rate: number
  buy_date: string
  maturity_date: string
  quantity: number
  face_value_usd: number
  price_usd: number
  accrued_interest_usd: number
  usd_idr_at_purchase: number
}

/** Every purchase filed under one bond name, rolled into one position.
 *  has_conflicts flags lots that disagree on rate or maturity — i.e. two
 *  different bonds accidentally sharing a name. */
export interface BondNameSummary {
  bond_name: string
  platforms: string[]
  interest_rate: number
  maturity_date: string
  quantity: number
  total_usd: number
  total_idr: number
  settlement_usd: number
  settlement_idr: number
  average_usd_idr: number
  face_total_usd: number
  coupon_per_cycle_usd: number
  coupon_per_year_usd: number
  ytm_pct: number
  coupon_months: number[]
  coupon_day: number
  is_matured: boolean
  has_conflicts: boolean
  purchases: BondPurchase[]
}

export interface BondLedgerSummary {
  bonds: BondNameSummary[]
  total_usd: number
  total_idr: number
  settlement_usd: number
  settlement_idr: number
  /** Blended rate actually paid — historical, unlike the displayed IDR
   *  figures which convert at the latest rate. */
  average_usd_idr: number
  coupon_per_year_usd: number
  coupon_per_year_idr: number
  ytm_pct: number
  latest_usd_idr: number
}

/** Where one income entry came from. A `coupon` is derived from the bond
 *  ledger — it has a dollar amount and an empty `id`, so it's editable only
 *  on the Bonds page. A `manual` entry is a row in the passive-income
 *  ledger: IDR-only, and its `id` is what the Passive Income page edits. */
export type IncomeEntryKind = 'coupon' | 'manual'

/** One payment landing on one date, from either source. `source` is the
 *  bond's platform for coupons and the income type for manual entries —
 *  distinct from the category fields, which are the asset class it came
 *  from (coupons file under Bonds USD). */
export interface IncomeEntry {
  kind: IncomeEntryKind
  id: string
  name: string
  source: string
  category_key: string
  category_label: string
  color_oklch: string
  pay_date: string
  amount_usd: number
  amount_idr: number
  note: string
}

/** One category's contribution to a single month — what the month's bar is
 *  stacked out of. Emitted in the same order for every month, so a category
 *  keeps its position along the bar and months stay comparable. */
export interface IncomeMonthSlice {
  category_key: string
  label: string
  color_oklch: string
  amount_idr: number
}

/** One calendar month's income. The two sources stay separately visible
 *  alongside the combined `amount_idr` — coupons are forward-looking and
 *  dollar-denominated, manual entries are realized and in Rupiah. */
export interface IncomeMonth {
  month: number
  label: string
  coupon_usd: number
  coupon_idr: number
  manual_idr: number
  amount_idr: number
  percent: number
  slices: IncomeMonthSlice[]
  entries: IncomeEntry[]
}

/** One asset class's share of the year's income — which part of the
 *  portfolio actually paid out. Bond coupons file under `bonds_usd`, so the
 *  ring covers every source and not just the hand-logged ones. Ordered
 *  biggest-first by the backend. */
export interface IncomeCategory {
  category_key: string
  label: string
  color_oklch: string
  amount_idr: number
  percent: number
  count: number
}

/** Every source of passive income for one reference year, bucketed by month
 *  and by category. `months` is always exactly 12 — a month with nothing in
 *  it still needs a row. */
export interface IncomeCalendar {
  reference_year: number
  months: IncomeMonth[]
  categories: IncomeCategory[]
  coupon_total_usd: number
  coupon_total_idr: number
  manual_total_idr: number
  total_idr: number
  latest_usd_idr: number
}

/** One manually-logged big expense — a bond, a laptop, a hotel stay. Kept
 *  separate from the recurring monthly envelope budget in FixedExpense. */
export interface BigExpense {
  id: string
  name: string
  amount_idr: number
  expense_date: string
  category: string
  created_at: string
  updated_at: string
}

export interface BigExpenseInput {
  name: string
  amount_idr: number
  expense_date: string
  category: string
}

export interface BigExpenseMonth {
  month: number
  label: string
  amount_idr: number
  percent: number
  entries: BigExpense[]
}

export interface BigExpenseCategory {
  category: string
  color_oklch: string
  amount_idr: number
  percent: number
  count: number
}

/** Big Expense ledger summary for one reference year. `months` is always
 *  exactly 12 — a month with no entries still needs a row. */
export interface BigExpenseSummary {
  reference_year: number
  months: BigExpenseMonth[]
  categories: BigExpenseCategory[]
  total_idr: number
  entries_count: number
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
    actual_by_envelope: DashboardCategoryRow[]
    committed_by_envelope: DashboardCategoryRow[]
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
