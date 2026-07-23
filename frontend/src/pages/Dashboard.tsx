import { useState } from 'react'
import { useMoney } from '../context/MoneyVisibilityContext'
import { useDashboard } from '../hooks/useDashboard'
import { DonutChart, type DonutDatum } from '../components/charts/DonutChart'
import { formatShortDate } from '../lib/format'
import './Dashboard.css'

export function Dashboard() {
  const { fmt } = useMoney()
  const { data, isLoading, isError } = useDashboard()
  const [hoverAllocation, setHoverAllocation] = useState<DonutDatum | null>(null)
  const [hoverSpent, setHoverSpent] = useState<DonutDatum | null>(null)
  const [hoverCommitted, setHoverCommitted] = useState<DonutDatum | null>(null)

  if (isLoading) return <div className="empty-state">Loading dashboard…</div>
  if (isError || !data) return <div className="empty-state">Couldn't load the dashboard.</div>

  const { equity, debt, expense, allocation } = data

  return (
    <div>
      <div className="dash-section">
        <div className="dash-top-grid">
          <div className="dash-hero">
            <div className="dash-hero-label">Total Equity</div>
            <div className="dash-hero-value mono">{fmt(equity.total_idr)}</div>
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
            {allocation.length === 0 && <div className="empty-state">No holdings yet.</div>}
            {allocation.map((c) => (
              <div className="dash-cat-row" key={c.category_key}>
                <span className="dash-cat-swatch" style={{ background: c.color_oklch }} />
                <span className="dash-cat-name">{c.label}</span>
                <span className="mono dash-cat-val">{fmt(c.value_idr)}</span>
                <span className="mono dash-cat-pct">{c.percent.toFixed(2)}%</span>
              </div>
            ))}
            {equity.as_of_date && <div className="dash-updated">As of {formatShortDate(equity.as_of_date)}</div>}
          </div>
        </div>
      </div>

      <div className="dash-section">
        <div className="dash-expense-grid">
          <div className="card dash-donut-card">
            <div className="card-title">Expenses — Spent by Category</div>
            <div className="dash-donut-wrap">
              <DonutChart
                data={expense.actual_by_category.map((c) => ({ value: c.value_idr, color: c.color_oklch, label: c.label }))}
                onHover={setHoverSpent}
              />
              <div className="dash-donut-center">
                <div className="dash-donut-center-label">{hoverSpent ? hoverSpent.label : 'Spent'}</div>
                <div className="dash-donut-center-value mono">
                  {fmt(hoverSpent ? hoverSpent.value : expense.actual_total_idr)}
                </div>
              </div>
            </div>
            {expense.actual_by_category.length === 0 && (
              <div className="empty-state">No expenses logged yet.</div>
            )}
            {expense.actual_by_category.map((c) => (
              <div className="dash-cat-row" key={c.category_key}>
                <span className="dash-cat-swatch" style={{ background: c.color_oklch }} />
                <span className="dash-cat-name">{c.label}</span>
                <span className="mono dash-cat-val">{fmt(c.value_idr)}</span>
                <span className="mono dash-cat-pct">{c.percent.toFixed(2)}%</span>
              </div>
            ))}
          </div>

          <div className="card dash-donut-card">
            <div className="card-title">Expenses — Committed by Category</div>
            <div className="dash-donut-wrap">
              <DonutChart
                data={expense.committed_by_category.map((c) => ({
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
            {expense.committed_by_category.length === 0 && (
              <div className="empty-state">No budget envelopes yet.</div>
            )}
            {expense.committed_by_category.map((c) => (
              <div className="dash-cat-row" key={c.category_key}>
                <span className="dash-cat-swatch" style={{ background: c.color_oklch }} />
                <span className="dash-cat-name">{c.label}</span>
                <span className="mono dash-cat-val">{fmt(c.value_idr)}</span>
                <span className="mono dash-cat-pct">{c.percent.toFixed(2)}%</span>
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  )
}
