import { useLayoutEffect, useRef, useState } from 'react'
import { useMoney } from '../../context/MoneyVisibilityContext'
import { useProgress } from '../../hooks/useProgress'
import { useDebtProgress } from '../../hooks/useDebtProgress'
import { LineChart } from '../../components/charts/LineChart'
import { ACCENT, RED, BLUE } from '../../lib/colors'
import type { ProgressGranularity } from '../../types'
import './Progress.css'

const VIEWS: { id: ProgressGranularity; label: string }[] = [
  { id: 'monthly', label: 'Monthly' },
  { id: 'quarterly', label: 'Quarterly' },
  { id: 'yearly', label: 'Yearly' },
]

function pctFmt(v: number): string {
  return `${v.toFixed(1)}%`
}

const MIN_LABEL_WIDTH = 46

// Renders one label per data point, but blanks out as many as needed so the
// remaining ones never crowd together — evenly thinning down to whatever
// actually fits the row's current width.
function AxisLabels({ labels, dual }: { labels: string[]; dual?: boolean }) {
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
    <div ref={ref} className={'progress-labels' + (dual ? ' progress-labels-dual' : '')}>
      {labels.map((label, i) => (
        <span className="progress-label" key={i}>
          {i % step === 0 ? label : ''}
        </span>
      ))}
    </div>
  )
}

export function Progress() {
  const { fmt } = useMoney()
  const [granularity, setGranularity] = useState<ProgressGranularity>('monthly')
  const { data, isLoading, isError } = useProgress(granularity)
  const { data: debtData, isLoading: debtLoading, isError: debtIsError } = useDebtProgress(granularity)

  return (
    <div>
      <div className="progress-header">
        <div>
          <div className="mono progress-latest">{data ? fmt(data.latest_value_idr) : '—'}</div>
        </div>
        <div className="segmented">
          {VIEWS.map((v) => (
            <button
              key={v.id}
              type="button"
              className={'segmented-btn' + (granularity === v.id ? ' segmented-btn-active' : '')}
              onClick={() => setGranularity(v.id)}
            >
              {v.label}
            </button>
          ))}
        </div>
      </div>

      <div className="card progress-chart-card">
        <div className="card-title">Asset value trend</div>
        {isLoading && <div className="empty-state">Loading…</div>}
        {isError && <div className="empty-state">Couldn't load progress data.</div>}
        {data && (
          <>
            <LineChart
              series={data.series.map((p) => ({ label: p.label, value: p.net_equity_idr }))}
              color={ACCENT}
              formatValue={fmt}
            />
            <AxisLabels labels={data.series.map((p) => p.label)} />
          </>
        )}
      </div>

      <div className="card progress-chart-card">
        <div className="card-title">Debt trend</div>
        {debtLoading && <div className="empty-state">Loading…</div>}
        {debtIsError && <div className="empty-state">Couldn't load debt progress data.</div>}
        {debtData && debtData.series.length === 0 && (
          <div className="empty-state">No debt snapshots yet.</div>
        )}
        {debtData && debtData.series.length > 0 && (
          <>
            <div className="progress-chart-legend">
              <span className="progress-chart-legend-item">
                <span className="progress-chart-legend-swatch" style={{ borderTopColor: RED }} />
                Debt value
              </span>
              <span className="progress-chart-legend-item">
                <span
                  className="progress-chart-legend-swatch progress-chart-legend-swatch-dashed"
                  style={{ borderTopColor: BLUE }}
                />
                Debt-to-equity ratio
              </span>
            </div>
            <LineChart
              series={debtData.series.map((p) => ({ label: p.label, value: p.debt_idr }))}
              color={RED}
              formatValue={fmt}
              secondarySeries={debtData.series.map((p) => ({ label: p.label, value: p.ratio_pct }))}
              secondaryColor={BLUE}
              secondaryFormatValue={pctFmt}
            />
            <AxisLabels labels={debtData.series.map((p) => p.label)} dual />
          </>
        )}
      </div>
    </div>
  )
}
