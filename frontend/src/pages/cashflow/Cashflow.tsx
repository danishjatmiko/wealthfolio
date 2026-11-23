import { useState } from 'react'
import { Link } from 'react-router-dom'
import { useMoney } from '../../context/MoneyVisibilityContext'
import { errorMessage } from '../../context/ToastContext'
import { useCashflow } from '../../hooks/useCashflow'
import type { CashflowMonth } from '../../types'
import { ActiveIncomeModal } from './ActiveIncomeModal'
import './Cashflow.css'

function balanceClass(value: number) {
  return value > 0 ? 'cashflow-positive' : value < 0 ? 'cashflow-negative' : ''
}

interface CashflowSegment {
  label: string
  amount: number
  className: string
}

// Positive sources stack to the right of zero; recorded losses sit to the
// left, so a negative passive-income entry never inflates the income bar.
function CashflowBar({ segments, positiveScale, negativeScale }: {
  segments: CashflowSegment[]
  positiveScale: number
  negativeScale: number
}) {
  const scale = positiveScale + negativeScale
  const positive = segments.filter((segment) => segment.amount > 0)
  const negative = segments.filter((segment) => segment.amount < 0)
  const positiveTotal = positive.reduce((sum, segment) => sum + segment.amount, 0)
  const negativeTotal = negative.reduce((sum, segment) => sum - segment.amount, 0)

  return (
    <span className="cashflow-bar-line">
      {negativeScale > 0 && <span className="cashflow-zero-line" style={{ left: `${negativeScale / scale * 100}%` }} />}
      {positiveTotal > 0 && <span className="cashflow-bar-stack" style={{ left: `${negativeScale / scale * 100}%`, width: `${positiveTotal / scale * 100}%` }}>
        {positive.map((segment) => <span key={segment.label} className={`cashflow-bar-segment ${segment.className}`}
          style={{ width: `${segment.amount / positiveTotal * 100}%` }} />)}
      </span>}
      {negativeTotal > 0 && <span className="cashflow-bar-stack cashflow-loss-stack" style={{ left: `${(negativeScale - negativeTotal) / scale * 100}%`, width: `${negativeTotal / scale * 100}%` }}>
        {negative.map((segment) => <span key={segment.label} className={`cashflow-bar-segment cashflow-loss-segment ${segment.className}`}
          style={{ width: `${-segment.amount / negativeTotal * 100}%` }} />)}
      </span>}
    </span>
  )
}

export function Cashflow() {
  const now = new Date()
  const currentYear = now.getFullYear()
  const [year, setYear] = useState(Math.max(2026, currentYear))
  const [editingMonth, setEditingMonth] = useState<CashflowMonth | null>(null)
  const [chartTooltip, setChartTooltip] = useState<{ month: number; below: boolean } | null>(null)
  const { data, isLoading, error, refetch } = useCashflow(year)
  const { fmtExact, hidden } = useMoney()
  const months = data?.months ?? []
  const isFuture = (m: CashflowMonth) => year > currentYear || (year === currentYear && m.month > now.getMonth() + 1)
  const elapsed = months.filter((m) => !isFuture(m))
  const totalIncome = elapsed.reduce((sum, m) => sum + m.total_income_idr, 0)
  const totalExpenses = elapsed.reduce((sum, m) => sum + m.total_expenses_idr, 0)
  const net = totalIncome - totalExpenses
  const missing = elapsed.filter((m) => !m.active_income || !m.has_expense_period).length
  const positiveScale = Math.max(1, ...months.flatMap((m) => [
    (m.active_income?.amount_idr ?? 0) + Math.max(0, m.passive_income_idr),
    Math.max(0, m.expenses_idr) + Math.max(0, m.big_expenses_idr),
  ]))
  const negativeScale = Math.max(0, ...months.flatMap((m) => [
    -Math.min(0, m.passive_income_idr),
    -Math.min(0, m.expenses_idr) - Math.min(0, m.big_expenses_idr),
  ]))
  const periodLabel = elapsed.length ? `${elapsed[0].label}–${elapsed[elapsed.length - 1].label} ${year}` : 'No elapsed months in this year'

  function showChartTooltip(month: number, element: HTMLElement) {
    setChartTooltip({ month, below: element.getBoundingClientRect().top < 250 })
  }

  return (
    <div className="cashflow-page">
      <div className="row-wrap cashflow-toolbar">
        <p className="cashflow-copy">See what comes in, what goes out, and what remains each month.</p>
        <div className="snapshot-pill">
          <label className="snapshot-pill-label" htmlFor="cashflow-year">Year</label>
          <select id="cashflow-year" className="snapshot-pill-select" value={year} onChange={(e) => { setChartTooltip(null); setYear(Number(e.target.value)) }}>
            {Array.from({ length: Math.max(currentYear + 2, year) - 2026 + 1 }, (_, i) => 2026 + i).map((y) => <option key={y} value={y}>{y}</option>)}
          </select>
        </div>
      </div>

      {isLoading ? <div className="empty-state">Loading cashflow…</div> : error ?
        <div className="card" role="alert"><p className="error-text">{errorMessage(error)}</p>
          <button className="btn btn-secondary" onClick={() => { void refetch() }}>Try again</button></div> : data && <>
        <div className="cashflow-summary">
          <div className="card"><div className="cashflow-kicker">Total income</div>
            <div className="mono cashflow-summary-value">{fmtExact(totalIncome)}</div>
            <div className="cashflow-caption">Active + passive income</div></div>
          <div className="card"><div className="cashflow-kicker">Total expenses</div>
            <div className="mono cashflow-summary-value">{fmtExact(totalExpenses)}</div>
            <div className="cashflow-caption">Monthly expenses + big expense</div></div>
          <div className="card cashflow-net-card"><div className="cashflow-kicker">Net cashflow</div>
            <div className={`mono cashflow-summary-value ${balanceClass(net)}`}>{fmtExact(net)}</div>
            <div className="cashflow-caption">{periodLabel} · {net > 0 ? 'Surplus' : net < 0 ? 'Deficit' : 'Balanced'}</div></div>
        </div>
        <div className="cashflow-notice">
          Tracking starts August 2026. Totals include the current month’s entries so far and exclude future months.
          {missing > 0 && <span> {missing} {missing === 1 ? 'month has' : 'months have'} missing active income or an expense period; balances are provisional.</span>}
        </div>

        <div className="card cashflow-chart-card">
          <div className="row-wrap"><h2 className="card-title">Income vs. total expenses</h2>
            <div className="cashflow-legend">
              <div className="cashflow-legend-group">
                <span><i className="cashflow-active-color" />Active income</span>
                <span><i className="cashflow-passive-color" />Passive income</span>
              </div>
              <div className="cashflow-legend-group">
                <span><i className="cashflow-monthly-color" />Monthly expenses</span>
                <span><i className="cashflow-big-color" />Big expense</span>
              </div>
            </div></div>
          {hidden ? <div className="empty-state">Amounts are hidden.</div> : months.map((m) => (
            <button type="button" key={m.month} className="cashflow-chart-row"
              onMouseEnter={(e) => showChartTooltip(m.month, e.currentTarget)}
              onMouseLeave={() => setChartTooltip(null)}
              onFocus={(e) => showChartTooltip(m.month, e.currentTarget)}
              onBlur={() => setChartTooltip(null)}
              onKeyDown={(e) => { if (e.key === 'Escape') setChartTooltip(null) }}
              onClick={() => { setChartTooltip(null); setEditingMonth(m) }}
              aria-describedby={chartTooltip?.month === m.month && !editingMonth ? `cashflow-tooltip-${m.month}` : undefined}
              aria-label={`${m.label} ${year}: active income ${fmtExact(m.active_income?.amount_idr ?? 0)}, passive income ${fmtExact(m.passive_income_idr)}, monthly expenses ${fmtExact(m.expenses_idr)}, big expense ${fmtExact(m.big_expenses_idr)}, net ${fmtExact(m.net_idr)}. Edit active income.`}>
              <span className="cashflow-chart-label">{m.label.slice(0, 3)}{isFuture(m) && <small>Upcoming</small>}</span>
              <span className="cashflow-bars" aria-hidden="true">
                <CashflowBar positiveScale={positiveScale} negativeScale={negativeScale} segments={[
                  { label: 'Active income', amount: m.active_income?.amount_idr ?? 0, className: 'cashflow-active-color' },
                  { label: 'Passive income', amount: m.passive_income_idr, className: 'cashflow-passive-color' },
                ]} />
                <CashflowBar positiveScale={positiveScale} negativeScale={negativeScale} segments={[
                  { label: 'Monthly expenses', amount: m.expenses_idr, className: 'cashflow-monthly-color' },
                  { label: 'Big expense', amount: m.big_expenses_idr, className: 'cashflow-big-color' },
                ]} />
              </span>
              <span className={`mono cashflow-chart-net ${balanceClass(m.net_idr)}`}>{fmtExact(m.net_idr)}<small>net</small></span>
              {chartTooltip?.month === m.month && !editingMonth && (
                <span id={`cashflow-tooltip-${m.month}`} role="tooltip"
                  className={`cashflow-chart-tooltip${chartTooltip.below ? ' cashflow-chart-tooltip-below' : ''}`}>
                  <span className="cashflow-tooltip-title">{m.label} {year}</span>
                  <span className="cashflow-tooltip-item"><span><i className="cashflow-active-color" />Active income</span><span className="mono">{fmtExact(m.active_income?.amount_idr ?? 0)}</span></span>
                  <span className="cashflow-tooltip-item"><span><i className="cashflow-passive-color" />Passive income</span><span className="mono">{fmtExact(m.passive_income_idr)}</span></span>
                  <span className="cashflow-tooltip-item cashflow-tooltip-total"><span>Total income</span><span className="mono">{fmtExact(m.total_income_idr)}</span></span>
                  <span className="cashflow-tooltip-item"><span><i className="cashflow-monthly-color" />Monthly expenses</span><span className="mono">{fmtExact(m.expenses_idr)}</span></span>
                  <span className="cashflow-tooltip-item"><span><i className="cashflow-big-color" />Big expense</span><span className="mono">{fmtExact(m.big_expenses_idr)}</span></span>
                  <span className="cashflow-tooltip-item cashflow-tooltip-total"><span>Total expenses</span><span className="mono">{fmtExact(m.total_expenses_idr)}</span></span>
                  <span className="cashflow-tooltip-item cashflow-tooltip-net"><span>Net cashflow</span><span className={`mono ${balanceClass(m.net_idr)}`}>{fmtExact(m.net_idr)}</span></span>
                </span>
              )}
            </button>
          ))}
          {!hidden && negativeScale > 0 && <p className="cashflow-chart-note">Hatched segments to the left of zero represent negative amounts, such as a passive-income loss.</p>}
        </div>

        <h2 className="cashflow-section-title">Monthly records</h2>
        <div className="card cashflow-table-card">
          <table className="cashflow-table">
            <thead><tr><th>Month</th><th>Active income</th><th>Passive income</th><th>Monthly expenses</th><th>Big expense</th><th>Net cashflow</th><th><span className="cashflow-action-heading">Active income</span></th></tr></thead>
            <tbody>{months.map((m) => <tr key={m.month}>
              <th scope="row">{m.label}<div className="cashflow-caption">{isFuture(m) ? 'Upcoming' : year === currentYear && m.month === now.getMonth() + 1 ? 'In progress' : year}</div></th>
              <td data-label="Active income"><span className="mono">{m.active_income ? fmtExact(m.active_income.amount_idr) : '—'}</span>
                {!m.active_income && <div className="cashflow-caption cashflow-missing">Not entered</div>}
              </td>
              <td data-label="Passive income"><span className="mono">{fmtExact(m.passive_income_idr)}</span>
                {m.projected_passive_income_idr !== 0 && <div className="cashflow-caption">Includes {fmtExact(m.projected_passive_income_idr)} projected</div>}</td>
              <td data-label="Monthly expenses"><span className="mono">{fmtExact(m.expenses_idr)}</span>
                {!m.has_expense_period && <div className="cashflow-caption cashflow-missing">No period</div>}</td>
              <td data-label="Big expense" className="mono">{fmtExact(m.big_expenses_idr)}</td>
              <td data-label="Net cashflow"><span className={`mono cashflow-net-value ${balanceClass(m.net_idr)}`}>{fmtExact(m.net_idr)}</span>
                <div className={`cashflow-caption ${balanceClass(m.net_idr)}`}>{m.net_idr > 0 ? 'Surplus' : m.net_idr < 0 ? 'Deficit' : 'Balanced'}{(!m.active_income || !m.has_expense_period || isFuture(m) || m.projected_passive_income_idr !== 0) && ' · Provisional'}</div></td>
              <td className="cashflow-actions"><button type="button" className="btn btn-secondary" onClick={() => setEditingMonth(m)}
                aria-label={`${m.active_income ? 'Edit' : 'Add'} active income for ${m.label} ${year}`}>
                {m.active_income ? 'Edit income' : '+ Add income'}</button></td>
            </tr>)}</tbody>
          </table>
        </div>
        <div className="cashflow-method card">
          <h2 className="card-title">How cashflow is calculated</h2>
          <p><strong>Net = active income + passive income − monthly expenses − big expense.</strong></p>
          <p><Link to="/expenses">Monthly Expenses</Link> uses actual spending, grouped by the period’s end month (25 Jul–24 Aug counts as August). Budget commitments are excluded.</p>
          <p><Link to="/big-expenses">Big Expense</Link> uses the expense date’s calendar month. <Link to="/passive-income">Passive Income</Link> uses the same monthly totals as its tab, including projected bond coupons and receivable installments, and any recorded losses. Projected amounts can change with source data and USD/IDR rates.</p>
          <p>Enter one active income total per month. These balances update automatically when source records change.</p>
        </div>
      </>}
      {editingMonth && <ActiveIncomeModal key={`${year}-${editingMonth.month}`} year={year} month={editingMonth} onClose={() => setEditingMonth(null)} />}
    </div>
  )
}
