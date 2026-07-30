import { useEffect, useLayoutEffect, useMemo, useRef, useState } from 'react'
import { Navigate } from 'react-router-dom'
import { useMoney } from '../../context/MoneyVisibilityContext'
import { useDashboard } from '../../hooks/useDashboard'
import { LineChart } from '../../components/charts/LineChart'
import { ACCENT, BLUE } from '../../lib/colors'
import { parseNumeric } from '../../lib/format'
import { runSimulation } from '../../lib/simulate'
import './Simulation.css'

const RATE_MIN = 5
const RATE_MAX = 50
const YEARS_MIN = 3
const YEARS_MAX = 25
const INFLATION_MIN = 3
const INFLATION_MAX = 20
const MIN_LABEL_WIDTH = 46

// Same "thin down to whatever fits" approach as Progress.tsx's AxisLabels
// — up to 26 points (25 years + "Now") would otherwise crowd together.
function AxisLabels({ labels }: { labels: string[] }) {
  const ref = useRef<HTMLDivElement>(null)
  const [width, setWidth] = useState(0)

  useLayoutEffect(() => {
    const el = ref.current
    if (!el) return
    const update = () => setWidth(el.clientWidth)
    update()
    const obs = new ResizeObserver(update)
    obs.observe(el)
    return () => obs.disconnect()
  }, [])

  const maxVisible = width > 0 ? Math.max(1, Math.floor(width / MIN_LABEL_WIDTH)) : labels.length
  const step = Math.max(1, Math.ceil(labels.length / maxVisible))

  return (
    <div ref={ref} className="simulation-labels">
      {labels.map((label, i) => (
        <span className="simulation-label" key={i}>
          {i % step === 0 ? label : ''}
        </span>
      ))}
    </div>
  )
}

// A plain <input type="range"> gives no feedback at all about its current
// value while hovering/dragging — this adds a small bubble that tracks the
// thumb (approximated by percent-of-track, which is close enough for a
// slider tooltip) and fades in on hover/focus/drag via CSS, not JS state.
function RangeField({
  label,
  min,
  max,
  value,
  onChange,
  displayValue,
}: {
  label: string
  min: number
  max: number
  value: number
  onChange: (value: number) => void
  displayValue: string
}) {
  const pct = ((value - min) / (max - min)) * 100
  return (
    <label className="field">
      {label}
      <div className="simulation-range-row">
        <div className="simulation-range-track">
          <input
            type="range"
            min={min}
            max={max}
            step={1}
            value={value}
            onChange={(e) => onChange(Number(e.target.value))}
          />
          <span className="simulation-range-tooltip" style={{ left: `${pct}%` }}>
            {displayValue}
          </span>
        </div>
        <span className="mono simulation-range-value">{displayValue}</span>
      </div>
    </label>
  )
}

export function Simulation() {
  const { fmt, fmtExact } = useMoney()
  const { data: dashboard } = useDashboard()

  const [annualRatePct, setAnnualRatePct] = useState(10)
  const [years, setYears] = useState(10)
  const [monthlyCost, setMonthlyCost] = useState('')
  const [annualInflationPct, setAnnualInflationPct] = useState(5)
  const [monthlyIncome, setMonthlyIncome] = useState('')

  // Defaults to the current net worth (assets minus liabilities — exactly
  // what dashboard.equity.total_idr already is, see DashboardService),
  // fetched once the dashboard loads, but stays freely editable after
  // that — a hypothetical "what if I started with X instead" override.
  const [startingBalanceInput, setStartingBalanceInput] = useState('')
  const defaultStartingBalanceIdr = dashboard?.equity.total_idr ?? null
  useEffect(() => {
    if (defaultStartingBalanceIdr !== null) {
      setStartingBalanceInput((prev) => (prev === '' ? String(defaultStartingBalanceIdr) : prev))
    }
  }, [defaultStartingBalanceIdr])

  const startingBalanceIdr = parseNumeric(startingBalanceInput)

  const result = useMemo(
    () =>
      runSimulation({
        startingBalanceIdr,
        annualRatePct,
        years,
        monthlyCostIdr: parseNumeric(monthlyCost),
        annualInflationPct,
        monthlyIncomeIdr: parseNumeric(monthlyIncome),
      }),
    [startingBalanceIdr, annualRatePct, years, monthlyCost, annualInflationPct, monthlyIncome],
  )

  // Desktop-only, and never reachable from inside the Android app — the
  // nav link is already hidden for both cases (see AppShell.tsx), this is
  // the belt-and-suspenders guard against a direct/deep-linked visit. Kept
  // after every hook call above so hook call order never varies.
  if (window.WealthfolioNative) {
    return <Navigate to="/" replace />
  }

  return (
    <div>
      <div className="card simulation-form-card">
        <div className="card-title">Simulation parameters</div>

        <div className="simulation-field-grid">
          <label className="field simulation-starting-field">
            Starting asset (Rp)
            <input
              className="field-input mono"
              value={startingBalanceInput}
              onChange={(e) => setStartingBalanceInput(e.target.value)}
              placeholder="0"
            />
            <div className="simulation-field-hint">
              Defaults to your current net worth (total equity − liability)
              {defaultStartingBalanceIdr !== null && (
                <>
                  {' — '}
                  <button
                    type="button"
                    className="simulation-reset-link"
                    onClick={() => setStartingBalanceInput(String(defaultStartingBalanceIdr))}
                  >
                    reset to {fmtExact(defaultStartingBalanceIdr)}
                  </button>
                </>
              )}
            </div>
          </label>

          <RangeField
            label="Growth rate per year"
            min={RATE_MIN}
            max={RATE_MAX}
            value={annualRatePct}
            onChange={setAnnualRatePct}
            displayValue={`${annualRatePct}%`}
          />

          <RangeField
            label="Years to simulate"
            min={YEARS_MIN}
            max={YEARS_MAX}
            value={years}
            onChange={setYears}
            displayValue={`${years}y`}
          />

          <RangeField
            label="Inflation per year"
            min={INFLATION_MIN}
            max={INFLATION_MAX}
            value={annualInflationPct}
            onChange={setAnnualInflationPct}
            displayValue={`${annualInflationPct}%`}
          />

          <label className="field">
            Cost per month (Rp)
            <input
              className="field-input mono"
              value={monthlyCost}
              onChange={(e) => setMonthlyCost(e.target.value)}
              placeholder="15000000"
            />
          </label>

          <label className="field">
            Additional income per month (Rp)
            <input
              className="field-input mono"
              value={monthlyIncome}
              onChange={(e) => setMonthlyIncome(e.target.value)}
              placeholder="0"
            />
          </label>
        </div>
      </div>

      {result.depletedAtMonth !== null && (
        <div className="simulation-warning">
          At these settings, the balance runs out around year {Math.ceil(result.depletedAtMonth / 12)} — monthly
          cost outpaces growth + additional income.
        </div>
      )}

      <div className="simulation-result-grid">
        <div className="card simulation-result-card simulation-result-card-hero">
          <div className="simulation-result-label">Final balance (nominal)</div>
          <div className="mono simulation-result-value">{fmtExact(result.finalNominalBalanceIdr)}</div>
        </div>
        <div className="card simulation-result-card">
          <div className="simulation-result-label">Final balance (today's Rupiah)</div>
          <div className="mono simulation-result-value">{fmtExact(result.finalRealBalanceIdr)}</div>
        </div>
        <div className="card simulation-result-card">
          <div className="simulation-result-label">Total additional income added</div>
          <div className="mono simulation-result-value">{fmtExact(result.totalIncomeAddedIdr)}</div>
        </div>
        <div className="card simulation-result-card">
          <div className="simulation-result-label">Total cost paid</div>
          <div className="mono simulation-result-value">{fmtExact(result.totalCostPaidIdr)}</div>
        </div>
      </div>

      <div className="card simulation-chart-card">
        <div className="card-title">Balance over time</div>
        <div className="simulation-chart-legend">
          <span className="simulation-chart-legend-item">
            <span className="simulation-chart-legend-swatch" style={{ borderTopColor: ACCENT }} />
            Nominal balance
          </span>
          <span className="simulation-chart-legend-item">
            <span className="simulation-chart-legend-swatch simulation-chart-legend-swatch-dashed" style={{ borderTopColor: BLUE }} />
            Today's Rupiah (inflation-adjusted)
          </span>
        </div>
        <LineChart
          series={result.points.map((p) => ({ label: p.label, value: p.nominalBalanceIdr }))}
          color={ACCENT}
          formatValue={fmtExact}
          axisFormatValue={fmt}
          secondarySeries={result.points.map((p) => ({ label: p.label, value: p.realBalanceIdr }))}
          secondaryColor={BLUE}
          secondaryFormatValue={fmtExact}
          sharedScale
          tooltipExtra={{
            label: 'Cost that year',
            series: result.points.map((p) => ({ label: p.label, value: p.annualCostIdr })),
            formatValue: fmtExact,
          }}
        />
        <AxisLabels labels={result.points.map((p) => p.label)} />
      </div>
    </div>
  )
}
