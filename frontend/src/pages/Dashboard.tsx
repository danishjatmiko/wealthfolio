import { useMoney } from '../context/MoneyVisibilityContext'
import { useDashboard } from '../hooks/useDashboard'
import { DonutChart } from '../components/charts/DonutChart'
import './Dashboard.css'

export function Dashboard() {
  const { fmt } = useMoney()
  const { data, isLoading, isError } = useDashboard()

  if (isLoading) return <div className="empty-state">Loading dashboard…</div>
  if (isError || !data) return <div className="empty-state">Couldn't load the dashboard.</div>

  const { equity, debt, passive, allocation } = data
  const passivePct = Math.round(passive.percent)

  return (
    <div>
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
          </div>
        </div>

        <div className="card dash-mini-card">
          <div className="dash-mini-label">Debt</div>
          <div className="dash-mini-value mono">{fmt(debt.total_debt_idr)}</div>
          <div className="dash-mini-note">ratio {debt.ratio_pct.toFixed(2)}%</div>
          <div className="dash-mini-divider">
            Owed to me <span className="mono dash-mini-divider-val">{fmt(debt.total_receivable_idr)}</span>
          </div>
        </div>

        <div className="card dash-mini-card">
          <div className="dash-mini-label">Passive / year</div>
          <div className="dash-mini-value mono">{fmt(passive.per_year_idr)}</div>
          <div className="dash-mini-note">{passivePct}% of {fmt(passive.target_per_year_idr)} target</div>
          <div className="progress-track dash-passive-track">
            <div
              className="progress-fill"
              style={{ width: `${Math.min(100, passivePct)}%`, background: 'var(--blue)' }}
            />
          </div>
        </div>
      </div>

      <div className="dash-bottom-grid">
        <div className="card dash-donut-card">
          <div className="card-title">Allocation</div>
          <div className="dash-donut-wrap">
            <DonutChart data={allocation.map((a) => ({ value: a.value_idr, color: a.color_oklch }))} />
            <div className="dash-donut-center">
              <div className="dash-donut-center-label">Invested</div>
              <div className="dash-donut-center-value mono">{fmt(equity.invested_idr)}</div>
            </div>
          </div>
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
        </div>
      </div>
    </div>
  )
}
