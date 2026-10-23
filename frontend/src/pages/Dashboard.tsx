import { useState } from 'react'
import { useMoney } from '../context/MoneyVisibilityContext'
import { useDashboard, useDashboardAssets, useDashboardExpense } from '../hooks/useDashboard'
import { useExpensePeriods } from '../hooks/useExpensePeriods'
import { useSnapshots } from '../hooks/useSnapshots'
import { DonutChart, type DonutDatum } from '../components/charts/DonutChart'
import { CategoryRowList } from '../components/CategoryRowList'
import { formatShortDate, parseLocalDate } from '../lib/format'
import { BudgetVsActualCard, PeriodHistoryCard } from './DashboardBudget'
import './Dashboard.css'

const snapshotDateFormat = new Intl.DateTimeFormat('en-GB', { day: 'numeric', month: 'long', year: 'numeric' })

function snapshotLabel(date: string): string {
  return snapshotDateFormat.format(parseLocalDate(date))
}

export function Dashboard() {
  const { fmt, fmtExact } = useMoney()
  const { data, isLoading, isError } = useDashboard()
  const [hoverAllocation, setHoverAllocation] = useState<DonutDatum | null>(null)
  const [hoverSpent, setHoverSpent] = useState<DonutDatum | null>(null)
  const [hoverCommitted, setHoverCommitted] = useState<DonutDatum | null>(null)
  const { data: snapshots = [] } = useSnapshots()
  const [selectedSnapshotDate, setSelectedSnapshotDate] = useState<string | undefined>(undefined)
  const { data: periods = [] } = useExpensePeriods()
  const [selectedPeriodId, setSelectedPeriodId] = useState<string | undefined>(undefined)

  // The latest period's expense block already arrives with the dashboard;
  // only a different pick needs its own request. A pick that no longer
  // exists (deleted on the Expenses page) falls back to the latest.
  const latestPeriodId = data?.expense.period_id ?? undefined
  const activePeriodId =
    selectedPeriodId && periods.some((p) => p.id === selectedPeriodId) ? selectedPeriodId : latestPeriodId
  const otherPeriodId = activePeriodId !== latestPeriodId ? activePeriodId : undefined
  const { data: otherExpense } = useDashboardExpense(otherPeriodId)

  // Same arrangement for asset snapshots: the latest rides along with the
  // dashboard, any other pick is fetched on its own.
  const latestSnapshotDate = data?.equity.as_of_date ?? undefined
  const activeSnapshotDate =
    selectedSnapshotDate && snapshots.some((s) => s.snapshot_date === selectedSnapshotDate)
      ? selectedSnapshotDate
      : latestSnapshotDate
  const otherSnapshotDate = activeSnapshotDate !== latestSnapshotDate ? activeSnapshotDate : undefined
  const { data: otherAssets } = useDashboardAssets(otherSnapshotDate)

  if (isLoading) return <div className="empty-state">Loading dashboard…</div>
  if (isError || !data) return <div className="empty-state">Couldn't load the dashboard.</div>

  const { debt } = data
  const assets = (otherSnapshotDate && otherAssets) || data
  const { equity, allocation } = assets
  const expense = (otherPeriodId && otherExpense) || data.expense

  function selectPeriod(id: string) {
    setSelectedPeriodId(id)
    setHoverSpent(null)
    setHoverCommitted(null)
  }

  return (
    <div>
      <div className="dash-section">
        {snapshots.length > 0 && (
          <div className="row-wrap dash-expense-head">
            <div className="dash-section-title">Assets</div>
            <div className="snapshot-pill">
              <span className="snapshot-pill-label">Snapshot</span>
              <select
                className="snapshot-pill-select"
                value={activeSnapshotDate ?? ''}
                onChange={(e) => {
                  setSelectedSnapshotDate(e.target.value)
                  setHoverAllocation(null)
                }}
              >
                {snapshots.map((s) => (
                  <option key={s.id} value={s.snapshot_date}>
                    {snapshotLabel(s.snapshot_date)}
                  </option>
                ))}
              </select>
            </div>
            {otherSnapshotDate && <span className="dash-expense-range">Debt card still shows your latest debt figures</span>}
          </div>
        )}
        <div className="dash-top-grid">
          <div className="dash-hero">
            <div className="dash-hero-label">Total Equity</div>
            <div className="dash-hero-value mono">{fmtExact(equity.total_idr)}</div>
            <div className="dash-hero-footer">
              <div className="hovwrap">
                <div className="dash-hero-footer-label">Invested ⓘ</div>
                <div className="dash-hero-footer-value mono">{fmt(equity.invested_idr)}</div>
                <div className="hovtip">
                  <div className="hovtip-title">Allocation breakdown</div>
                  {equity.by_category.map((c) => (
                    <div className="hovtip-row" key={c.category_key}>
                      <span className="hovtip-swatch" style={{ background: c.color_oklch }} />
                      <span className="hovtip-name">{c.label}</span>
                      <span className="hovtip-val mono">{fmt(c.value_idr)}</span>
                      <span className="hovtip-pct mono">{c.percent.toFixed(2)}%</span>
                    </div>
                  ))}
                </div>
              </div>
              <div>
                <div className="dash-hero-footer-label">Liability</div>
                <div className="dash-hero-footer-value mono">{fmt(equity.total_idr - equity.invested_idr)}</div>
              </div>
            </div>
            {equity.as_of_date && (
              <div className="dash-updated dash-updated-hero">As of {formatShortDate(equity.as_of_date)}</div>
            )}
          </div>

          <div className="card dash-mini-card">
            <div className="dash-mini-label">Debt</div>
            <div className="dash-mini-value mono">{fmt(debt.total_debt_idr)}</div>
            <div className="dash-mini-note">ratio {debt.ratio_pct.toFixed(2)}%</div>
            <div className="dash-mini-divider">
              Owed to me <span className="mono dash-mini-divider-val">{fmt(debt.total_receivable_idr)}</span>
            </div>
            {debt.updated_at && <div className="dash-updated">Updated {formatShortDate(debt.updated_at)}</div>}
          </div>
        </div>
      </div>

      <div className="dash-section">
        <div className="dash-bottom-grid">
          <div className="card dash-donut-card">
            <div className="card-title">Allocation</div>
            <div className="dash-donut-wrap">
              <DonutChart
                data={allocation.map((a) => ({ value: a.value_idr, color: a.color_oklch, label: a.label }))}
                onHover={setHoverAllocation}
              />
              <div className="dash-donut-center">
                <div className="dash-donut-center-label">{hoverAllocation ? hoverAllocation.label : 'Invested'}</div>
                <div className="dash-donut-center-value mono">
                  {fmt(hoverAllocation ? hoverAllocation.value : equity.invested_idr)}
                </div>
              </div>
            </div>
            {equity.as_of_date && <div className="dash-updated">As of {formatShortDate(equity.as_of_date)}</div>}
          </div>

          <div className="card">
            <div className="card-title">By category</div>
            <CategoryRowList rows={allocation} fmt={fmt} emptyMessage="No holdings yet." />
            {equity.as_of_date && <div className="dash-updated">As of {formatShortDate(equity.as_of_date)}</div>}
          </div>
        </div>
      </div>

      <div className="dash-section">
        <div className="row-wrap dash-expense-head">
          <div className="dash-section-title">Expenses</div>
          {periods.length > 0 && (
            <div className="snapshot-pill">
              <span className="snapshot-pill-label">Period</span>
              <select
                className="snapshot-pill-select"
                value={activePeriodId ?? ''}
                onChange={(e) => selectPeriod(e.target.value)}
              >
                {periods.map((p) => (
                  <option key={p.id} value={p.id}>
                    {p.label}
                  </option>
                ))}
              </select>
            </div>
          )}
          {expense.start_date && expense.end_date && (
            <span className="dash-expense-range">
              {formatShortDate(expense.start_date)} – {formatShortDate(expense.end_date)}
            </span>
          )}
        </div>
        <div className="dash-expense-grid">
          <div className="card dash-donut-card">
            <div className="card-title">Expenses — Spent by Envelope</div>
            <div className="dash-donut-wrap">
              <DonutChart
                data={expense.actual_by_envelope.map((c) => ({ value: c.value_idr, color: c.color_oklch, label: c.label }))}
                onHover={setHoverSpent}
              />
              <div className="dash-donut-center">
                <div className="dash-donut-center-label">{hoverSpent ? hoverSpent.label : 'Spent'}</div>
                <div className="dash-donut-center-value mono">
                  {fmt(hoverSpent ? hoverSpent.value : expense.actual_total_idr)}
                </div>
              </div>
            </div>
            <CategoryRowList rows={expense.actual_by_envelope} fmt={fmt} emptyMessage="No expenses logged yet." />
          </div>

          <div className="card dash-donut-card">
            <div className="card-title">Expenses — Committed by Envelope</div>
            <div className="dash-donut-wrap">
              <DonutChart
                data={expense.committed_by_envelope.map((c) => ({
                  value: c.value_idr,
                  color: c.color_oklch,
                  label: c.label,
                }))}
                onHover={setHoverCommitted}
              />
              <div className="dash-donut-center">
                <div className="dash-donut-center-label">{hoverCommitted ? hoverCommitted.label : 'Committed'}</div>
                <div className="dash-donut-center-value mono">
                  {fmt(hoverCommitted ? hoverCommitted.value : expense.committed_total_idr)}
                </div>
              </div>
            </div>
            <CategoryRowList rows={expense.committed_by_envelope} fmt={fmt} emptyMessage="No budget envelopes yet." />
          </div>
        </div>
      </div>

      <div className="dash-section">
        <div className="dash-budget-grid">
          <BudgetVsActualCard expense={expense} fmt={fmt} />
          <PeriodHistoryCard
            periods={periods}
            selectedId={activePeriodId}
            onSelect={selectPeriod}
            fmt={fmt}
          />
        </div>
      </div>
    </div>
  )
}
